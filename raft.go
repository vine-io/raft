// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package raft

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/rafthttp"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	stats "go.etcd.io/etcd/server/v3/etcdserver/api/v2stats"
	"go.etcd.io/etcd/server/v3/wal"
	"go.etcd.io/etcd/server/v3/wal/walpb"
	"go.uber.org/zap"
)

type RaftNode interface {
	Propose(ctx context.Context, op *Operation) error
	ProposeConfChange(ctx context.Context, cc *raftpb.ConfChange) error
	Stop(ctx context.Context) error
}

type commit struct {
	data       *CommitData
	applyDoneC chan<- struct{}
}

func NewCommitChannel(n int) chan *commit {
	if n == 0 {
		return make(chan *commit)
	}
	return make(chan *commit, n)
}

// A key-value stream backed by raft
type raftNode struct {
	Config

	ctx    context.Context
	cancel context.CancelFunc

	commitC chan<- *commit // entries committed to log (k,v)
	errorC  chan<- error   // errors from raft session

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	node        raft.Node
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL

	snapshotter *snap.Snapshotter

	transport *rafthttp.Transport
	stopc     chan struct{} // signals proposal channel closed
	httpstopc chan struct{} // signals http server to shutdown
	httpdonec chan struct{} // signals http server shutdown complete

	logger *zap.Logger
}

// NewRaftNode initiates a raft instance and returns a committed log entry
// channel and error channel. Proposals for log updates are sent over the
// provided the proposal channel. All log entries are replayed over the
// commit channel, followed by a nil message (to indicate the channel is
// current), then new log entries. To shutdown, close proposeC and read errorC.
func NewRaftNode(logger *zap.Logger, snapshotter *snap.Snapshotter, config Config, commitC chan *commit, errorC chan error) RaftNode {

	ctx, cancel := context.WithCancel(context.Background())

	rc := &raftNode{
		Config:      config,
		ctx:         ctx,
		cancel:      cancel,
		commitC:     commitC,
		errorC:      errorC,
		stopc:       make(chan struct{}),
		httpstopc:   make(chan struct{}),
		httpdonec:   make(chan struct{}),
		snapshotter: snapshotter,

		logger: logger,
	}

	// rest of structure populated after WAL replay
	go rc.startRaft()
	return rc
}

func (rc *raftNode) Propose(ctx context.Context, op *Operation) error {
	data, err := op.Marshal()
	if err != nil {
		return err
	}

	return rc.node.Propose(ctx, data)
}

func (rc *raftNode) ProposeConfChange(ctx context.Context, cc *raftpb.ConfChange) error {
	return rc.node.ProposeConfChange(ctx, cc)
}

func (rc *raftNode) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case rc.stopc <- struct{}{}:
	}

	return nil
}

func (rc *raftNode) saveSnap(snap raftpb.Snapshot) error {
	walSnap := walpb.Snapshot{
		Index:     snap.Metadata.Index,
		Term:      snap.Metadata.Term,
		ConfState: &snap.Metadata.ConfState,
	}
	// save the snapshot file before writing the snapshot to the wal.
	// This makes it possible for the snapshot file to become orphaned, but prevents
	// a WAL snapshot entry from having no corresponding snapshot file.
	if err := rc.snapshotter.SaveSnap(snap); err != nil {
		return err
	}
	if err := rc.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}
	return rc.wal.ReleaseLockTo(snap.Metadata.Index)
}

func (rc *raftNode) entriesToApply(ents []raftpb.Entry) (nents []raftpb.Entry) {
	if len(ents) == 0 {
		return ents
	}
	firstIdx := ents[0].Index
	if firstIdx > rc.appliedIndex+1 {
		rc.logger.Sugar().Infof("first index of committed entry[%d] should <= progress.appliedIndex[%d]+1", firstIdx, rc.appliedIndex)
	}
	if rc.appliedIndex-firstIdx+1 < uint64(len(ents)) {
		nents = ents[rc.appliedIndex-firstIdx+1:]
	}
	return nents
}

// publishEntries writes committed log entries to commit channel and returns
// whether all entries could be published.
func (rc *raftNode) publishEntries(ents []raftpb.Entry) (<-chan struct{}, bool) {
	if len(ents) == 0 {
		return nil, true
	}

	data := &CommitData{}
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}

			op := &Operation{}
			err := op.Unmarshal(ents[i].Data)
			if err == nil {
				data.Term, data.Index = ents[i].Term, ents[i].Index
				data.Ops = append(data.Ops, op)
			}
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(ents[i].Data)
			rc.confState = *rc.node.ApplyConfChange(cc)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					rc.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}
			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(rc.Id) {
					rc.logger.Sugar().Infof("I've been removed from the cluster! Shutting down.")
					return nil, false
				}
				rc.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}
	}

	var applyDoneC chan struct{}

	if len(data.Ops) > 0 {
		applyDoneC = make(chan struct{}, 1)
		select {
		case rc.commitC <- &commit{data, applyDoneC}:
		case <-rc.stopc:
			return nil, false
		}
	}

	// after commit, update appliedIndex
	rc.appliedIndex = ents[len(ents)-1].Index
	rc.logger.Sugar().Debugf("apply index %d", rc.appliedIndex)

	return applyDoneC, true
}

func (rc *raftNode) loadSnapshot() *raftpb.Snapshot {
	if wal.Exist(rc.Dir) {
		walSnaps, err := wal.ValidSnapshotEntries(rc.logger, rc.Dir)
		if err != nil {
			rc.logger.Sugar().Fatalf("failed to listing snapshots (%v)", err)
		}
		snapshot, err := rc.snapshotter.LoadNewestAvailable(walSnaps)
		if err != nil && err != snap.ErrNoSnapshot {
			rc.logger.Sugar().Fatalf("failed to load snapshot (%v)", err)
		}
		return snapshot
	}
	return &raftpb.Snapshot{}
}

// openWAL returns a WAL ready for reading.
func (rc *raftNode) openWAL(snapshot *raftpb.Snapshot) (*wal.WAL, error) {
	if !wal.Exist(rc.Dir) {
		if err := os.Mkdir(rc.Dir, 0750); err != nil {
			return nil, fmt.Errorf("cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(rc.logger, rc.Dir, nil)
		if err != nil {
			return nil, fmt.Errorf("create wal error (%v)", err)
		}
		w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	rc.logger.Sugar().Infof("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index)
	w, err := wal.Open(rc.logger, rc.Dir, walsnap)
	if err != nil {
		return nil, fmt.Errorf("error loading wal (%v)", err)
	}

	return w, nil
}

// replayWAL replays WAL entries into the raft instance.
func (rc *raftNode) replayWAL() *wal.WAL {
	rc.logger.Sugar().Infof("replaying WAL of member %d", rc.Id)
	snapshot := rc.loadSnapshot()
	w, err := rc.openWAL(snapshot)
	if err != nil {
		rc.logger.Sugar().Fatalf("failed to open wal: %v", err)
	}
	_, st, ents, err := w.ReadAll()
	if err != nil {
		rc.logger.Sugar().Fatalf("failed to read WAL (%v)", err)
	}
	rc.raftStorage = raft.NewMemoryStorage()
	if snapshot != nil {
		_ = rc.raftStorage.ApplySnapshot(*snapshot)
	}
	_ = rc.raftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log
	_ = rc.raftStorage.Append(ents)

	return w
}

func (rc *raftNode) writeError(err error) {
	rc.stopHTTP()
	close(rc.commitC)
	rc.errorC <- err
	close(rc.errorC)
	rc.node.Stop()
}

func (rc *raftNode) startRaft() {
	var err error
	//if !fileutil.Exist(rc.snapdir) {
	//	if err = os.Mkdir(rc.snapdir, 0750); err != nil {
	//		rc.logger.Sugar().Fatalf("cannot create dir for snapshot (%v)", err)
	//	}
	//}
	//rc.snapshotter = snap.New(rc.logger, rc.snapdir)

	oldwal := wal.Exist(rc.Dir)
	rc.wal = rc.replayWAL()

	rpeers := make([]raft.Peer, len(rc.Peers))
	for i, peer := range rc.Peers {
		name, uri, _ := peerSplit(peer)
		rpeers[i] = raft.Peer{ID: uint64(hash(name)), Context: []byte(uri)}
	}
	c := &raft.Config{
		ID:                        uint64(rc.Id),
		ElectionTick:              int(rc.ElectionTick),
		HeartbeatTick:             int(rc.HeartbeatTick),
		Storage:                   rc.raftStorage,
		MaxSizePerMsg:             rc.MaxSizePerMsg,
		MaxInflightMsgs:           int(rc.MaxInflightMsgs),
		MaxUncommittedEntriesSize: rc.MaxUncommittedEntriesSize,
	}

	if oldwal || rc.Join {
		rc.logger.Sugar().Infof("raft node (%v) restart", rc.Id)
		rc.node = raft.RestartNode(c)
	} else {
		rc.logger.Sugar().Infof("raft node (%v) start", rc.Id)
		rc.node = raft.StartNode(c, rpeers)
	}

	rc.transport = &rafthttp.Transport{
		Logger:      rc.logger,
		ID:          types.ID(rc.Id),
		ClusterID:   types.ID(rc.ClusterID),
		Raft:        rc,
		ServerStats: stats.NewServerStats(rc.Name, strconv.Itoa(rc.Id)),
		LeaderStats: stats.NewLeaderStats(rc.logger, strconv.Itoa(rc.Id)),
		ErrorC:      make(chan error),
	}

	if err = rc.transport.Start(); err != nil {
		rc.logger.Sugar().Fatalf("failed to start transport (%v)", err)
	}
	for _, peer := range rc.Peers {
		name, uri, _ := peerSplit(peer)
		rc.transport.AddPeer(types.ID(hash(name)), []string{uri})
	}

	go rc.serveRaft()
	go rc.serveChannels()
}

// stop closes http, closes all channels, and stops raft.
func (rc *raftNode) stop() {
	rc.stopHTTP()
	close(rc.commitC)
	close(rc.errorC)
	rc.node.Stop()
}

func (rc *raftNode) stopHTTP() {
	rc.transport.Stop()
	close(rc.httpstopc)
	<-rc.httpdonec
}

func (rc *raftNode) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	rc.logger.Sugar().Debugf("publishing snapshot at index %d", rc.snapshotIndex)
	defer rc.logger.Sugar().Debugf("finished publishing snapshot at index %d", rc.snapshotIndex)

	if snapshotToSave.Metadata.Index <= rc.appliedIndex {
		rc.logger.Sugar().Fatalf("snapshot index [%d] should > progress.appliedIndex [%d]", snapshotToSave.Metadata.Index, rc.appliedIndex)
	}
	rc.commitC <- nil // trigger kvstore to load snapshot

	rc.confState = snapshotToSave.Metadata.ConfState
	rc.snapshotIndex = snapshotToSave.Metadata.Index
	rc.appliedIndex = snapshotToSave.Metadata.Index
}

func (rc *raftNode) maybeTriggerSnapshot(applyDoneC <-chan struct{}) error {
	if rc.appliedIndex-rc.snapshotIndex <= rc.SnapCount {
		return nil
	}

	// wait until all committed entries are applied (or server is closed)
	if applyDoneC != nil {
		select {
		case <-applyDoneC:
		case <-rc.stopc:
			return nil
		}
	}

	rc.logger.Sugar().Debugf("start snapshot [applied index: %d | last snapshot index: %d]", rc.appliedIndex, rc.snapshotIndex)
	data, err := rc.GetSnapshot()
	if err != nil {
		rc.logger.Sugar().Fatal(err)
	}
	snap, err := rc.raftStorage.CreateSnapshot(rc.appliedIndex, &rc.confState, data)
	if err != nil {
		return err
	}
	if err = rc.saveSnap(snap); err != nil {
		return err
	}

	compactIndex := uint64(1)
	if rc.appliedIndex > rc.SnapshotCatchUpEntriesN {
		compactIndex = rc.appliedIndex - rc.SnapshotCatchUpEntriesN
	}
	if err = rc.raftStorage.Compact(compactIndex); err != nil {
		return err
	}

	rc.logger.Sugar().Debugf("compacted log at index %d", compactIndex)
	rc.snapshotIndex = rc.appliedIndex

	return nil
}

func (rc *raftNode) serveChannels() {
	snap, err := rc.raftStorage.Snapshot()
	if err != nil {
		rc.logger.Sugar().Fatalf("get snapshot (%v)", err)
	}
	rc.confState = snap.Metadata.ConfState
	rc.snapshotIndex = snap.Metadata.Index
	rc.appliedIndex = snap.Metadata.Index

	defer rc.wal.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// event loop on raft state machine updates
	for {
		select {
		case <-ticker.C:
			rc.node.Tick()

		// store raft entries to wal, then publish over commit channel
		case rd := <-rc.node.Ready():
			rc.wal.Save(rd.HardState, rd.Entries)
			if !raft.IsEmptySnap(rd.Snapshot) {
				rc.saveSnap(rd.Snapshot)
				rc.raftStorage.ApplySnapshot(rd.Snapshot)
				rc.publishSnapshot(rd.Snapshot)
			}
			rc.raftStorage.Append(rd.Entries)
			rc.transport.Send(rd.Messages)
			applyDoneC, ok := rc.publishEntries(rc.entriesToApply(rd.CommittedEntries))
			if !ok {
				rc.stop()
				return
			}
			rc.maybeTriggerSnapshot(applyDoneC)
			rc.node.Advance()

		case err = <-rc.transport.ErrorC:
			rc.writeError(err)
			return

		case <-rc.stopc:
			rc.stop()
			return
		}
	}
}

func (rc *raftNode) serveRaft() {
	peer, _ := peerUrl(rc.Peers, rc.Name)
	url, err := url.Parse(peer)
	if err != nil {
		rc.logger.Sugar().Fatalf("failed parsing URL (%v)", err)
	}

	ln, err := newStoppableListener(url.Host, rc.httpstopc)
	if err != nil {
		rc.logger.Sugar().Fatalf("failed to listen rafthttp (%v)", err)
	}

	server := &http.Server{Handler: rc.transport.Handler()}

	err = server.Serve(ln)
	select {
	case <-rc.httpstopc:
	default:
		rc.logger.Sugar().Fatalf("failed to serve rafthttp (%v)", err)
	}
	close(rc.httpdonec)
}

func (rc *raftNode) Process(ctx context.Context, m raftpb.Message) error {
	return rc.node.Step(ctx, m)
}
func (rc *raftNode) IsIDRemoved(id uint64) bool  { return false }
func (rc *raftNode) ReportUnreachable(id uint64) { rc.node.ReportUnreachable(id) }
func (rc *raftNode) ReportSnapshot(id uint64, status raft.SnapshotStatus) {
	rc.node.ReportSnapshot(id, status)
}

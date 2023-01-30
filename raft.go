// MIT License
//
// Copyright (c) 2023 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package raft

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
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
	Get(ctx context.Context, option *GetOption) (*GetResult, error)
	Propose(ctx context.Context, data []byte) error
	ProposeConfChange(ctx context.Context, cc *raftpb.ConfChange) error
	Stop(ctx context.Context) error
	Err() <-chan error
}

type commit struct {
	frame      *frame
	applyDoneC chan<- struct{}
}

// A key-value stream backed by raft
type raftNode struct {
	Config

	ctx    context.Context
	cancel context.CancelFunc

	commitC chan<- *commit // entries committed to log (k,v)
	errorC  chan error     // errors from raft session

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing for the commit/error channel
	node        raft.Node
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL

	storage     *persistentStorage
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
func NewRaftNode(logger *zap.Logger, applier Applier, config Config) (RaftNode, error) {

	raft.SetLogger(newLogger(logger))

	var err error
	if !fileutil.Exist(config.snapdir) {
		err = os.MkdirAll(config.snapdir, 0755)
		if err != nil {
			return nil, err
		}
	}
	snapshotter := snap.New(logger, config.snapdir)

	ctx, cancel := context.WithCancel(context.Background())

	commitC := make(chan *commit, 10)
	errorC := make(chan error)

	storage, err := newPersistentStorage(ctx, logger, applier, snapshotter, commitC)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create raft storage: %v", err)
	}

	rc := &raftNode{
		Config:      config,
		ctx:         ctx,
		cancel:      cancel,
		commitC:     commitC,
		errorC:      errorC,
		stopc:       make(chan struct{}),
		httpstopc:   make(chan struct{}),
		httpdonec:   make(chan struct{}),
		storage:     storage,
		snapshotter: snapshotter,

		logger: logger,
	}

	// rest of structure populated after WAL replay
	err = rc.startRaft()
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func (rc *raftNode) Get(ctx context.Context, option *GetOption) (*GetResult, error) {
	return rc.storage.Get(ctx, option)
}

func (rc *raftNode) Propose(ctx context.Context, data []byte) error {
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

func (rc *raftNode) Err() <-chan error {
	return rc.errorC
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

	fr := &frame{}
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				// ignore empty messages
				break
			}

			fr.term, fr.index = ents[i].Term, ents[i].Index
			fr.data = append(fr.data, ents[i].Data)
		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			_ = cc.Unmarshal(ents[i].Data)
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

	if len(fr.data) > 0 {
		applyDoneC = make(chan struct{}, 1)
		select {
		case rc.commitC <- &commit{fr, applyDoneC}:
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
	if wal.Exist(rc.waldir) {
		walSnaps, err := wal.ValidSnapshotEntries(rc.logger, rc.waldir)
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
	if !wal.Exist(rc.waldir) {
		if err := os.Mkdir(rc.waldir, 0750); err != nil {
			return nil, fmt.Errorf("cannot create dir for wal (%v)", err)
		}

		w, err := wal.Create(rc.logger, rc.waldir, nil)
		if err != nil {
			return nil, fmt.Errorf("create wal error (%v)", err)
		}
		_ = w.Close()
	}

	walsnap := walpb.Snapshot{}
	if snapshot != nil {
		walsnap.Index, walsnap.Term = snapshot.Metadata.Index, snapshot.Metadata.Term
	}
	rc.logger.Sugar().Infof("loading WAL at term %d and index %d", walsnap.Term, walsnap.Index)
	w, err := wal.Open(rc.logger, rc.waldir, walsnap)
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
		// ignore error
		_ = rc.raftStorage.ApplySnapshot(*snapshot)
	}
	// ignore error
	_ = rc.raftStorage.SetHardState(st)

	// append to storage so raft starts at the right place in log, ignore error
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

func (rc *raftNode) startRaft() error {
	var err error

	oldwal := wal.Exist(rc.waldir)
	rc.wal = rc.replayWAL()

	rpeers := make([]raft.Peer, len(rc.Peers))
	for i, peer := range rc.Peers {
		name, uri, err := peerSplit(peer)
		if err != nil {
			return err
		}
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
		return err
	}
	for _, peer := range rc.Peers {
		name, uri, _ := peerSplit(peer)
		rc.transport.AddPeer(types.ID(hash(name)), []string{uri})
	}

	go rc.serveRaft()
	go rc.serveChannels()

	return err
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
		err := fmt.Errorf("snapshot index [%d] should > progress.appliedIndex [%d]", snapshotToSave.Metadata.Index, rc.appliedIndex)
		rc.logger.Sugar().Warn(err)
		rc.writeError(err)
		return
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
	snapshot, err := rc.raftStorage.Snapshot()
	if err != nil {
		err = fmt.Errorf("get snapshot (%v)", err)
		rc.logger.Sugar().Warn(err)
		rc.writeError(err)
		return
	}
	rc.confState = snapshot.Metadata.ConfState
	rc.snapshotIndex = snapshot.Metadata.Index
	rc.appliedIndex = snapshot.Metadata.Index

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
			err = rc.maybeTriggerSnapshot(applyDoneC)
			if err != nil {
				rc.logger.Sugar().Warn(err)
				rc.writeError(err)
				return
			}

			rc.node.Advance()

		case err = <-rc.transport.ErrorC:
			rc.logger.Sugar().Warn(err)
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
		err = fmt.Errorf("failed parsing URL (%v)", err)
		rc.logger.Sugar().Warn(err)
		rc.writeError(err)
		return
	}

	ln, err := newStoppableListener(url.Host, rc.httpstopc)
	if err != nil {
		err = fmt.Errorf("failed to listen rafthttp (%v)", err)
		rc.logger.Sugar().Warn(err)
	}

	server := &http.Server{Handler: rc.transport.Handler()}

	err = server.Serve(ln)
	select {
	case <-rc.httpstopc:
	default:
		err = fmt.Errorf("failed to serve rafthttp (%v)", err)
		rc.logger.Sugar().Warn(err)
		rc.writeError(err)
		return
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

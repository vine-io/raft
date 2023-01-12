package raft

import (
	"fmt"
	"path/filepath"
)

var (
	DefaultCluster                   string = "default"
	DefaultSnapshotCount             uint64 = 10000
	DefaultSnapshotCatchUpEntriesN   uint64 = 10
	DefaultElectionTick              uint32 = 10
	DefaultHeartBeatTick             uint32 = 1
	DefaultMaxSizePerMsg             uint64 = 1024 * 1024
	DefaultMaxInflightMsgs           uint32 = 256
	DefaultMaxUncommittedEntriesSize uint64 = 1 << 30
)

type Config struct {
	// client ID for raft session
	Id int
	// name for node
	Name string
	// path to wal and snapshot
	Dir     string
	waldir  string
	snapdir string
	// raft peer URLs
	Peers []string
	// name for raft cluster
	ClusterName string
	// id for raft cluster
	ClusterID int
	// node is joining an existing cluster
	Join bool
	//
	GetSnapshot func() ([]byte, error)
	// count for trigger to create snapshot
	SnapCount uint64
	//
	SnapshotCatchUpEntriesN uint64
	// election tick duration for raft
	ElectionTick uint32
	// heartbeat tick duration for raft
	HeartbeatTick uint32
	// max size for message raft
	MaxSizePerMsg uint64
	// max inflight messages for raft
	MaxInflightMsgs uint32
	// max uncommitted entries size for raft
	MaxUncommittedEntriesSize uint64
}

func NewConfig(name, dir string, peers []string, join bool) (Config, error) {

	if len(peers) != 0 {
		_, err := peerUrl(peers, name)
		if err != nil {
			return Config{}, fmt.Errorf("check dir and peers: %v", err)
		}

		for _, peer := range peers {
			_, _, err = peerSplit(peer)
			if err != nil {
				return Config{}, fmt.Errorf("check peers: %v", err)
			}
		}
	}

	cfg := Config{
		Id:                        int(hash(name)),
		Name:                      name,
		Dir:                       dir,
		waldir:                    filepath.Join(dir, "wal"),
		snapdir:                   filepath.Join(dir, "snap"),
		Peers:                     peers,
		ClusterName:               DefaultCluster,
		ClusterID:                 int(hash(DefaultCluster)),
		Join:                      join,
		SnapCount:                 DefaultSnapshotCount,
		SnapshotCatchUpEntriesN:   DefaultSnapshotCatchUpEntriesN,
		ElectionTick:              DefaultElectionTick,
		HeartbeatTick:             DefaultHeartBeatTick,
		MaxSizePerMsg:             DefaultMaxSizePerMsg,
		MaxInflightMsgs:           DefaultMaxInflightMsgs,
		MaxUncommittedEntriesSize: DefaultMaxUncommittedEntriesSize,
	}

	return cfg, nil
}

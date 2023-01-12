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

	"github.com/vine-io/apimachinery/schema"
	"github.com/vine-io/vine/lib/dao/clause"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"go.uber.org/zap"
)

type GetOption struct {
	GVK   schema.GroupVersionKind
	Id    string
	List  bool
	Page  int32
	Size  int32
	Exprs []clause.Expression
}

type GetResult struct {
	List  bool
	Out   interface{}
	Total *int64
}

type frame struct {
	term  uint64
	index uint64
	data  [][]byte
}

type Applier interface {
	Get(ctx context.Context, option *GetOption) (*GetResult, error)
	Put(ctx context.Context, data []byte) error
	GetEpoch(ctx context.Context) (uint64, uint64, error)
	ApplyEpoch(ctx context.Context, term, index uint64) error
	GetSnapshot(ctx context.Context) ([]byte, error)
	RecoverFromSnapshot(ctx context.Context, snapshot []byte) error
}

type persistentStorage struct {
	ctx         context.Context
	lg          *zap.Logger
	applier     Applier
	snapshotter *snap.Snapshotter
}

func newPersistentStorage(ctx context.Context, lg *zap.Logger, applier Applier, snapshotter *snap.Snapshotter, commitC <-chan *commit) (*persistentStorage, error) {

	s := &persistentStorage{
		lg:          lg,
		ctx:         ctx,
		applier:     applier,
		snapshotter: snapshotter,
	}

	snapshot, err := s.loadSnapshot()
	if err != nil {
		return nil, err
	}

	if snapshot != nil {
		if err = s.RecoverFromSnapshot(s.ctx, snapshot.Metadata.Term, snapshot.Metadata.Index, snapshot.Data); err != nil {
			return nil, err
		}
	}

	// read commits from raft into PersistentStorage Applier until error
	go s.readCommits(commitC)
	return s, nil
}

func (s *persistentStorage) Get(ctx context.Context, option *GetOption) (*GetResult, error) {
	return s.applier.Get(ctx, option)
}

func (s *persistentStorage) readCommits(commitC <-chan *commit) {
	for cc := range commitC {
		if cc == nil {
			// signaled to load snapshot
			snapshot, err := s.loadSnapshot()
			if err != nil {
				s.lg.Sugar().Fatalf(err.Error())
			}
			if snapshot != nil {
				s.lg.Sugar().Debugf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
				if err = s.RecoverFromSnapshot(s.ctx, snapshot.Metadata.Term, snapshot.Metadata.Index, snapshot.Data); err != nil {
					s.lg.Sugar().Panic(err)
				}
			}
			continue
		}

		cTerm, cIndex := cc.frame.term, cc.frame.index

		term, index, _ := s.applier.GetEpoch(s.ctx)

		if cTerm >= term && cIndex > index {
			for _, item := range cc.frame.data {
				_ = s.applier.Put(s.ctx, item)
			}
			_ = s.applier.ApplyEpoch(s.ctx, cTerm, cIndex)
		}

		close(cc.applyDoneC)
	}
}

func (s *persistentStorage) GetSnapshot(ctx context.Context) ([]byte, error) {
	return s.applier.GetSnapshot(ctx)
}

func (s *persistentStorage) loadSnapshot() (*raftpb.Snapshot, error) {
	snapshot, err := s.snapshotter.Load()
	if err == snap.ErrNoSnapshot {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *persistentStorage) RecoverFromSnapshot(ctx context.Context, term, index uint64, snapshot []byte) error {

	err := s.applier.RecoverFromSnapshot(ctx, snapshot)
	if err != nil {
		return err
	}

	return s.applier.ApplyEpoch(ctx, term, index)
}

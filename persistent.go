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

	"github.com/vine-io/apimachinery/schema"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
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

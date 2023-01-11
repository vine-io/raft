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

	json "github.com/json-iterator/go"
	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/vine/lib/dao/clause"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"go.uber.org/zap"
)

type ListOption struct {
	In         runtime.Object
	Page, Size int32
	Exprs      []clause.Expression
}

type Op string

const (
	Create Op = "create"
	Update Op = "update"
	Delete Op = "delete"
)

type Operation struct {
	Op     Op             `json:"op"`
	Target runtime.Object `json:"target"`
}

func (o *Operation) GetOp() Op {
	return o.Op
}

func (o *Operation) GetObject() runtime.Object {
	return o.Target
}

func (o *Operation) Marshal() ([]byte, error) {
	return json.Marshal(o)
}

func (o *Operation) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &o)
}

type CommitData struct {
	Term  uint64
	Index uint64
	Ops   []*Operation
}

type IterFunc func(item runtime.Object, index int)

type Applier interface {
	List(ctx context.Context, option ListOption, iterFunc IterFunc) (int64, error)
	Get(ctx context.Context, in runtime.Object, id string) error
	Create(ctx context.Context, in runtime.Object) error
	Update(ctx context.Context, in runtime.Object) error
	Delete(ctx context.Context, in runtime.Object, soft bool) error
	GetEpoch(ctx context.Context) (uint64, uint64, error)
	SetEpoch(ctx context.Context, term, index uint64) error
	GetSnapshot(ctx context.Context) ([]byte, error)
	RecoverFromSnapshot(ctx context.Context, snapshot []byte) error
}

type PersistentStorage interface {
	List(ctx context.Context, option ListOption, iterFunc IterFunc) (int64, error)
	Get(ctx context.Context, in runtime.Object, id string) error
	GetSnapshot(ctx context.Context) ([]byte, error)
	RecoverFromSnapshot(ctx context.Context, term, index uint64, snapshot []byte) error
}

type persistentStorage struct {
	ctx         context.Context
	lg          *zap.Logger
	applier     Applier
	snapshotter *snap.Snapshotter
}

func NewPersistentStorage(lg *zap.Logger, applier Applier, snapshotter *snap.Snapshotter, commitC <-chan *commit) (PersistentStorage, error) {

	s := &persistentStorage{
		lg:          lg,
		ctx:         context.Background(),
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

func (s *persistentStorage) List(ctx context.Context, option ListOption, iter IterFunc) (int64, error) {
	return s.applier.List(ctx, option, iter)
}

func (s *persistentStorage) Get(ctx context.Context, in runtime.Object, id string) error {
	return s.applier.Get(ctx, in, id)
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

		cTerm, cIndex := cc.data.Term, cc.data.Index

		term, index, _ := s.applier.GetEpoch(s.ctx)

		if cTerm >= term && cIndex > index {
			for _, o := range cc.data.Ops {
				switch o.GetOp() {
				case Create:
					s.applier.Create(s.ctx, o.GetObject())
				case Update:
					s.applier.Update(s.ctx, o.GetObject())
				case Delete:
					s.applier.Delete(s.ctx, o.GetObject(), true)
				}
			}

			_ = s.applier.SetEpoch(s.ctx, cTerm, cIndex)
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

	return s.applier.SetEpoch(ctx, term, index)
}

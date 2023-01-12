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

package main

import (
	"context"
	"os"

	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/apimachinery/storage"
	"github.com/vine-io/plugins/dao/sqlite"
	"github.com/vine-io/raft"
	pb "github.com/vine-io/raft/test/proto"
	"github.com/vine-io/vine"
	"github.com/vine-io/vine/lib/dao"
	"go.etcd.io/etcd/server/v3/etcdserver/api/snap"
	"go.uber.org/zap"
)

func main() {
	logger := zap.NewExample()

	_ = os.MkdirAll("snap", 0755)
	snapshotter := snap.New(logger, "snap")

	dialect := sqlite.NewDialect(sqlite.DriverName("sqlite3"), dao.DSN("db.sqlite3"))
	if err := dialect.Init(); err != nil {
		logger.Sugar().Fatal(err)
	}
	dao.DefaultDialect = dialect

	s := vine.NewService(vine.Address("127.0.0.1:33379"), vine.Dialect(dialect))

	s.Init()

	scheme := runtime.NewScheme()
	if err := pb.AddToScheme(scheme); err != nil {
		logger.Sugar().Fatal(err)
	}

	factory := storage.NewStorageFactory()
	if err := pb.AddToFactory(factory); err != nil {
		logger.Sugar().Fatal(err)
	}

	applier, err := NewApplier(scheme, factory)
	if err != nil {
		logger.Sugar().Error(err)
	}
	_ = applier
	_ = snapshotter

	commitC := raft.NewCommitChannel(10)
	errorC := make(chan error)

	ps, err := raft.NewPersistentStorage(logger, applier, snapshotter, commitC)
	if err != nil {
		logger.Sugar().Error(err)
	}

	getSnapshot := func() ([]byte, error) { return ps.GetSnapshot(context.TODO()) }
	config, err := raft.NewConfig("raft1", "wal", []string{"raft1=http://127.0.0.1:12380"}, false, getSnapshot)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	raftNode := raft.NewRaftNode(logger, snapshotter, config, commitC, errorC)

	server := NewServer(raftNode, scheme, ps)

	if err = pb.RegisterTestHandler(s.Server(), server); err != nil {
		logger.Sugar().Fatal(err)
	}

	if err = s.Run(); err != nil {
		logger.Sugar().Fatal(err)
	}
}

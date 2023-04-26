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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/apimachinery/storage"
	"github.com/vine-io/plugins/logger/zap"
	"github.com/vine-io/raft"
	pb "github.com/vine-io/raft/test/proto"
	"github.com/vine-io/raft/test/server/hello"
	"github.com/vine-io/vine"
	log "github.com/vine-io/vine/lib/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dir := flag.String("dir", ".", "data directory")
	name := flag.String("name", "raft1", "name for raft node")
	peer := flag.String("peer", "raft1=http://127.0.0.1:33380", "raft cluster peer")
	address := flag.String("addr", "127.0.0.1:33379", "service address")

	flag.Parse()

	zapLog, err := zap.New(zap.WithJSONEncode())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.DefaultLogger = zapLog
	logger := zapLog.GetLogger()

	_ = os.MkdirAll(*dir, 0755)

	s := vine.NewService(vine.Address(*address), vine.Cmd(nil))

	s.Init()

	scheme := runtime.NewScheme()
	if err := pb.AddToScheme(scheme); err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(*dir, "raft.db")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	factory := storage.NewStorageFactory()
	if err = pb.AddToBuilder(db, factory); err != nil {
		log.Fatal(err)
	}

	applier, err := hello.NewApplier(db, scheme, factory)
	if err != nil {
		log.Error(err)
	}

	config, err := raft.NewConfig(*name, *dir, strings.Split(*peer, ","), false)
	if err != nil {
		log.Fatal(err)
	}

	raftNode, err := raft.NewRaftNode(logger, applier, config)
	if err != nil {
		log.Fatal(err)
	}

	server := hello.NewServer(raftNode, scheme)

	if err = pb.RegisterTestHandler(s.Server(), server); err != nil {
		log.Fatal(err)
	}

	if err = s.Run(); err != nil {
		log.Fatal(err)
	}
}

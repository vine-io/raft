# 介绍

[raft](https://github.com/vine-io/raft) 提供 raft 算法功能，可以使用 [vine] (https://github.com/vine-io/vine) 配合使用

```bash
// 核心接口
type RaftNode interface {
	Get(ctx context.Context, option *GetOption) (*GetResult, error)
	Propose(ctx context.Context, data []byte) error
	ProposeConfChange(ctx context.Context, cc *raftpb.ConfChange) error
	Stop(ctx context.Context) error
	Err() <-chan error
}
```

# 启动集群

[examples](https://github.com/vine-io/raft/examples) 目录中提供一个简单的实例

编译 server
```bash
cd examples
mkdir _output
go build -tags json1 -o _output/raft-server server/main.go
```

启动集群
```bash
cd _output

./raft-server -dir=raft1 -name raft1 --peer=raft1=http://127.0.0.1:12380,raft2=http://127.0.0.1:22380,raft3=http://127.0.0.1:32380 -addr=127.0.0.1:12379

./raft-server -dir=raft2 -name raft2 --peer=raft1=http://127.0.0.1:12380,raft2=http://127.0.0.1:22380,raft3=http://127.0.0.1:32380 -addr=127.0.0.1:22379

./raft-server -dir=raft3 -name raft3 --peer=raft1=http://127.0.0.1:12380,raft2=http://127.0.0.1:22380,raft3=http://127.0.0.1:32380 -addr=127.0.0.1:32379
```

# 验证
```bash
cd examples

go run client/main.go -addr=127.0.0.1:12379 -company=c17
```

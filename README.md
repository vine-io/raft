# 编译
```bash
mkdir _output
go build -o _output/raft-gorm
```

# 启动集群
```bash
cd _output
./raft-gorm --id 1 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 12380

./raft-gorm --id 2 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 22380

./raft-gorm --id 3 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 32380
```

# 验证
```bash
curl -L http://127.0.0.1:12380 -XPUT -d '{"op": "create", "target": { "code": "1234", "price": 122}}'

curl -L http://127.0.0.1:12380\?id\=1                                                                  
{"id":1,"createdAt":"2022-12-20T11:34:28.144063+08:00","updatedAt":"2022-12-20T11:34:28.144063+08:00","deletedAt":null,"code":"1234","price":122}%

curl -L http://127.0.0.1:12380 -XPUT -d '{"op": "update", "target": {"id":1, "code": "12345", "price": 122}}'

curl -L http://127.0.0.1:22380\?id\=1
{"id":1,"createdAt":"2022-12-20T11:34:28.144114+08:00","updatedAt":"2022-12-20T11:35:17.15704+08:00","deletedAt":null,"code":"12345","price":122}%                                        

curl -L http://127.0.0.1:12380/1 -XPUT -d '{"op": "delete", "target": {"id":1, "code": "12345", "price": 122}}'

curl -L http://127.0.0.1:32380\?id\=1                                                                         
Failed to GET
```# raft

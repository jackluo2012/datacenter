#!/bin/bash
## author by jackluo
## Email net.webjoy@gmail.com
## wechat 13228191831
##### 启动mysql 服务
docker kill mysql 2>/dev/null
docker rm mysql 2>/dev/null
docker run --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=admin -v $(pwd)/mysql8:/var/lib/mysql -d mysql:8.4

##### 启动redis 服务（如宿主机 6379 被占用，请自行调整映射端口）
docker stop redis 2>/dev/null
docker rm redis 2>/dev/null
echo "Start Redis Service..."
docker run --name redis -d \
  --publish 6379:6379 \
  redis:8-alpine \
  redis-server --requirepass admin

### 启动 etcd 服务
rm -rf $(pwd)/etcd && mkdir -p $(pwd)/etcd && \
  docker stop etcd 2>/dev/null
  docker rm etcd 2>/dev/null || true && \
  docker run -d \
  -p 2379:2379 \
  -p 2380:2380 \
  --mount type=bind,source=$(pwd)/etcd,destination=/etcd-data \
  --name etcd \
  quay.io/coreos/etcd:v3.5.21 \
  /usr/local/bin/etcd \
  --name s1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster s1=http://0.0.0.0:2380 \
  --initial-cluster-token tkn \
  --initial-cluster-state new


#### 启动elasticsearch 服务
docker stop es 2>/dev/null
docker rm es 2>/dev/null

docker run -d --name es \
	-p 9200:9200 -p 9300:9300 \
	-e "discovery.type=single-node" \
	-e "xpack.security.enabled=true" \
	-e "ELASTIC_PASSWORD=admin" \
	-v $(pwd)/elasticsearch:/usr/share/elasticsearch/data \
	docker.elastic.co/elasticsearch/elasticsearch:7.17.28

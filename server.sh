#!/bin/bash
## author by jackluo
## Email net.webjoy@gmail.com
## wechat 13228191831
##### 启动mysql 服务
docker kill mysql
docker rm mysql
docker run --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=admin -v $(pwd)/mysql8:/var/lib/mysql -d mysql:8.0.21

##### 启动redis 服务
docker stop redis
docker rm redis
echo "Start Redis Service..."
docker run --name redis -d \
  --publish 6379:6379 \
  --env 'REDIS_PASSWORD=admin' \
  --volume $(pwd)/redis:/var/lib/redis \
  sameersbn/redis:latest

### 启动 etcd 服务
#!/bin/bash
rm -rf $(pwd)/etcd && mkdir -p $(pwd)/etcd && \
  docker stop etcd && \
  docker rm etcd || true && \
  docker run -d \
  -p 2379:2379 \
  -p 2380:2380 \
  --mount type=bind,source=$(pwd)/etcd,destination=/etcd-data \
  --name etcd \
  quay.io/coreos/etcd \
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
docker stop es
docker rm es

docker run -d --name es \
	-p 9200:9200 -p 9300:9300 \
	-e "discovery.type=single-node" \
	-v $(pwd)/elasticsearch:/usr/share/elasticsearch/data \
	spencezhou/elasticsearch:7.6.2
# 基于go-zero 框架写的一个数据中台中心

详细介绍 [我是如何用go-zero实现一个中台系统的](https://www.cnblogs.com/jackluo/p/14148518.html)

### 架构图
![中台系统](https://img2020.cnblogs.com/blog/203395/202012/203395-20201217094615171-335437652.jpg "中台架构")

## 如何运行
### 先启动mysql redis etcd 服务
```shell
sh server.sh
```
### 输出如下 就显示 成功了
``` 
mysql
mysql
38dc8ec735dc64fb545c6d046bc2c61d1b45533027a2ff0eec5a67842317ffd0
redis
redis
Start Redis Service...
2127bb7462ad2ddf5b86a596f3392d5b38c32245cbf4a8d5ff57f66aea087313
etcd
etcd
7eddff3f1ce27b029c53c91a55302bb29b4b5c41b902fa707008fd4062d6a2e
```
接着导入 sql.sql到 mysql数据中 ,如果有工具自行导入,下面仅参考
```
mysql -uroot -padmin
mysql > set name utf8
mysql > create databse datacenter
mysql > use datacenter
mysql > source sql.sql
```
### 然后分别把配置文件 ,文件下面分别对应了一个rpc.example.yaml的文件，复制，基本就没有问题

```
vi etc/datacenter-api.yaml #网关配置
vi user/rpc/etc/rpc.yaml #用户信息配置
vi common/rpc/etc/rpc.yaml #公共配置
vi common/rpc/etc/rpc.yaml #公共配置
vi votes/rpc/etc/rpc.yaml #投票配置
```
### 然后启动 服务 ,应该我们要启动
```
sh restart.sh
```
### 输出如下
```
➜  datacenter.bak git:(master) ✗ sh restart.sh              
appending output to nohup.out
appending output to nohup.out
appending output to nohup.out
appending output to nohup.out    
```
### 可以分别查看是否启动成功
```
tail -F nohup.out  #网关的服务
tail -F user/rpc/nohup.out #用户的rpc服务
tail -F common/rpc/nohup.out #公共的
tail -F votes/rpc/nohup.out #投票的
```

### 在postman 导入数据 数据中台中心.postman_collection.json  就可以很愉快的玩耍了
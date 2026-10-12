#!/bin/bash
## author by jackluo
## Email net.webjoy@gmail.com
## wechat 13228191831

getewayPath=$(pwd) #网关服务
userPath=$(pwd)/user/rpc #用户服务
commonPath=$(pwd)/common/rpc #公共服务
votesPath=$(pwd)/votes/rpc #投票服务
configPath=/etc/rpc.yaml #配置文件
geteWayCnf=/etc/datacenter-api.yaml
UserRpc=user-server #定义网关服务
CommonRpc=common-server #定义公共服务
VotesRpc=votes-server #定义投票服务
geteWayApi=datacenter-server #定义网关服务


RpcServer(){
    mydir=$1
    myserver=$2
    mycnf=$3
    cd ${mydir}
    go build -o ${myserver} $mydir/rpc.go
    kill -9 $(ps -ef|grep "${myserver}"|awk '{print $2}') >/dev/null 2>&1
    nohup ${mydir}/${myserver} -f ${mydir}${mycnf} &
}
RpcServerPlus(){
    mydir=$1/$2/rpc
    myserver=${2}-server
    mycnf=/etc/${2}.yaml
    cd ${mydir}
    go build -o ${myserver} $mydir/$2.go
    kill -9 $(ps -ef|grep "${myserver}"|awk '{print $2}') >/dev/null 2>&1
    nohup ${mydir}/${myserver} -f ${mydir}${mycnf} &
}

StartServer(){
    mydir=$1
    myserver=$2
    mycnf=$3
    cd ${mydir}
    go build -o ${myserver} $mydir/datacenter.go
    kill -9 $(ps -ef|grep "${myserver}"|awk '{print $2}') >/dev/null 2>&1
    nohup ${mydir}/${myserver} -f ${mydir}${mycnf} &
}
#公共服务
RpcServer ${commonPath} ${CommonRpc} ${configPath}
#用户服务
RpcServer ${userPath} ${UserRpc} ${configPath}
#投票服务
RpcServer ${votesPath} ${VotesRpc} ${configPath}
#搜索服务
RpcServerPlus ${getewayPath} search
#数据源服务（Phase 1）
RpcServerPlus ${getewayPath} datasource
#元数据服务（Phase 1）
RpcServerPlus ${getewayPath} metadata
#数据服务API（Phase 1）
RpcServerPlus ${getewayPath} datasvc
#资产目录服务（Phase 1）
RpcServerPlus ${getewayPath} catalog
#数据质量服务（Phase 2）
RpcServerPlus ${getewayPath} quality
#数据指标服务（Phase 2）
RpcServerPlus ${getewayPath} metric
#数据标签服务（Phase 2）
RpcServerPlus ${getewayPath} tag
#调度集成服务（Phase 3）
RpcServerPlus ${getewayPath} job
#SQL IDE 服务（Phase 3）
RpcServerPlus ${getewayPath} ide
#AI 问数服务（Phase 4）
RpcServerPlus ${getewayPath} ai
#数据标准服务（Phase 5）
RpcServerPlus ${getewayPath} standard
#主数据服务（Phase 5）
RpcServerPlus ${getewayPath} mdm
#数据运维与日志服务（Phase 5）
RpcServerPlus ${getewayPath} ops
#数仓建模服务（Phase 6）
RpcServerPlus ${getewayPath} theme
#数据安全服务（Phase 6）
RpcServerPlus ${getewayPath} security
#问答抽奖服务
RpcServerPlus ${getewayPath} questions
#调度引擎（内置轻量调度：扫 dc_quality_job/dc_ops_archive_task/dc_theme_mv 的 cron 到期任务，
#经 etcd 调 quality/ops/theme rpc 执行；无 yaml 配置，DSN/etcd 走环境变量，
#默认连 127.0.0.1:3306 / 127.0.0.1:2379，与 server.sh 起的依赖端口映射一致；
#单实例运行，勿多副本。详见 docs/delivery/01-部署手册.md §8）
SchedulerRun(){
    cd ${getewayPath}
    go build -o scheduler ./cmd/scheduler
    kill -9 $(ps -ef|grep "\./scheduler"|awk '{print $2}') >/dev/null 2>&1
    nohup ./scheduler >> scheduler.log 2>&1 &
}
SchedulerRun
#Api服务
StartServer ${getewayPath} ${geteWayApi} ${geteWayCnf}



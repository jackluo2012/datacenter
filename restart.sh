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
#Api服务
StartServer ${getewayPath} ${geteWayApi} ${geteWayCnf}



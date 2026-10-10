// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package job

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/job/rpc/jobClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SchedulerInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 调度器连接信息
func NewSchedulerInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SchedulerInfoLogic {
	return &SchedulerInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 调度器连接信息: 该接口按租户查询(不针对单个任务, id 置 0), 转发后把连接信息转为网关类型返回
func (l *SchedulerInfoLogic) SchedulerInfo(req *types.SyncTaskIdReq) (resp *types.SchedulerInfoResp, err error) {
	reply, err := l.svcCtx.JobRpc.GetSchedulerInfo(l.ctx, &jobClient.IdReq{
		Id:   0,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.SchedulerInfoResp{
		Configured: reply.GetConfigured(),
		Connected:  reply.GetConnected(),
		BaseUrl:    reply.GetBaseUrl(),
		Msg:        reply.GetMsg(),
	}, nil
}

package logic

import (
	"context"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSchedulerInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSchedulerInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSchedulerInfoLogic {
	return &GetSchedulerInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 调度器连接信息（运维视图：是否配置 / 当前是否可达 / 基址）
func (l *GetSchedulerInfoLogic) GetSchedulerInfo(in *job.IdReq) (*job.SchedulerReply, error) {
	client := ds()
	if !client.enabled() {
		return &job.SchedulerReply{
			Configured: false,
			Connected:  false,
			BaseUrl:    "",
			Msg:        "未配置 DolphinScheduler：需设置环境变量 DS_BASE_URL/DS_TOKEN/DS_PROJECT",
		}, nil
	}
	reply := &job.SchedulerReply{
		Configured: true,
		BaseUrl:    client.base,
	}
	// 连通性探测：拉取项目列表第一页（校验地址可达与 token 有效）
	if err := client.ping(); err != nil {
		l.Errorf("调度器连通性探测失败: %v", err)
		reply.Connected = false
		reply.Msg = "调度器不可达: " + err.Error()
		return reply, nil
	}
	reply.Connected = true
	reply.Msg = "调度器连接正常"
	return reply, nil
}

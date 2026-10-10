package logic

import (
	"context"
	"fmt"

	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskStatusLogic {
	return &GetTaskStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务状态（本地 status + 若已部署则聚合调度器最新实例状态，展示口径不落库）
func (l *GetTaskStatusLogic) GetTaskStatus(in *job.IdReq) (*job.StatusReply, error) {
	task, err := loadTask(l.ctx, l.svcCtx.SyncTaskModel, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	reply := &job.StatusReply{Status: task.Status}
	// 未部署：无调度器实例可查，直接展示本地状态
	if task.DsWorkflowCode <= 0 {
		reply.Msg = "任务未部署，展示本地状态"
		return reply, nil
	}
	client := ds()
	if !client.enabled() {
		reply.Msg = "调度器未配置，展示本地状态"
		return reply, nil
	}
	// 查询调度器最新实例并聚合（DS 状态可识别时以调度器为准）
	state, instanceId, err := client.latestInstance(int(task.DsWorkflowCode))
	if err != nil {
		l.Errorf("查询任务 %d 的 DS 实例状态失败: %v", task.Id, err)
		reply.Msg = "调度器不可达，展示本地状态: " + err.Error()
		return reply, nil
	}
	if state == "" {
		reply.Msg = fmt.Sprintf("调度器暂无该任务（工作流编码 %d）的运行实例，展示本地状态", task.DsWorkflowCode)
		return reply, nil
	}
	reply.Status = AggregateStatus(task.Status, state)
	reply.DsInstanceId = int32(instanceId)
	reply.Msg = fmt.Sprintf("调度器最新实例状态 %s（工作流编码 %d）", state, task.DsWorkflowCode)
	return reply, nil
}

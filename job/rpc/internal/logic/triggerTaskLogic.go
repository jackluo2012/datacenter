package logic

import (
	"context"
	"time"

	"datacenter/job/model"
	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type TriggerTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTriggerTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TriggerTaskLogic {
	return &TriggerTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动触发一次（未部署 → 业务错误；DS 未配置/不可达/启动失败 → 业务结果 ok=false 带原因）
func (l *TriggerTaskLogic) TriggerTask(in *job.IdReq) (*job.TriggerReply, error) {
	task, err := loadTask(l.ctx, l.svcCtx.SyncTaskModel, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	// 未部署（无工作流编码）不允许触发
	if task.DsWorkflowCode <= 0 {
		return nil, ErrTaskNotDeployed
	}
	client := ds()
	if !client.enabled() {
		return &job.TriggerReply{Ok: false, Msg: "调度器未配置：需设置环境变量 DS_BASE_URL/DS_TOKEN/DS_PROJECT"}, nil
	}
	// 调 DS 启动一次工作流实例
	instanceId, err := client.start(int(task.DsWorkflowCode))
	if err != nil {
		l.Errorf("触发任务 %d 失败(工作流 %d): %v", task.Id, task.DsWorkflowCode, err)
		return &job.TriggerReply{Ok: false, Msg: dsErrMsg("调度器启动实例失败", err)}, nil
	}
	// 触发成功：状态流转 running 并回填最近运行时间
	if err := l.svcCtx.SyncTaskModel.UpdateRunStatus(l.ctx, task.Id, in.Beid, model.TaskStatusRunning, time.Now()); err != nil {
		return nil, err
	}
	l.Infof("任务 %d 已触发，DS 实例 id %d", task.Id, instanceId)
	return &job.TriggerReply{Ok: true, DsProcessInstanceId: int32(instanceId), Msg: "已触发"}, nil
}

package logic

import (
	"context"

	"datacenter/job/model"
	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeployToSchedulerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeployToSchedulerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployToSchedulerLogic {
	return &DeployToSchedulerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 部署到调度器（DolphinScheduler；未配置/不可达时 ok=false 带原因——部署失败是业务结果而非 error）：
// 幂等确保工作流定义存在（同名复用）→ 上线（releaseState=ONLINE）→ 回填工作流编码与 deployed 状态。
// TODO(联调): 需真实 DolphinScheduler 3.x 实例验证；任务的 cron 定时调度未实现
// （cron 仅存档，上线后需在 DS 侧对该工作流提交 schedule 定时，联调后可在此补充）
func (l *DeployToSchedulerLogic) DeployToScheduler(in *job.IdReq) (*job.DeployReply, error) {
	task, err := loadTask(l.ctx, l.svcCtx.SyncTaskModel, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	client := ds()
	if !client.enabled() {
		return &job.DeployReply{Ok: false, Msg: "调度器未配置：需设置环境变量 DS_BASE_URL/DS_TOKEN/DS_PROJECT"}, nil
	}
	// 幂等创建/复用工作流定义
	code, err := client.ensureWorkflow(task.Name)
	if err != nil {
		l.Errorf("部署任务 %d 创建工作流失败: %v", task.Id, err)
		return &job.DeployReply{Ok: false, Msg: dsErrMsg("创建工作流失败", err)}, nil
	}
	// 上线工作流（未上线不允许被调度/手动运行）
	if err := client.release(code); err != nil {
		l.Errorf("部署任务 %d 上线工作流 %d 失败: %v", task.Id, code, err)
		return &job.DeployReply{Ok: false, DsWorkflowCode: int32(code), Msg: dsErrMsg("上线工作流失败", err)}, nil
	}
	// 回填工作流编码并流转到 deployed
	if err := l.svcCtx.SyncTaskModel.UpdateDeploy(l.ctx, task.Id, in.Beid, int64(code), model.TaskStatusDeployed); err != nil {
		return nil, err
	}
	l.Infof("任务 %d 已部署到调度器，工作流编码 %d", task.Id, code)
	return &job.DeployReply{Ok: true, DsWorkflowCode: int32(code), Msg: "部署成功"}, nil
}

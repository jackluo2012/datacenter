package logic

import (
	"context"
	"strings"

	"datacenter/job/model"
	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSyncTaskLogic {
	return &CreateSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建同步任务（落库为草稿态，部署到调度器走 DeployToScheduler）
func (l *CreateSyncTaskLogic) CreateSyncTask(in *job.CreateSyncTaskReq) (*job.TaskReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	// 名称/mode 白名单（offline/cdc）/source_ds_id>0/mapping_json 合法性静态校验
	name := strings.TrimSpace(in.Name)
	if err := ValidateSyncTask(name, in.Mode, in.SourceDsId, in.MappingJson); err != nil {
		return nil, err
	}
	id, err := l.svcCtx.SyncTaskModel.Insert(l.ctx, &model.SyncTask{
		Name:        name,
		SourceDsId:  in.SourceDsId,
		TargetDsId:  in.TargetDsId,
		Mode:        in.Mode,
		MappingJson: strings.TrimSpace(in.MappingJson),
		Cron:        strings.TrimSpace(in.Cron),
		// V1 为 DS 推送任务，无内置引擎字段：源查询/目标库/目标表留空，batch_size 落默认值
		SourceSql:   "",
		TargetDb:    "",
		TargetTable: "",
		BatchSize:   model.DefaultSyncBatchSize,
		Beid:        in.Beid,
	})
	if err != nil {
		return nil, err
	}
	task, err := loadTask(l.ctx, l.svcCtx.SyncTaskModel, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &job.TaskReply{Task: taskToPb(task)}, nil
}

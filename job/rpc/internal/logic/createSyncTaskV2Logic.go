package logic

import (
	"context"
	"strings"

	"datacenter/job/model"
	"datacenter/job/rpc/internal/svc"
	"datacenter/job/rpc/job"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSyncTaskV2Logic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateSyncTaskV2Logic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSyncTaskV2Logic {
	return &CreateSyncTaskV2Logic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建同步任务（内置引擎：源 SQL→目标表 分批同步，落库为草稿态，RunSync 可直跑）。
// 与 V1（DS 推送）并存：本入口新增 source_sql/target_table/batch_size 列
func (l *CreateSyncTaskV2Logic) CreateSyncTaskV2(in *job.SaveSyncTaskReq) (*job.TaskReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	name := strings.TrimSpace(in.Name)
	// 静态校验：名称/mode/源数据源/源 SQL（SELECT 单语句）/目标表白名单/列映射数组
	//（batch_size 不校验拒绝：落库/执行时统一经 NormalizeBatchSize 归一）
	if err := ValidateSyncTaskV2(name, in.Mode, in.SourceDsId,
		in.SourceSql, in.TargetTable, in.MappingJson); err != nil {
		return nil, err
	}
	id, err := l.svcCtx.SyncTaskModel.Insert(l.ctx, &model.SyncTask{
		Name:       name,
		SourceDsId: in.SourceDsId,
		TargetDsId: in.TargetDsId,
		Mode:       in.Mode,
		// mapping_json 原样存档（ParseMapping 已校验为 [{src,dst,type}] 数组或空）
		MappingJson: strings.TrimSpace(in.MappingJson),
		Cron:        strings.TrimSpace(in.Cron),
		SourceSql:   strings.TrimSpace(in.SourceSql),
		TargetDb:    "", // proto 未透出目标库：走目标数据源 DSN 默认库
		TargetTable: strings.TrimSpace(in.TargetTable),
		BatchSize:   model.NormalizeBatchSize(int64(in.BatchSize)), // <=0 默认 1000、上限 10000
		Beid:        in.Beid,
	})
	if err != nil {
		return nil, err
	}
	task, err := loadTask(l.ctx, l.svcCtx.SyncTaskModel, id, in.Beid)
	if err != nil {
		return nil, err
	}
	l.Infof("内置引擎同步任务已创建: id=%d name=%s target=%s batch=%d",
		task.Id, task.Name, task.TargetTable, task.BatchSize)
	return &job.TaskReply{Task: taskToPb(task)}, nil
}

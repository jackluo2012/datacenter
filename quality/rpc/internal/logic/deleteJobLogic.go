package logic

import (
	"context"
	"errors"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteJobLogic {
	return &DeleteJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除任务：存在性校验（带租户隔离）后删除任务行；
// 结果历史（dc_quality_result）不级联清理，保留供问题追溯
func (l *DeleteJobLogic) DeleteJob(in *quality.IdReq) (*quality.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidRuleParams
	}
	// 存在性校验，任务不存在/跨租户不可见直接拒绝
	if _, err := l.svcCtx.JobModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.JobModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		l.Errorf("删除质量任务失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	return &quality.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

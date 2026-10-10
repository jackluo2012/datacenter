package logic

import (
	"context"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListJobsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListJobsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListJobsLogic {
	return &ListJobsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 任务列表（分页）
func (l *ListJobsLogic) ListJobs(in *quality.JobListReq) (*quality.JobListReply, error) {
	// 多租户条件必填
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.JobModel.FindList(l.ctx, in.Beid, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	jobs := make([]*quality.QualityJob, 0, len(list))
	for _, j := range list {
		jobs = append(jobs, jobToPb(j))
	}
	return &quality.JobListReply{List: jobs, Total: total}, nil
}

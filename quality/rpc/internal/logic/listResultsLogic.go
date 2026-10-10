package logic

import (
	"context"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListResultsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListResultsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListResultsLogic {
	return &ListResultsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 结果列表（job_id + 可选 pass_filter 分页；pass_filter 0.全部 1.仅通过 2.仅不通过）
func (l *ListResultsLogic) ListResults(in *quality.ResultListReq) (*quality.ResultListReply, error) {
	if in.JobId <= 0 {
		return nil, model.ErrJobIdRequired
	}
	list, total, err := l.svcCtx.ResultModel.FindListByJob(l.ctx, in.JobId, in.PassFilter, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	results := make([]*quality.QualityResult, 0, len(list))
	for _, r := range list {
		results = append(results, resultToPb(r))
	}
	return &quality.ResultListReply{List: results, Total: total}, nil
}

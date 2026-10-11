package logic

import (
	"context"

	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListErrorDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListErrorDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListErrorDataLogic {
	return &ListErrorDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 错误数据下钻：job_id 必带 + rule_id 可选过滤的分页查询（rows_json 内含违规行样本明细）
func (l *ListErrorDataLogic) ListErrorData(in *quality.ErrorDataListReq) (*quality.ErrorDataListReply, error) {
	list, total, err := l.svcCtx.ErrorDataModel.FindList(l.ctx, in.JobId, in.RuleId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	items := make([]*quality.ErrorData, 0, len(list))
	for _, e := range list {
		items = append(items, errorDataToPb(e))
	}
	return &quality.ErrorDataListReply{List: items, Total: total}, nil
}

package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCheckReportsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCheckReportsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCheckReportsLogic {
	return &ListCheckReportsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 核对历史（按标准模型分页，报告按时间倒序）
func (l *ListCheckReportsLogic) ListCheckReports(in *standard.CheckListReq) (*standard.CheckListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.CheckReportModel.FindListByModel(l.ctx, in.Beid, in.ModelId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	reports := make([]*standard.StdCheckReport, 0, len(list))
	for _, r := range list {
		reports = append(reports, reportToPb(r))
	}
	return &standard.CheckListReply{List: reports, Total: total}, nil
}

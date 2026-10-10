// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package standard

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/standard/rpc/standardClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type StdCheckListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标准核对历史
func NewStdCheckListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StdCheckListLogic {
	return &StdCheckListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 标准核对历史: 页码默认值处理后转发(model_id 为 0 表示不过滤), 业务错误直接透传
func (l *StdCheckListLogic) StdCheckList(req *types.StdCheckListReq) (resp *types.StdCheckListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.StandardRpc.ListCheckReports(l.ctx, &standardClient.CheckListReq{
		ModelId: req.ModelId,
		Beid:    req.Beid,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.StdCheckReport, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *checkReportToTypes(item))
	}
	return &types.StdCheckListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}

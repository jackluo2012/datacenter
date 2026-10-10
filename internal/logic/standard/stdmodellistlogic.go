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

type StdModelListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标准模型列表
func NewStdModelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StdModelListLogic {
	return &StdModelListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 标准模型列表: 页码默认值处理后转发(status 为发布状态过滤, 空串表示不过滤), 业务错误直接透传
func (l *StdModelListLogic) StdModelList(req *types.StdModelListReq) (resp *types.StdModelListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.StandardRpc.ListModels(l.ctx, &standardClient.ModelListReq{
		Beid:    req.Beid,
		Status:  req.Status,
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.StdModel, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *stdModelToTypes(item))
	}
	return &types.StdModelListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}

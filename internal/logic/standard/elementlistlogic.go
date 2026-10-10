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

type ElementListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 数据元列表
func NewElementListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ElementListLogic {
	return &ElementListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 数据元列表: 页码默认值处理后转发, 业务错误直接透传
func (l *ElementListLogic) ElementList(req *types.ElementListReq) (resp *types.ElementListResp, err error) {
	page, size := normalizePage(req.Page, req.Size)
	reply, err := l.svcCtx.StandardRpc.ListElements(l.ctx, &standardClient.ElementListReq{
		Beid:    req.Beid,
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.StdElement, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *elementToTypes(item))
	}
	return &types.ElementListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}

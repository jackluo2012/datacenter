// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package mdm

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/mdm/rpc/mdmClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type MdmModelListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 主数据模型列表
func NewMdmModelListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MdmModelListLogic {
	return &MdmModelListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 主数据模型列表: 页码默认值处理(page<=0 取 1, size<=0 取 10)后转发, 业务错误直接透传
func (l *MdmModelListLogic) MdmModelList(req *types.MdmModelListReq) (resp *types.MdmModelListResp, err error) {
	page, size := req.Page, req.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	reply, err := l.svcCtx.MdmRpc.ListModels(l.ctx, &mdmClient.ModelListReq{
		Beid:    req.Beid,
		Keyword: req.Keyword,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	list := make([]types.MdmModelInfo, 0, len(reply.GetList()))
	for _, item := range reply.GetList() {
		list = append(list, *mdmModelToTypes(item))
	}
	return &types.MdmModelListResp{
		List:  list,
		Total: reply.GetTotal(),
	}, nil
}

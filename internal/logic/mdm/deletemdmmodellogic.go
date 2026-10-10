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

type DeleteMdmModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除主数据模型
func NewDeleteMdmModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMdmModelLogic {
	return &DeleteMdmModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除主数据模型: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteModel 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteMdmModelLogic) DeleteMdmModel(req *types.MdmIdReq) error {
	_, err := l.svcCtx.MdmRpc.DeleteModel(l.ctx, &mdmClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}

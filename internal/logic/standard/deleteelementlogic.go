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

type DeleteElementLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除数据元
func NewDeleteElementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteElementLogic {
	return &DeleteElementLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除数据元: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteElement 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteElementLogic) DeleteElement(req *types.StdIdReq) error {
	_, err := l.svcCtx.StandardRpc.DeleteElement(l.ctx, &standardClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}

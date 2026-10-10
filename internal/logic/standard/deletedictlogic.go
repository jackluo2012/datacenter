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

type DeleteDictLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除字典
func NewDeleteDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDictLogic {
	return &DeleteDictLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除字典: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteDict 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteDictLogic) DeleteDict(req *types.StdIdReq) error {
	_, err := l.svcCtx.StandardRpc.DeleteDict(l.ctx, &standardClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}

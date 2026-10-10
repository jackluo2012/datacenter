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

type DeleteWordRootLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除词根
func NewDeleteWordRootLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteWordRootLogic {
	return &DeleteWordRootLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 删除词根: 网关请求逐字段映射为 rpc 请求后转发; pb DeleteWordRoot 虽返回 IdReq, 网关无返回体, 仅透传错误
func (l *DeleteWordRootLogic) DeleteWordRoot(req *types.StdIdReq) error {
	_, err := l.svcCtx.StandardRpc.DeleteWordRoot(l.ctx, &standardClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	return err
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasvc

import (
	"context"

	"datacenter/datasvc/rpc/dataSvcClient"
	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除数据API
func NewDeleteApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteApiLogic {
	return &DeleteApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteApi 转发 datasvc rpc 删除数据 API
func (l *DeleteApiLogic) DeleteApi(req *types.DsIdReq) error {
	// 逐字段映射为 rpc 请求
	_, err := l.svcCtx.DataSvcRpc.DeleteApi(l.ctx, &dataSvcClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	// rpc 业务错误原样透传
	return err
}

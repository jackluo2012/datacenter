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

type PublishApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发布API
func NewPublishApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishApiLogic {
	return &PublishApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// PublishApi 转发 datasvc rpc 发布 API
func (l *PublishApiLogic) PublishApi(req *types.DsIdReq) error {
	// 逐字段映射为 rpc 请求
	_, err := l.svcCtx.DataSvcRpc.PublishApi(l.ctx, &dataSvcClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	// rpc 业务错误原样透传
	return err
}

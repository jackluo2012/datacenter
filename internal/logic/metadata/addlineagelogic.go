// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package metadata

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/metadata/rpc/metadataClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddLineageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 登记血缘
func NewAddLineageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLineageLogic {
	return &AddLineageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 登记血缘: 网关请求逐字段映射为 rpc 请求后转发, 业务错误直接透传
func (l *AddLineageLogic) AddLineage(req *types.AddLineageReq) error {
	_, err := l.svcCtx.MetadataRpc.AddLineage(l.ctx, &metadataClient.AddLineageReq{
		FromId:    req.FromId,
		ToId:      req.ToId,
		PropsJson: req.PropsJson,
		Beid:      req.Beid,
	})
	return err
}

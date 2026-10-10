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

type StdModelInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标准模型详情
func NewStdModelInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StdModelInfoLogic {
	return &StdModelInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 标准模型详情: 网关请求逐字段映射为 rpc 请求, 转发后把模型转为网关类型返回
func (l *StdModelInfoLogic) StdModelInfo(req *types.StdIdReq) (resp *types.StdModel, err error) {
	reply, err := l.svcCtx.StandardRpc.GetModel(l.ctx, &standardClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return stdModelToTypes(reply.GetModel()), nil
}

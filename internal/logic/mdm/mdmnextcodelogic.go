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

type MdmNextCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 生成下一个业务编码
func NewMdmNextCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MdmNextCodeLogic {
	return &MdmNextCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 生成下一个业务编码: 网关请求逐字段映射为 rpc 请求后转发, 仅取号段回复中的业务编码返回
func (l *MdmNextCodeLogic) MdmNextCode(req *types.MdmIdReq) (resp *types.NextCodeResp, err error) {
	reply, err := l.svcCtx.MdmRpc.NextCode(l.ctx, &mdmClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.NextCodeResp{
		BizCode: reply.GetBizCode(),
	}, nil
}

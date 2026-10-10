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

type MdmRecordInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 主数据记录详情
func NewMdmRecordInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MdmRecordInfoLogic {
	return &MdmRecordInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 主数据记录详情: 网关请求逐字段映射为 rpc 请求后转发, 转发后把记录转为网关类型返回
func (l *MdmRecordInfoLogic) MdmRecordInfo(req *types.MdmIdReq) (resp *types.MdmRecordInfo, err error) {
	reply, err := l.svcCtx.MdmRpc.GetRecord(l.ctx, &mdmClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return mdmRecordToTypes(reply.GetRecord()), nil
}

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

type PublishMdmRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发布主数据记录
func NewPublishMdmRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishMdmRecordLogic {
	return &PublishMdmRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 发布主数据记录: 网关请求逐字段映射为 rpc 请求后转发(draft→published 由 rpc 侧处理),
// 转发后把记录转为网关类型返回, 业务错误直接透传
func (l *PublishMdmRecordLogic) PublishMdmRecord(req *types.MdmIdReq) (resp *types.MdmRecordInfo, err error) {
	reply, err := l.svcCtx.MdmRpc.PublishRecord(l.ctx, &mdmClient.IdReq{
		Id:   req.Id,
		Beid: req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return mdmRecordToTypes(reply.GetRecord()), nil
}

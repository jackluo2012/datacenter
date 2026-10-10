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

type DistributeMdmRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 分发主数据记录
func NewDistributeMdmRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DistributeMdmRecordLogic {
	return &DistributeMdmRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 分发主数据记录: 网关请求逐字段映射为 rpc 请求后转发(真实写入目标表由 rpc 侧处理),
// 转发后把分发结果转为网关类型返回, 业务错误直接透传
func (l *DistributeMdmRecordLogic) DistributeMdmRecord(req *types.DistributeMdmReq) (resp *types.DistributeMdmResp, err error) {
	reply, err := l.svcCtx.MdmRpc.DistributeRecord(l.ctx, &mdmClient.DistributeReq{
		RecordId:    req.RecordId,
		TargetDsId:  req.TargetDsId,
		TargetDb:    req.TargetDb,
		TargetTable: req.TargetTable,
		Beid:        req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.DistributeMdmResp{
		Distribution: *mdmDistributionToTypes(reply.GetDistribution()),
	}, nil
}

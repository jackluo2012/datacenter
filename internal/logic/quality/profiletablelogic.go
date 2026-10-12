// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package quality

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/quality/rpc/qualityClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileTableLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 表剖析（即时剖析并落健康分）
func NewProfileTableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileTableLogic {
	return &ProfileTableLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 表剖析: 网关请求逐字段映射为 rpc 请求转发, 剖析结果(rpc TableProfile → 网关
// TableProfileInfo, ran_at 对齐为 run_at)经共享映射转出; Profile 为 nil 时返回零值结构
func (l *ProfileTableLogic) ProfileTable(req *types.ProfileTableReq) (resp *types.ProfileResp, err error) {
	reply, err := l.svcCtx.QualityRpc.ProfileTable(l.ctx, &qualityClient.ProfileTableReq{
		DatasourceId: req.DatasourceId,
		Beid:         req.Beid,
		DbName:       req.DbName,
		TableName:    req.TableName,
		PkColumn:     req.PkColumn,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.ProfileResp{
		Profile: tableProfileToTypes(reply.GetProfile()),
	}, nil
}

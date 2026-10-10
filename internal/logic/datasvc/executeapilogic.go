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

type ExecuteApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 执行数据API（外部统一入口）
func NewExecuteApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteApiLogic {
	return &ExecuteApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ExecuteApi 转发 datasvc rpc 执行 API（rpc 侧完成鉴权→限流→缓存→执行→脱敏）
func (l *ExecuteApiLogic) ExecuteApi(req *types.ExecuteApiReq) (resp *types.ExecuteApiResp, err error) {
	// 逐字段映射为 rpc 请求；网关契约无 ip / beid 字段，
	// beid 传 0（proto 零值），ip 传空串，由 rpc 侧按 path 归属处理
	rpcResp, err := l.svcCtx.DataSvcRpc.ExecuteApi(l.ctx, &dataSvcClient.ExecuteReq{
		AppKey:     req.AppKey,
		Sign:       req.Sign,
		Ts:         req.Ts,
		Path:       req.Path,
		ParamsJson: req.ParamsJson,
	})
	if err != nil {
		// rpc 业务错误原样透传
		return nil, err
	}

	// 逐字段映射执行结果
	return &types.ExecuteApiResp{
		RowsJson: rpcResp.GetRowsJson(),
		CostMs:   rpcResp.GetCostMs(),
		CacheHit: rpcResp.GetCacheHit(),
		Affected: rpcResp.GetAffected(),
	}, nil
}

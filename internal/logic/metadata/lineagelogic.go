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

type LineageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 血缘图
func NewLineageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LineageLogic {
	return &LineageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 血缘图: direction 缺省 downstream, depth 缺省 2, 转发后转换节点与边, 业务错误直接透传
func (l *LineageLogic) Lineage(req *types.LineageReq) (resp *types.LineageResp, err error) {
	direction, depth := req.Direction, req.Depth
	if direction == "" {
		direction = "downstream"
	}
	if depth <= 0 {
		depth = 2
	}
	reply, err := l.svcCtx.MetadataRpc.GetLineage(l.ctx, &metadataClient.LineageReq{
		EntityId:  req.EntityId,
		Direction: direction,
		Depth:     depth,
		Beid:      req.Beid,
	})
	if err != nil {
		// 业务错误直接透传 rpc error
		return nil, err
	}
	return &types.LineageResp{
		Nodes: lineageNodesToTypes(reply.Nodes),
		Edges: lineageEdgesToTypes(reply.Edges),
	}, nil
}

package logic

import (
	"context"

	"datacenter/metadata/model"
	"datacenter/metadata/rpc/internal/svc"
	"datacenter/metadata/rpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLineageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLineageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLineageLogic {
	return &GetLineageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 血缘图（上下游 N 跳）
func (l *GetLineageLogic) GetLineage(in *metadata.LineageReq) (*metadata.LineageReply, error) {
	// 1. 参数校验
	if in.EntityId <= 0 {
		return nil, ErrInvalidLineageParams
	}
	direction := in.Direction
	if direction != model.DirectionUpstream && direction != model.DirectionDownstream {
		return nil, ErrInvalidDirection
	}
	depth := int(in.Depth)

	// 2. BFS 逐层展开（遍历逻辑抽为纯函数 traverseLineage，见 lineage.go 及其单测）
	next := func(id int64) ([]LineageEdgeData, error) {
		rels, err := l.svcCtx.RelationModel.FindNeighbors(l.ctx, model.RelTypeLineage, id, direction)
		if err != nil {
			return nil, err
		}
		edges := make([]LineageEdgeData, 0, len(rels))
		for _, r := range rels {
			edges = append(edges, LineageEdgeData{FromId: r.FromId, ToId: r.ToId, PropsJson: r.PropsJson})
		}
		return edges, nil
	}
	nodeIds, edgeData, err := traverseLineage(in.EntityId, direction, depth, next)
	if err != nil {
		return nil, err
	}

	// 3. 回填节点展示信息（fqn/name），不存在的实体（如跨租户）跳过
	entities, err := l.svcCtx.EntityModel.FindListByIds(l.ctx, in.Beid, nodeIds)
	if err != nil {
		return nil, err
	}
	entityMap := make(map[int64]*model.Entity, len(entities))
	for _, e := range entities {
		entityMap[e.Id] = e
	}
	nodes := make([]*metadata.LineageNode, 0, len(nodeIds))
	for _, id := range nodeIds {
		e, ok := entityMap[id]
		if !ok {
			continue
		}
		nodes = append(nodes, &metadata.LineageNode{Id: e.Id, Fqn: e.Fqn, Name: e.Name})
	}

	edges := make([]*metadata.LineageEdge, 0, len(edgeData))
	for _, ed := range edgeData {
		edges = append(edges, &metadata.LineageEdge{FromId: ed.FromId, ToId: ed.ToId, PropsJson: ed.PropsJson})
	}
	return &metadata.LineageReply{Nodes: nodes, Edges: edges}, nil
}

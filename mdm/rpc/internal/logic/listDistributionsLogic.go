package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDistributionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDistributionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDistributionsLogic {
	return &ListDistributionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListDistributions 分页查询分发记录（recordId 可选，按记录过滤）
func (l *ListDistributionsLogic) ListDistributions(in *mdm.DistListReq) (*mdm.DistListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.DistributionModel.FindList(l.ctx, in.Beid, in.RecordId, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	reply := &mdm.DistListReply{Total: total}
	for _, d := range list {
		reply.List = append(reply.List, toPbDistribution(d))
	}
	return reply, nil
}

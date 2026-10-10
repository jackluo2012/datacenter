package logic

import (
	"context"

	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAppliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAppliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAppliesLogic {
	return &ListAppliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 申请列表（分页 + 可选状态过滤）
func (l *ListAppliesLogic) ListApplies(in *catalog.ApplyListReq) (*catalog.ApplyListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	list, total, err := l.svcCtx.ApplyModel.ListByBeid(in.Beid, in.Status, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	// pb 契约中 ApplyListReply 无 total 字段，总数记录日志供排查
	l.Debugf("ListApplies beid=%d status=%s total=%d", in.Beid, in.Status, total)
	reply := make([]*catalog.AssetApply, 0, len(list))
	for _, a := range list {
		reply = append(reply, applyToPb(a))
	}
	return &catalog.ApplyListReply{List: reply}, nil
}

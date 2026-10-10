package logic

import (
	"context"
	"errors"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGroupLogic {
	return &DeleteGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除人群（仅删圈群定义行；人群物化表留档不删，历史人群快照可追溯，
// 由运维侧按 materialized_table 另行清理）
func (l *DeleteGroupLogic) DeleteGroup(in *tag.IdReq) (*tag.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	// 先确认存在（带租户隔离），避免误删
	group, err := model.NewGroupModel(l.svcCtx.Conn).FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	if err := model.NewGroupModel(l.svcCtx.Conn).Delete(l.ctx, group.Id, group.Beid); err != nil {
		l.Errorf("删除人群失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	return &tag.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

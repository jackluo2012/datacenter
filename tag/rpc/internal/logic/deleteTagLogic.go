package logic

import (
	"context"
	"errors"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagLogic {
	return &DeleteTagLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除标签定义（级联删除其全部标签值，事务内先删值后删定义）
func (l *DeleteTagLogic) DeleteTag(in *tag.IdReq) (*tag.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	// 先确认存在（带租户隔离），避免误删
	def, err := l.svcCtx.TagDefModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.TagDefModel.Delete(l.ctx, def.Id, def.Beid); err != nil {
		l.Errorf("删除标签定义失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	return &tag.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

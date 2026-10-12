package logic

import (
	"context"
	"errors"
	"fmt"

	"datacenter/tag/model"
	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteObjectLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteObjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteObjectLogic {
	return &DeleteObjectLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除标签对象：存在性校验 → 对象下仍有标签定义拒删 → 仍被人群引用拒删 → 物理删除
// （对象行无冗余快照被别处引用，删前先清引用保证不留孤儿定义/人群）
func (l *DeleteObjectLogic) DeleteObject(in *tag.IdReq) (*tag.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, fmt.Errorf("%w: id 不能为空", ErrInvalidParams)
	}
	// 存在性校验（带租户隔离），避免误删
	if _, err := l.svcCtx.ObjectModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	// 引用校验一：对象下仍有标签定义时拒绝删除
	defCnt, err := l.svcCtx.TagDefModel.CountByObject(l.ctx, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	if defCnt > 0 {
		return nil, ErrObjectHasTagDefs
	}
	// 引用校验二：对象仍被人群圈选定义引用时拒绝删除（人群 model 与既有 logic 一致就地构造）
	groupCnt, err := model.NewGroupModel(l.svcCtx.Conn).CountByObject(l.ctx, in.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	if groupCnt > 0 {
		return nil, ErrObjectHasGroups
	}
	if err := l.svcCtx.ObjectModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		l.Errorf("删除标签对象失败, id: %d, beid: %d, err: %v", in.Id, in.Beid, err)
		return nil, err
	}
	return &tag.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

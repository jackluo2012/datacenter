package logic

import (
	"context"
	"errors"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDictLogic {
	return &DeleteDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除字典（级联删字典项：事务内先删项再删字典，失败整体回滚）
func (l *DeleteDictLogic) DeleteDict(in *standard.IdReq) (*standard.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	if _, err := l.svcCtx.DictModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrDictNotFound
		}
		return nil, err
	}
	err := l.svcCtx.Conn.TransactCtx(l.ctx, func(txCtx context.Context, session sqlx.Session) error {
		if err := l.svcCtx.DictItemModel.DeleteByDictId(txCtx, session, in.Id); err != nil {
			return err
		}
		return l.svcCtx.DictModel.Delete(txCtx, session, in.Id, in.Beid)
	})
	if err != nil {
		return nil, err
	}
	return &standard.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

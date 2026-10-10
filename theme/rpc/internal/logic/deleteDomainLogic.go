package logic

import (
	"context"
	"errors"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDomainLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDomainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDomainLogic {
	return &DeleteDomainLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteDomain 删除主题域：域下存在主题时拒绝删除（防悬空引用）
func (l *DeleteDomainLogic) DeleteDomain(in *theme.IdReq) (*theme.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}
	//须存在
	if _, err := l.svcCtx.DomainModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrDomainNotFound
		}
		return nil, err
	}
	//有主题拒绝删除
	count, err := l.svcCtx.ThemeModel.CountByDomain(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrDomainInUse
	}
	if err := l.svcCtx.DomainModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	l.Infof("[DeleteDomain] 删除主题域 id=%d beid=%d", in.Id, in.Beid)
	return &theme.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

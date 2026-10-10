package logic

import (
	"context"
	"errors"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteThemeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteThemeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteThemeLogic {
	return &DeleteThemeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteTheme 删除主题：主题下存在数仓模型时拒绝删除（防悬空引用）
func (l *DeleteThemeLogic) DeleteTheme(in *theme.IdReq) (*theme.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}
	//须存在
	if _, err := l.svcCtx.ThemeModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrThemeNotFound
		}
		return nil, err
	}
	//有模型拒绝删除
	count, err := l.svcCtx.ModelModel.CountByTheme(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrThemeInUse
	}
	if err := l.svcCtx.ThemeModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	l.Infof("[DeleteTheme] 删除主题 id=%d beid=%d", in.Id, in.Beid)
	return &theme.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

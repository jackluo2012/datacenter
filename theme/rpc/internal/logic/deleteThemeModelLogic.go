package logic

import (
	"context"
	"errors"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteThemeModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteThemeModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteThemeModelLogic {
	return &DeleteThemeModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteThemeModel 删除数仓模型（仅删建模元数据，不落目标库物理表——已在目标库
// 建出的表由数据运维流程另行处置）
func (l *DeleteThemeModelLogic) DeleteThemeModel(in *theme.IdReq) (*theme.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}
	//须存在
	if _, err := l.svcCtx.ModelModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.ModelModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	l.Infof("[DeleteThemeModel] 删除模型 id=%d beid=%d", in.Id, in.Beid)
	return &theme.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

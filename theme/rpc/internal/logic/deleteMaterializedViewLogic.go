package logic

import (
	"context"
	"errors"

	"datacenter/theme/model"
	"datacenter/theme/rpc/internal/svc"
	"datacenter/theme/rpc/theme"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMaterializedViewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMaterializedViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMaterializedViewLogic {
	return &DeleteMaterializedViewLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteMaterializedView 删除物化视图定义（仅删中台账本，已在目标库建出的视图
// 不联动 drop，由数据运维流程另行处置）
func (l *DeleteMaterializedViewLogic) DeleteMaterializedView(in *theme.IdReq) (*theme.IdReq, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}
	//须存在
	if _, err := l.svcCtx.ThemeMvModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrMvNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.ThemeMvModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	l.Infof("[DeleteMaterializedView] 删除物化视图 id=%d beid=%d", in.Id, in.Beid)
	return &theme.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

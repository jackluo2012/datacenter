package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteWordRootLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteWordRootLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteWordRootLogic {
	return &DeleteWordRootLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除词根（幂等：词根不存在时不报错）
func (l *DeleteWordRootLogic) DeleteWordRoot(in *standard.IdReq) (*standard.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	if err := l.svcCtx.WordRootModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return &standard.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

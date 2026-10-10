package logic

import (
	"context"
	"errors"

	"datacenter/security/model"
	"datacenter/security/rpc/internal/svc"
	"datacenter/security/rpc/security"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteGrantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteGrantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGrantLogic {
	return &DeleteGrantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除授权（按 id+beid 定位，不存在时报授权不存在）
func (l *DeleteGrantLogic) DeleteGrant(in *security.IdReq) (*security.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	if _, err := l.svcCtx.GrantModel.FindOne(l.ctx, in.Id, in.Beid); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrGrantNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.GrantModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return &security.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

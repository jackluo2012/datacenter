package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteElementLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteElementLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteElementLogic {
	return &DeleteElementLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除数据元（幂等：数据元不存在时不报错；
// 已被标准模型 elements_json 引用的数据元不做强约束，由下次保存/发布时存在性校验兜底）
func (l *DeleteElementLogic) DeleteElement(in *standard.IdReq) (*standard.IdReq, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	if err := l.svcCtx.ElementModel.Delete(l.ctx, in.Id, in.Beid); err != nil {
		return nil, err
	}
	return &standard.IdReq{Id: in.Id, Beid: in.Beid}, nil
}

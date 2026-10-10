package logic

import (
	"context"
	"errors"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictLogic {
	return &GetDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 字典详情（字典 + 全部字典项，项按 sort,id 升序）
func (l *GetDictLogic) GetDict(in *standard.IdReq) (*standard.DictReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if in.Id <= 0 {
		return nil, ErrIdRequired
	}
	dict, err := l.svcCtx.DictModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrDictNotFound
		}
		return nil, err
	}
	items, err := l.svcCtx.DictItemModel.FindByDictId(l.ctx, dict.Id)
	if err != nil {
		return nil, err
	}
	pbItems := make([]*standard.DictItem, 0, len(items))
	for _, item := range items {
		pbItems = append(pbItems, itemToPb(item))
	}
	return &standard.DictReply{Dict: dictToPb(dict), Items: pbItems}, nil
}

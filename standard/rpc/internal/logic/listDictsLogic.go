package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDictsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDictsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDictsLogic {
	return &ListDictsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 字典列表（分页 + keyword 对 dict_code/dict_name 模糊匹配；不含字典项）
func (l *ListDictsLogic) ListDicts(in *standard.DictListReq) (*standard.DictListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.DictModel.FindList(l.ctx, in.Beid, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	dicts := make([]*standard.Dict, 0, len(list))
	for _, d := range list {
		dicts = append(dicts, dictToPb(d))
	}
	return &standard.DictListReply{List: dicts, Total: total}, nil
}

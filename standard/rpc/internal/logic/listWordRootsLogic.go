package logic

import (
	"context"

	"datacenter/standard/model"
	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWordRootsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWordRootsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWordRootsLogic {
	return &ListWordRootsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 词根列表（分页 + keyword 对 name/cn_name 模糊匹配）
func (l *ListWordRootsLogic) ListWordRoots(in *standard.WordRootListReq) (*standard.WordRootListReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	list, total, err := l.svcCtx.WordRootModel.FindList(l.ctx, in.Beid, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	roots := make([]*standard.WordRoot, 0, len(list))
	for _, r := range list {
		roots = append(roots, rootToPb(r))
	}
	return &standard.WordRootListReply{List: roots, Total: total}, nil
}

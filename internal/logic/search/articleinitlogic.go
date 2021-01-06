package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/search/rpc/search"

	"github.com/tal-tech/go-zero/core/logx"
)

type ArticleInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) ArticleInitLogic {
	return ArticleInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleInitLogic) ArticleInit(req types.SearchReq) (*types.SearchResp, error) {
	_, err := l.svcCtx.SearchRpc.ArticleInit(l.ctx, &search.Request{})
	return &types.SearchResp{}, err
}

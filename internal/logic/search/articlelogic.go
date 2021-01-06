package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/search/rpc/searchclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type ArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) ArticleLogic {
	return ArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleLogic) Article(req searchclient.SearchReq) (*searchclient.ArticleResp, error) {
	return l.svcCtx.SearchRpc.ArticleSearch(l.ctx, &req)
}

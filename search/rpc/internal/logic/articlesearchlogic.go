package logic

import (
	"context"

	"datacenter/search/rpc/internal/svc"
	"datacenter/search/rpc/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleSearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleSearchLogic {
	return &ArticleSearchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArticleSearchLogic) ArticleSearch(in *search.SearchReq) (*search.ArticleResp, error) {

	list, err := l.svcCtx.ArticeEs.Search(in)
	if err != nil {
		return nil, err
	}

	return &search.ArticleResp{List: list}, nil
}

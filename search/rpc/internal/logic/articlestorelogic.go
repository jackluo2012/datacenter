package logic

import (
	"context"

	"datacenter/search/rpc/internal/svc"
	"datacenter/search/rpc/search"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleStoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleStoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleStoreLogic {
	return &ArticleStoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArticleStoreLogic) ArticleStore(in *search.ArticleReq) (*search.Response, error) {
	id, err := l.svcCtx.ArticeEs.Index(in)
	if err != nil {
		return nil, err
	}
	logx.Info(id)
	return &search.Response{}, nil
}

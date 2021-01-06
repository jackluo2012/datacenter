package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/search/rpc/searchclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type ArticleStoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleStoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) ArticleStoreLogic {
	return ArticleStoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleStoreLogic) ArticleStore(req searchclient.ArticleReq) (*types.ArticleReq, error) {
	_, err := l.svcCtx.SearchRpc.ArticleStore(l.ctx, &req)
	if err != nil {
		return nil, err
	}

	return &types.ArticleReq{}, nil
}

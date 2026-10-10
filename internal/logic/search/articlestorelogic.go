package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/search/rpc/searchclient"

	"github.com/zeromicro/go-zero/core/logx"
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

// pb message 内含 sync.Mutex，按值拷贝不安全（copylocks），统一改为指针传递（issue #18）
func (l *ArticleStoreLogic) ArticleStore(req *searchclient.ArticleReq) (*types.ArticleReq, error) {
	_, err := l.svcCtx.SearchRpc.ArticleStore(l.ctx, req)
	if err != nil {
		return nil, err
	}

	return &types.ArticleReq{}, nil
}

package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/search/rpc/searchclient"

	"github.com/zeromicro/go-zero/core/logx"
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

// pb message 内含 sync.Mutex，按值拷贝不安全（copylocks），统一改为指针传递（issue #18）
func (l *ArticleLogic) Article(req *searchclient.SearchReq) (*searchclient.ArticleResp, error) {
	return l.svcCtx.SearchRpc.ArticleSearch(l.ctx, req)
}

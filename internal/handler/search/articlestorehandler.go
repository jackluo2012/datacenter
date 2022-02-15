package handler

import (
	"net/http"

	logic "datacenter/internal/logic/search"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/search/rpc/searchclient"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ArticleStoreHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ArticleReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		sreq := searchclient.ArticleReq{
			ImageUrl:  req.ImageUrl,
			NewsId:    req.NewsId,
			NewsTitle: req.NewsTitle,
		}
		l := logic.NewArticleStoreLogic(r.Context(), ctx)
		resp, err := l.ArticleStore(sreq)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

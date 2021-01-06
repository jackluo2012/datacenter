package handler

import (
	"net/http"

	logic "datacenter/internal/logic/search"
	"datacenter/internal/svc"
	"datacenter/search/rpc/searchclient"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/rest/httpx"
)

func ArticleHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		offset, size := shared.ToLimitOffset(query.Get("page"), query.Get("size"))
		req := searchclient.SearchReq{
			Limit: &searchclient.LimitReq{
				Offset: offset,
				Size:   size,
			},
			Keyword: query.Get("keyword"),
		}
		l := logic.NewArticleLogic(r.Context(), ctx)
		resp, err := l.Article(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

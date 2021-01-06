package handler

import (
	"net/http"

	logic "datacenter/internal/logic/search"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/rest/httpx"
)

func ArticleInitHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchReq
		l := logic.NewArticleInitLogic(r.Context(), ctx)
		resp, err := l.ArticleInit(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

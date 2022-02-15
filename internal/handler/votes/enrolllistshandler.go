package handler

import (
	"net/http"

	logic "datacenter/internal/logic/votes"
	"datacenter/internal/svc"
	"datacenter/shared"
	"datacenter/votes/rpc/votesclient"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func EnrollListsHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		actid := shared.StrToInt64(query.Get("actid"))
		offset, size := shared.ToLimitOffset(query.Get("page"), query.Get("size"))

		if actid == 0 {
			httpx.Error(w, shared.ErrorNoRequiredParameters)
			return
		}
		actreq := votesclient.ActidReq{
			Actid:   actid,
			Keyword: query.Get("keyword"),
			Limit: &votesclient.LimitReq{
				Offset: offset,
				Size:   size,
			},
		}
		l := logic.NewEnrollListsLogic(r.Context(), ctx)
		resp, err := l.EnrollLists(actreq)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

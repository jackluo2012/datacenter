package handler

import (
	"net/http"

	logic "datacenter/internal/logic/votes"
	"datacenter/internal/svc"
	"datacenter/shared"
	"datacenter/votes/rpc/votes"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ActivityIcrViewHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		actid := shared.StrToInt64(query.Get("actid"))
		activitReq := votes.ActInfoReq{
			Actid: actid,
		}

		l := logic.NewActivityIcrViewLogic(r.Context(), ctx)
		resp, err := l.ActivityIcrView(activitReq)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

package handler

import (
	"net/http"

	logic "datacenter/internal/logic/votes"
	"datacenter/internal/svc"
	"datacenter/shared"
	"datacenter/votes/rpc/votes"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func EnrollInfoHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		aeid := shared.StrToInt64(query.Get("aeid"))
		actid := shared.StrToInt64(query.Get("actid"))

		eReq := &votes.EnrollInfoReq{
			Actid: actid,
			Aeid:  aeid,
		}

		l := logic.NewEnrollInfoLogic(r.Context(), ctx)
		resp, err := l.EnrollInfo(eReq)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

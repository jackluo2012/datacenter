package handler

import (
	"net/http"

	logic "datacenter/internal/logic/questions"
	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ActivitiesInfoHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		actid := shared.StrToInt64(query.Get("actid"))
		activitReq := questionsclient.ActivitiesReq{
			Actid: actid,
		}

		l := logic.NewActivitiesInfoLogic(r.Context(), ctx)
		resp, err := l.ActivitiesInfo(activitReq)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

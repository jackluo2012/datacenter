package handler

import (
	"net/http"

	logic "datacenter/internal/logic/questions"
	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/rest/httpx"
)

func ListsHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		actid := shared.StrToInt64(query.Get("actid"))
		activitReq := questionsclient.ActivitiesReq{
			Actid: actid,
		}

		l := logic.NewListsLogic(r.Context(), ctx)
		resp, err := l.Lists(activitReq)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

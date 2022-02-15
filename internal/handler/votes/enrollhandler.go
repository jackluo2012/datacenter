package handler

import (
	"net/http"

	logic "datacenter/internal/logic/votes"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func EnrollHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req types.EnrollReq

		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		l := logic.NewEnrollLogic(r.Context(), ctx)
		resp, err := l.Enroll(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

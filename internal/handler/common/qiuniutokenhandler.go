package handler

import (
	"net/http"

	logic "datacenter/internal/logic/common"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/rest/httpx"
)

func QiuniuTokenHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Beid
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewQiuniuTokenLogic(r.Context(), ctx)
		resp, err := l.QiuniuToken(req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

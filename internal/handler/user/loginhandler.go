package handler

import (
	"net/http"

	logic "datacenter/internal/logic/user"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		ptyid := shared.StrToInt64(r.Header.Get("ptyid"))
		beid := shared.StrToInt64(r.Header.Get("beid"))

		l := logic.NewLoginLogic(r.Context(), ctx)
		resp, err := l.Login(req, beid, ptyid)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

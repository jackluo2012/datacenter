package handler

import (
	"fmt"
	"net/http"

	logic "datacenter/internal/logic/user"
	"datacenter/internal/svc"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func Code2SessionHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		code := query.Get("code")

		l := logic.NewCode2SessionLogic(r.Context(), ctx)

		fmt.Println("请求的code 是", code)
		resp, err := l.Code2Session(1, 1, code)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

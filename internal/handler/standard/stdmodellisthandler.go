// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package standard

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"datacenter/internal/logic/standard"
	"datacenter/internal/svc"
	"datacenter/internal/types"
)

func StdModelListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StdModelListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := standard.NewStdModelListLogic(r.Context(), svcCtx)
		resp, err := l.StdModelList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

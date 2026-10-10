// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package standard

import (
	"net/http"

	"datacenter/internal/logic/standard"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteWordRootHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StdIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := standard.NewDeleteWordRootLogic(r.Context(), svcCtx)
		err := l.DeleteWordRoot(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

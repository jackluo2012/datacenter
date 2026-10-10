// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package mdm

import (
	"net/http"

	"datacenter/internal/logic/mdm"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func MdmNextCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MdmIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := mdm.NewMdmNextCodeLogic(r.Context(), svcCtx)
		resp, err := l.MdmNextCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

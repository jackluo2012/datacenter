// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package mdm

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"datacenter/internal/logic/mdm"
	"datacenter/internal/svc"
	"datacenter/internal/types"
)

func DeleteMdmModelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MdmIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := mdm.NewDeleteMdmModelLogic(r.Context(), svcCtx)
		err := l.DeleteMdmModel(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

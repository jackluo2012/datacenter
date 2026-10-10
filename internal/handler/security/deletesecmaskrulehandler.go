// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package security

import (
	"net/http"

	"datacenter/internal/logic/security"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func DeleteSecMaskRuleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SecIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := security.NewDeleteSecMaskRuleLogic(r.Context(), svcCtx)
		err := l.DeleteSecMaskRule(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

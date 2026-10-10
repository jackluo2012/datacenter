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

func SaveSecMaskRuleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SaveSecMaskRuleReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := security.NewSaveSecMaskRuleLogic(r.Context(), svcCtx)
		resp, err := l.SaveSecMaskRule(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

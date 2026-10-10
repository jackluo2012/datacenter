// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package theme

import (
	"net/http"

	"datacenter/internal/logic/theme"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SaveThemeMvHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SaveThemeMvReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := theme.NewSaveThemeMvLogic(r.Context(), svcCtx)
		resp, err := l.SaveThemeMv(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

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

func ThemeModelListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ThemeModelListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := theme.NewThemeModelListLogic(r.Context(), svcCtx)
		resp, err := l.ThemeModelList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

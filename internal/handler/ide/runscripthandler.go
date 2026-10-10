// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ide

import (
	"net/http"

	"datacenter/internal/logic/ide"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 执行脚本（仅 SELECT，行数保护）
func RunScriptHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RunScriptReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ide.NewRunScriptLogic(r.Context(), svcCtx)
		resp, err := l.RunScript(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

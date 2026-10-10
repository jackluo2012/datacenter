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

// 删除脚本
func DeleteScriptHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ScriptIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ide.NewDeleteScriptLogic(r.Context(), svcCtx)
		err := l.DeleteScript(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

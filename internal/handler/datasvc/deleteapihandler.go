// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasvc

import (
	"net/http"

	"datacenter/internal/logic/datasvc"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除数据API
func DeleteApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DsIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := datasvc.NewDeleteApiLogic(r.Context(), svcCtx)
		err := l.DeleteApi(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

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

// 更新数据API
func UpdateApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateApiReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := datasvc.NewUpdateApiLogic(r.Context(), svcCtx)
		err := l.UpdateApi(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

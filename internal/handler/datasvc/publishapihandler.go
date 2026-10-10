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

// 发布API
func PublishApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DsIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := datasvc.NewPublishApiLogic(r.Context(), svcCtx)
		err := l.PublishApi(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

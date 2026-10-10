// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package metadata

import (
	"net/http"

	"datacenter/internal/logic/metadata"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 定版快照
func CreateVersionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MetaEntityIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := metadata.NewCreateVersionLogic(r.Context(), svcCtx)
		resp, err := l.CreateVersion(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

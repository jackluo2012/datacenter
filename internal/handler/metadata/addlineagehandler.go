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

// 登记血缘
func AddLineageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddLineageReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := metadata.NewAddLineageLogic(r.Context(), svcCtx)
		err := l.AddLineage(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

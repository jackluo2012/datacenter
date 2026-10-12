// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package catalog

import (
	"net/http"

	"datacenter/internal/logic/catalog"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除资产
func DeleteAssetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CatalogIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := catalog.NewDeleteAssetLogic(r.Context(), svcCtx)
		err := l.DeleteAsset(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

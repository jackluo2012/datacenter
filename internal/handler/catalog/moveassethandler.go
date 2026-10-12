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

// 移动资产（切换/移出分类）
func MoveAssetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MoveAssetReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := catalog.NewMoveAssetLogic(r.Context(), svcCtx)
		resp, err := l.MoveAsset(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

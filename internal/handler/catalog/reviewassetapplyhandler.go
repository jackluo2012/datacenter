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

// 审批资产申请
func ReviewAssetApplyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReviewAssetApplyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := catalog.NewReviewAssetApplyLogic(r.Context(), svcCtx)
		resp, err := l.ReviewAssetApply(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

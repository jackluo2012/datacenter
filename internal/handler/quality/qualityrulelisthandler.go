// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package quality

import (
	"net/http"

	"datacenter/internal/logic/quality"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 质量规则列表
func QualityRuleListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QualityRuleListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := quality.NewQualityRuleListLogic(r.Context(), svcCtx)
		resp, err := l.QualityRuleList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

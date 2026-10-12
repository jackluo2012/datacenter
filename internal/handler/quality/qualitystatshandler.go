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

// 质量统计大盘
func QualityStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StatsQReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := quality.NewQualityStatsLogic(r.Context(), svcCtx)
		resp, err := l.QualityStats(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

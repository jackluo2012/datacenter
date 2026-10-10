// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package metric

import (
	"net/http"

	"datacenter/internal/logic/metric"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除指标
func DeleteMetricHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MetricIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := metric.NewDeleteMetricLogic(r.Context(), svcCtx)
		err := l.DeleteMetric(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

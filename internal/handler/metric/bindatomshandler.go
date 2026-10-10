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

// 派生指标绑定原子指标
func BindAtomsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BindAtomsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := metric.NewBindAtomsLogic(r.Context(), svcCtx)
		err := l.BindAtoms(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

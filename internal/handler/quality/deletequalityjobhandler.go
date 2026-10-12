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

// 删除质量任务
func DeleteQualityJobHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QualityIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := quality.NewDeleteQualityJobLogic(r.Context(), svcCtx)
		err := l.DeleteQualityJob(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

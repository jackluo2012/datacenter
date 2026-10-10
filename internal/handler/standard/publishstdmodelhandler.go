// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package standard

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"datacenter/internal/logic/standard"
	"datacenter/internal/svc"
	"datacenter/internal/types"
)

func PublishStdModelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublishStdModelReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := standard.NewPublishStdModelLogic(r.Context(), svcCtx)
		resp, err := l.PublishStdModel(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

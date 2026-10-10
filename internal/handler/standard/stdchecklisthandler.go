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

func StdCheckListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.StdCheckListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := standard.NewStdCheckListLogic(r.Context(), svcCtx)
		resp, err := l.StdCheckList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

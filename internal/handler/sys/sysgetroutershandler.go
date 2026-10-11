// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sys

import (
	"net/http"

	"datacenter/internal/logic/sys"
	"datacenter/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SysGetroutersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := sys.NewSysGetroutersLogic(r.Context(), svcCtx)
		resp, err := l.SysGetrouters()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

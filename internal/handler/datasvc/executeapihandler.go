// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasvc

import (
	"net/http"

	"datacenter/internal/logic/datasvc"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 执行数据API（外部统一入口）
func ExecuteApiHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExecuteApiReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := datasvc.NewExecuteApiLogic(r.Context(), svcCtx)
		resp, err := l.ExecuteApi(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

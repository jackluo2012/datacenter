// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package datasource

import (
	"net/http"

	"datacenter/internal/logic/datasource"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 测试数据源连接
func DataSourceTestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DataSourceIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := datasource.NewDataSourceTestLogic(r.Context(), svcCtx)
		resp, err := l.DataSourceTest(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

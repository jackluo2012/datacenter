// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"net/http"

	"datacenter/internal/logic/tag"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除标签定义
func DeleteTagDefHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TagDefIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tag.NewDeleteTagDefLogic(r.Context(), svcCtx)
		err := l.DeleteTagDef(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

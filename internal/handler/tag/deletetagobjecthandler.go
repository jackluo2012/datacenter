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

// 删除标签对象（对象下仍有标签定义或人群时拒绝）
func DeleteTagObjectHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TagDefIdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tag.NewDeleteTagObjectLogic(r.Context(), svcCtx)
		err := l.DeleteTagObject(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}

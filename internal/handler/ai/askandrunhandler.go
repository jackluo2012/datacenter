// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package ai

import (
	"net/http"

	"datacenter/internal/logic/ai"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// AI 问数并执行（LLM 生成 SQL 后经 ide 安全闸自动执行）
func AskAndRunHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AskAndRunReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := ai.NewAskAndRunLogic(r.Context(), svcCtx)
		resp, err := l.AskAndRun(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

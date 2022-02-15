package handler

import (
	"net/http"

	logic "datacenter/internal/logic/votes"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func VoteHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VoteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		//获取用户的信息

		l := logic.NewVoteLogic(r.Context(), ctx)
		//获取用户的
		resp, err := l.Vote(req)
		if err != nil {
			httpx.Error(w, err)
			return
		} else {
			shared.OkJson(w, resp)
		}
	}
}

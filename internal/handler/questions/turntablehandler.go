package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	logic "datacenter/internal/logic/questions"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/questions/rpc/questionsclient"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func TurntableHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QuestionsAwardReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		uid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("uid"))).Int64()
		auid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("auid"))).Int64()
		ptyid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("ptyid"))).Int64()
		beid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("beid"))).Int64()

		in := questionsclient.TurnTableReq{
			ActivityId: req.ActivityId,
			Uid:        uid,
			Auid:       auid,
			AnswerId:   req.AnswerId,
			Beid:       beid,
			Ptyid:      ptyid,
		}

		l := logic.NewTurntableLogic(r.Context(), ctx)
		resp, err := l.Turntable(in)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

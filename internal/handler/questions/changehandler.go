package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	logic "datacenter/internal/logic/questions"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/questions/rpc/questions"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/rest/httpx"
)

func ChangeHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req types.AnswerReq

		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		uid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("uid"))).Int64()
		ptyid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("ptyid"))).Int64()
		beid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("beid"))).Int64()

		in := questions.QuestionsAnswerReq{
			Uid:        uid,
			Ptyid:      ptyid,
			Beid:       beid,
			ActivityId: req.ActivityId,
			Answers:    req.Answers,
			Score:      req.Score,
		}

		l := logic.NewChangeLogic(r.Context(), ctx)
		resp, err := l.Change(in)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

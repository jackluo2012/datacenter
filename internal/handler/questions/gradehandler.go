package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	logic "datacenter/internal/logic/questions"
	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"
	"datacenter/shared"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GradeHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		actid := shared.StrToInt64(query.Get("actid"))
		uid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("uid"))).Int64()
		auid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("auid"))).Int64()
		in := questionsclient.GradeReq{
			Actid: actid,
			Uid:   uid,
			Auid:  auid,
		}
		l := logic.NewGradeLogic(r.Context(), ctx)
		resp, err := l.Grade(in)
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

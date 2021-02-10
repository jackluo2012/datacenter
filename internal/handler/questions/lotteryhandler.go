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

	"github.com/tal-tech/go-zero/rest/httpx"
)

func LotteryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AwardConvertReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		uid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("uid"))).Int64()
		auid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("auid"))).Int64()
		ptyid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("ptyid"))).Int64()
		beid, _ := json.Number(fmt.Sprintf("%v", r.Context().Value("beid"))).Int64()
		l := logic.NewLotteryLogic(r.Context(), ctx)
		resp, err := l.Lottery(questionsclient.ConvertReq{
			Uid:       uid,
			Auid:      auid,
			LotteryId: req.LotteryId,
			Username:  req.UserName,
			Phone:     req.Phone,
			Beid:      beid,
			Ptyid:     ptyid,
		})
		if err != nil {
			httpx.Error(w, err)
		} else {
			shared.OkJson(w, resp)
		}
	}
}

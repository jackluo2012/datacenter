package handler

import (
	"net/http"

	logic "datacenter/internal/logic/common"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/rest/httpx"
)

func SnsInfoHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//		req.Beid.Beid = shared.StrToInt64(r.Header.Get("beid"))
		//		req.Ptyid = shared.StrToInt64(r.Header.Get("ptyid"))
		var req types.SnsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		//req.BackUrl = r.URL.Query().Get("backurl")
		if req.Ptyid == 0 || req.Beid.Beid == 0 {
			httpx.Error(w, shared.ErrorNoRequiredParameters)
			return
		}
		//站点的sns
		lsns := logic.NewSnsInfoLogic(r.Context(), ctx)
		//站点配置的
		lapp := logic.NewAppInfoLogic(r.Context(), ctx)
		app, err := lapp.AppInfo(req.Beid)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		resp, err := lsns.SnsInfo(req)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		app.Fullwebsite = app.Fullwebsite + req.BackUrl
		shared.OkJson(w, types.SnsResp{
			Beid:     req.Beid,
			Ptyid:    resp.Ptyid,
			Appid:    resp.Appid,
			Title:    resp.Title,
			LoginUrl: shared.GetWxLoginUrl(resp.Appid, app.Fullwebsite),
		})

	}
}

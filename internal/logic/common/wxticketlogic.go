package logic

import (
	"context"
	"time"

	"datacenter/common/rpc/common"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/chanxuehong/wechat/mp/core"
	"github.com/chanxuehong/wechat/mp/jssdk"
	"github.com/tal-tech/go-zero/core/logx"
)

type WxTicketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWxTicketLogic(ctx context.Context, svcCtx *svc.ServiceContext) WxTicketLogic {
	return WxTicketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

var (
	accessTokenServer core.AccessTokenServer
	wechatClient      *core.Client
	cardTicketServer  *jssdk.DefaultTicketServer
)

func (l *WxTicketLogic) WxTicket(req types.SnsReq) (*types.WxShareResp, error) {
	// todo: add your logic here and delete this line
	//= core.NewDefaultAccessTokenServer(wxAppId, wxAppSecret, nil)
	appconf, err := l.svcCtx.CommonRpc.GetAppConfig(l.ctx, &common.AppConfigReq{
		Beid:  req.Beid.Beid,
		Ptyid: req.Ptyid,
	})
	if err != nil {
		return nil, err
	}
	var ticketoken = ""
	//如果有
	l.svcCtx.Cache.Get("ticket_token", &ticketoken)
	if ticketoken == "" {
		//拿到微信操作的
		accessTokenServer = core.NewDefaultAccessTokenServer(appconf.Appid, appconf.Appsecret, nil)
		wechatClient = core.NewClient(accessTokenServer, nil)

		cardTicketServer = jssdk.NewDefaultTicketServer(wechatClient)

		ticketoken, err = cardTicketServer.Ticket()
		if err != nil {
			return nil, err
		}
		l.svcCtx.Cache.SetWithExpire("ticket_token", ticketoken, time.Duration(3600)*time.Second)
	}

	// jsapiTicket := "sM4AOVdWfPE4DxkXGEs8VMCPGGVi4C3VM0P37wVUCFvkVAy_90u5h9nbSlYy3-Sl-HhTdfl2fzFy1AOcHKP7qg"
	// nonceStr := "Wm3WZYTPz0wzccnW"
	// timestamp := "1414587457"
	// url := "http://mp.weixin.qq.com?params=value#xxxx"

	// wantSignature := "0f9de62fce790f9a083d5c99e95740ceb90c27ed"

	// haveSignature := WXConfigSign(jsapiTicket, nonceStr, timestamp, url)

	wxticket := &types.WxShareResp{
		Appid:     appconf.Appid,
		Noncestr:  shared.Noncestr,
		Timestamp: time.Now().Unix(),
	}
	wxticket.Signature = jssdk.WXConfigSign(ticketoken, wxticket.Noncestr, shared.Int64ToStr(wxticket.Timestamp), req.BackUrl)

	return wxticket, nil
}

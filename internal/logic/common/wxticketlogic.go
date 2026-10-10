// TODO(回归): 需真实公众号凭据验证 OAuth 与 JS-SDK 签名
//
// 本文件已从 chanxuehong/wechat（已停更）迁移至 silenceper/wechat/v2。
package logic

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"time"

	"datacenter/common/rpc/common"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/zeromicro/go-zero/core/logx"
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

// wxOfficialAccount 按租户的 appid/secret 动态构建公众号实例。
// 本系统为多租户模式：每次调用按 beid 查库得到不同的 appid/secret，
// 而 silenceper/wechat 的 Wechat 实例是固定配置，故每次调用按需新建实例，
// access_token/jsapi_ticket 由实例内部的内存 cache 承接；外层再用 svcCtx.Cache
// 缓存 ticket（见 WxTicket），管理面低频调用可接受，后续如调用量上升可增加实例池。
func wxOfficialAccount(appid, secret string) *officialaccount.OfficialAccount {
	wc := wechat.NewWechat()
	return wc.GetOfficialAccount(&offConfig.Config{
		AppID:     appid,
		AppSecret: secret,
		Cache:     cache.NewMemory(),
	})
}

// wxConfigSign 计算微信 JS-SDK wx.config 所需签名，官方算法为固定顺序拼接后取 sha1：
// signature = sha1("jsapi_ticket=...&noncestr=...&timestamp=...&url=...")
// 注意：
//  1. url 不包含 '#' 及其后面部分（与原 chanxuehong 实现、官方文档规则一致）；
//  2. 不能改用 silenceper 的 util.Signature——它会先对参数做字典序排序，
//     只适用于支付等排序签名场景，不符合 jsapi 签名的固定顺序要求。
func wxConfigSign(jsapiTicket, nonceStr, timestamp, url string) string {
	if i := strings.IndexByte(url, '#'); i >= 0 {
		url = url[:i]
	}
	sum := sha1.Sum([]byte("jsapi_ticket=" + jsapiTicket +
		"&noncestr=" + nonceStr +
		"&timestamp=" + timestamp +
		"&url=" + url))
	return hex.EncodeToString(sum[:])
}

func (l *WxTicketLogic) WxTicket(req types.SnsReq) (*types.WxShareResp, error) {
	appconf, err := l.svcCtx.CommonRpc.GetAppConfig(l.ctx, &common.AppConfigReq{
		Beid:  req.Beid.Beid,
		Ptyid: req.Ptyid,
	})
	if err != nil {
		return nil, err
	}

	// jsapi_ticket 微信侧有效期 7200 秒，沿用原有逻辑在服务侧缓存 1 小时；
	// 缓存 key 携带 appid，避免多租户之间互相覆盖（原实现为固定 key，存在串租户问题）
	cacheKey := "ticket_token_" + appconf.Appid
	var ticketoken string
	l.svcCtx.Cache.Get(cacheKey, &ticketoken)
	if ticketoken == "" {
		// 缓存未命中：通过 silenceper 拉取 access_token 与 jsapi_ticket
		oa := wxOfficialAccount(appconf.Appid, appconf.Appsecret)
		accessToken, err := oa.GetAccessToken()
		if err != nil {
			return nil, err
		}
		ticketoken, err = oa.GetJs().GetTicket(accessToken)
		if err != nil {
			return nil, err
		}
		l.svcCtx.Cache.SetWithExpire(cacheKey, ticketoken, time.Hour)
	}

	wxticket := &types.WxShareResp{
		Appid:     appconf.Appid,
		Noncestr:  shared.Noncestr,
		Timestamp: time.Now().Unix(),
	}
	wxticket.Signature = wxConfigSign(ticketoken, wxticket.Noncestr, shared.Int64ToStr(wxticket.Timestamp), req.BackUrl)

	return wxticket, nil
}

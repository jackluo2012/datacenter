// TODO(回归): 需真实公众号凭据验证 OAuth 与 JS-SDK 签名
//
// 本文件已从 chanxuehong/wechat（已停更）迁移至 silenceper/wechat/v2。
package logic

import (
	"context"

	"datacenter/common/rpc/commonclient"
	"datacenter/user/model"
	"datacenter/user/rpc/internal/svc"
	"datacenter/user/rpc/user"

	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/zeromicro/go-zero/core/logx"
)

type SnsLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSnsLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SnsLoginLogic {
	return &SnsLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SnsLoginLogic) SnsLogin(in *user.AppConfigReq) (apr *user.AppUserResp, err error) {
	//先取配置文件信息
	reply, err := l.svcCtx.CommonRpc.GetAppConfig(l.ctx, &commonclient.AppConfigReq{Beid: in.Beid, Ptyid: in.Ptyid})
	if err != nil {
		return
	}

	//多租户动态 appid/secret 模式：每次调用按 beid 查库得到不同的 appid/secret，
	//而 silenceper/wechat 的 Wechat 实例是固定配置，故每次按需新建实例（网页授权
	//access_token 由微信按 code 下发，无需实例级缓存），登录为低频调用可接受，
	//后续如调用量上升可增加实例池
	wc := wechat.NewWechat()
	oa := wc.GetOfficialAccount(&offConfig.Config{
		AppID:     reply.Appid,
		AppSecret: reply.Appsecret,
		Cache:     cache.NewMemory(),
	})

	//调用微信第三方登陆：通过网页授权 code 换取用户 access_token（区别于接口调用凭证）
	token, err := oa.GetOauth().GetUserAccessToken(in.Code)
	if err != nil {
		return
	}

	// 检查 用户是否已存在
	appuser, err := l.svcCtx.AppUserModel.FindOneByOpenid(in.Beid, in.Ptyid, token.OpenID)

	if err != nil && err != model.ErrNotFound {
		return
	}
	if err == model.ErrNotFound {
		//scope=snsapi_userinfo 时可拉取用户资料；lang 传空由 SDK 默认使用 zh_CN
		userinfo, errox := oa.GetOauth().GetUserInfo(token.AccessToken, token.OpenID, "")
		if errox != nil {
			//原实现此处返回 model.ErrNotFound 会掩盖真实错误，改为透出微信侧错误
			return nil, errox
		}
		appuser = &model.AppUser{
			Ptyid:    in.Ptyid,
			Beid:     in.Beid,
			Openid:   userinfo.OpenID,
			Nickname: userinfo.Nickname,
			UnionId:  userinfo.Unionid,
			Avator:   userinfo.HeadImgURL,
			Province: userinfo.Province,
			Sex:      int64(userinfo.Sex),
			City:     userinfo.City,
			Country:  userinfo.Country, //原实现误填 userinfo.City，此处修正
		}
		res, errox := l.svcCtx.AppUserModel.Insert(*appuser)
		if errox != nil {
			return nil, errox
		}
		appuser.Auid, errox = res.LastInsertId()
		if errox != nil {
			return nil, errox
		}
		err = nil
	}
	return &user.AppUserResp{
		Ptyid:    appuser.Ptyid,
		Beid:     appuser.Beid,
		Nickname: appuser.Nickname,
		Openid:   appuser.Openid,
		Auid:     appuser.Auid,
		Avator:   appuser.Avator,
		Province: appuser.Province,
		Sex:      appuser.Sex,
		City:     appuser.City,
		Country:  appuser.Country, //原实现误填 appuser.City，与入库侧一并修正
	}, nil
}

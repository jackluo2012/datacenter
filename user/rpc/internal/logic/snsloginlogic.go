package logic

import (
	"context"

	"datacenter/common/rpc/commonclient"
	"datacenter/user/model"
	"datacenter/user/rpc/internal/svc"
	"datacenter/user/rpc/user"

	mpoauth2 "github.com/chanxuehong/wechat/mp/oauth2"
	"github.com/chanxuehong/wechat/oauth2"
	"github.com/tal-tech/go-zero/core/logx"
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
	//调用微信第三方登陆
	//mpoauth2.GetSessionWithClient()
	oauth2Endpoint := mpoauth2.NewEndpoint(reply.Appid, reply.Appsecret)
	oauth2Client := oauth2.Client{
		Endpoint: oauth2Endpoint,
	}

	token, err := oauth2Client.ExchangeToken(in.Code)
	if err != nil {
		return
	}

	// 检查 用户是否已存在
	appuser, err := l.svcCtx.AppUserModel.FindOneByOpenid(in.Beid, in.Ptyid, token.OpenId)

	if err != nil && err != model.ErrNotFound {
		//		logx.Info("appuser==xxxxx=====xxxxxx", appuser, err)
		return
	}
	if err == model.ErrNotFound {
		userinfo, errox := mpoauth2.GetUserInfo(token.AccessToken, token.OpenId, "", nil)
		if errox != nil {
			// 处理一般错误信息
			return
		}
		appuser = &model.AppUser{
			Ptyid:    in.Ptyid,
			Beid:     in.Beid,
			Openid:   userinfo.OpenId,
			Nickname: userinfo.Nickname,
			UnionId:  userinfo.UnionId,
			Avator:   userinfo.HeadImageURL,
			Province: userinfo.Province,
			Sex:      int64(userinfo.Sex),
			City:     userinfo.City,
			Country:  userinfo.City,
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
		Country:  appuser.City,
	}, nil
}

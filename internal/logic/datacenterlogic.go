package logic

import (
	"context"
	"time"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/dgrijalva/jwt-go"
	"github.com/zeromicro/go-zero/core/logx"
)

type DatacenterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDatacenterLogic(ctx context.Context, svcCtx *svc.ServiceContext) DatacenterLogic {
	return DatacenterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DatacenterLogic) Datacenter(req types.Request) (*types.Response, error) {
	// todo: add your logic here and delete this line
	// 在这里加入业务逻辑，我们用打印日志来代表业务逻辑
	l.Infof("name: %v", req.Name)

	return &types.Response{Message: "Hello Jack"}, nil
}

func (l *DatacenterLogic) GetJwtToken(ap types.AppUser) (types.JwtToken, error) {
	iat := time.Now().Unix()
	accessExpire := l.svcCtx.Config.Auth.AccessExpire

	claims := make(jwt.MapClaims)
	claims["exp"] = iat + l.svcCtx.Config.Auth.AccessExpire
	claims["iat"] = iat
	claims["uid"] = ap.Uid
	claims["auid"] = ap.Auid
	claims["beid"] = ap.Beid
	claims["ptyid"] = ap.Ptyid
	claims["openid"] = ap.Openid
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	jwtstr, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		return types.JwtToken{}, err
	}
	return types.JwtToken{
		AccessToken:  jwtstr,
		AccessExpire: iat + accessExpire,
		RefreshAfter: iat + accessExpire/2,
	}, nil
}

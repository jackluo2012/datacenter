package logic

import (
	"context"

	"datacenter/common/rpc/commonclient"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/zeromicro/go-zero/core/logx"
)

type QiuniuTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQiuniuTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) QiuniuTokenLogic {
	return QiuniuTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QiuniuTokenLogic) QiuniuToken(req types.Beid) (*types.Token, error) {
	repl, err := l.svcCtx.CommonRpc.GetAppConfig(l.ctx, &commonclient.AppConfigReq{
		Beid:  req.Beid,
		Ptyid: shared.QiuniuPtyId,
	})

	if err != nil {
		return nil, err
	}
	//请求七牛
	// 简单上传凭证
	putPolicy := storage.PutPolicy{
		Scope: "tanzi-datacenter",
	}
	mac := auth.New(repl.Appid, repl.Appsecret)

	upToken := putPolicy.UploadToken(mac)

	return &types.Token{
		Token: upToken,
	}, nil
}

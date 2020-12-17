package logic

import (
	"context"

	"datacenter/common/rpc/common"
	"datacenter/common/rpc/internal/svc"

	"github.com/tal-tech/go-zero/core/logx"
)

type GetAppConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAppConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppConfigLogic {
	return &GetAppConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAppConfigLogic) GetAppConfig(in *common.AppConfigReq) (*common.AppConfigResp, error) {
	appConfig, err := l.svcCtx.AppConfigModel.FindOneByAppid(in.Beid, in.Ptyid)
	if err != nil {
		return nil, err
	}

	return &common.AppConfigResp{
		Id:        appConfig.Id,
		Beid:      appConfig.Beid,
		Ptyid:     appConfig.Ptyid,
		Appid:     appConfig.Appid,
		Appsecret: appConfig.Appsecret,
		Title:     appConfig.Title,
	}, nil
}

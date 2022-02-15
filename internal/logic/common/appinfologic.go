package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"

	"datacenter/common/model"
	"datacenter/common/rpc/common"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) AppInfoLogic {
	return AppInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppInfoLogic) AppInfo(req types.Beid) (appconfig *common.BaseAppResp, err error) {

	//检查 缓存中是否有值
	err = l.svcCtx.Cache.Get(model.GetcacheBaseAppIdPrefix(req.Beid), appconfig)
	if err != nil && err == shared.ErrNotFound {
		appconfig, err = l.svcCtx.CommonRpc.GetBaseApp(l.ctx, &common.BaseAppReq{
			Beid: req.Beid,
		})
		if err != nil {
			return
		}
		err = l.svcCtx.Cache.Set(model.GetcacheBaseAppIdPrefix(req.Beid), appconfig)
	}

	return
}

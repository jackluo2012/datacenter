package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/tal-tech/go-zero/core/logx"
)

type ActivitiesInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActivitiesInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) ActivitiesInfoLogic {
	return ActivitiesInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

/**
 * 获取 活动信息
 */
func (l *ActivitiesInfoLogic) ActivitiesInfo(req questionsclient.ActivitiesReq) (*questionsclient.ActInfoResp, error) {

	return l.svcCtx.QuestionsRpc.GetActivitiesInfo(l.ctx, &req)

}

package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/zeromicro/go-zero/core/logx"
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

// pb message 内含 sync.Mutex，按值拷贝不安全（copylocks），统一改为指针传递（issue #18）
func (l *ActivitiesInfoLogic) ActivitiesInfo(req *questionsclient.ActivitiesReq) (*questionsclient.ActInfoResp, error) {

	return l.svcCtx.QuestionsRpc.GetActivitiesInfo(l.ctx, req)

}

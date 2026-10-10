package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AwardListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAwardListLogic(ctx context.Context, svcCtx *svc.ServiceContext) AwardListLogic {
	return AwardListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// pb message 内含 sync.Mutex，按值拷贝不安全（copylocks），统一改为指针传递（issue #18）
func (l *AwardListLogic) AwardList(in *questionsclient.ActivitiesReq) (*questionsclient.AwardListResp, error) {

	return l.svcCtx.QuestionsRpc.GetAwardList(l.ctx, in)

}

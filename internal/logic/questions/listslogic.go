package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questionsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListsLogic(ctx context.Context, svcCtx *svc.ServiceContext) ListsLogic {
	return ListsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// pb message 内含 sync.Mutex，按值拷贝不安全（copylocks），统一改为指针传递（issue #18）
func (l *ListsLogic) Lists(in *questionsclient.ActivitiesReq) (*questionsclient.QuestionsListResp, error) {
	return l.svcCtx.QuestionsRpc.GetQuestionsList(l.ctx, in)

}

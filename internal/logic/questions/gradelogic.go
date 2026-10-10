package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/questions/rpc/questions"
	"datacenter/questions/rpc/questionsclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GradeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGradeLogic(ctx context.Context, svcCtx *svc.ServiceContext) GradeLogic {
	return GradeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// pb message 内含 sync.Mutex，按值拷贝不安全（copylocks），统一改为指针传递（issue #18）
func (l *GradeLogic) Grade(in *questionsclient.GradeReq) (*questions.QuestionsAnswerResp, error) {
	// todo: add your logic here and delete this line
	return l.svcCtx.QuestionsRpc.GetQuestionsGrade(l.ctx, in)

}

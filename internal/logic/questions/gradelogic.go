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

func (l *GradeLogic) Grade(in questionsclient.GradeReq) (*questions.QuestionsAnswerResp, error) {
	// todo: add your logic here and delete this line
	return l.svcCtx.QuestionsRpc.GetQuestionsGrade(l.ctx, &in)

}

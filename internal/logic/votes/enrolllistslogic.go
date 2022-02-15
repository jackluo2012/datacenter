package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/votes/rpc/votes"
	"datacenter/votes/rpc/votesclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnrollListsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEnrollListsLogic(ctx context.Context, svcCtx *svc.ServiceContext) EnrollListsLogic {
	return EnrollListsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EnrollListsLogic) EnrollLists(req votesclient.ActidReq) ([]*votes.EnrollResp, error) {

	// todo: add your logic here and delete this line

	repl, err := l.svcCtx.VotesRpc.GetEnrollList(l.ctx, &req)
	if err != nil {
		return nil, err
	}
	logx.Info("GetEnrollList===", repl)

	return repl.Data, nil
}

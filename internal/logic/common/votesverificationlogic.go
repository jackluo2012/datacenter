package logic

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"

	"github.com/tal-tech/go-zero/core/logx"
)

type VotesVerificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVotesVerificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) VotesVerificationLogic {
	return VotesVerificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VotesVerificationLogic) VotesVerification(req types.SnsReq) (*types.SnsResp, error) {
	// todo: add your logic here and delete this line

	return &types.SnsResp{}, nil
}

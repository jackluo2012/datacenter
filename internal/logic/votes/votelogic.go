package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/shared"
	"datacenter/votes/rpc/votesclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type VoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) VoteLogic {
	return VoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VoteLogic) Vote(req types.VoteReq) (*votesclient.VotesResp, error) {
	auid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("auid"))).Int64()
	uid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("uid"))).Int64()
	ptyid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("ptyid"))).Int64()
	beid, _ := json.Number(fmt.Sprintf("%v", l.ctx.Value("beid"))).Int64()
	if auid == 0 && uid == 0 {
		return nil, shared.ErrorUserNotFound
	}
	return l.svcCtx.VotesRpc.Votes(l.ctx, &votesclient.VotesReq{
		Aeid:  req.Aeid,
		Actid: req.Actid.Actid,
		Auid:  auid,
		Uid:   uid,
		Ptyid: ptyid,
		Beid:  beid,
		Ip:    shared.GetLocalIP(),
	})
}

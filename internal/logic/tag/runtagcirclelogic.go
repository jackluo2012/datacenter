// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package tag

import (
	"context"

	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/tag/rpc/tagClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunTagCircleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 执行圈选
func NewRunTagCircleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunTagCircleLogic {
	return &RunTagCircleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RunTagCircleLogic) RunTagCircle(req *types.RunTagCircleReq) (resp *types.RunTagCircleResp, err error) {
	reply, err := l.svcCtx.TagRpc.RunCircle(l.ctx, &tagClient.RunCircleReq{
		GroupId: req.GroupId,
		Beid:    req.Beid,
	})
	if err != nil {
		return nil, err
	}
	return &types.RunTagCircleResp{MemberCount: reply.GetMemberCount(), Msg: reply.GetMsg()}, nil
}

package logic

import (
	"context"

	"datacenter/tag/rpc/internal/svc"
	"datacenter/tag/rpc/tag"

	"github.com/zeromicro/go-zero/core/logx"
)

type RunCircleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRunCircleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RunCircleLogic {
	return &RunCircleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行圈选（跑圈选 SQL 物化人群到 dc_tag_group_member）
func (l *RunCircleLogic) RunCircle(in *tag.RunCircleReq) (*tag.RunCircleReply, error) {
	// todo: add your logic here and delete this line

	return &tag.RunCircleReply{}, nil
}

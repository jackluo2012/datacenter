package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type NextCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNextCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NextCodeLogic {
	return &NextCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 生成下一个业务编码（号段模式）
func (l *NextCodeLogic) NextCode(in *mdm.IdReq) (*mdm.SegmentReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.SegmentReply{}, nil
}

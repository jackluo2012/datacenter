package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckTableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckTableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckTableLogic {
	return &CheckTableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标准核对（表结构与标准模型比对，经数据源真实执行）
func (l *CheckTableLogic) CheckTable(in *standard.CheckReq) (*standard.CheckReply, error) {
	// todo: add your logic here and delete this line

	return &standard.CheckReply{}, nil
}

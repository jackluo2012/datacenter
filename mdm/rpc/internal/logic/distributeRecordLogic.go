package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type DistributeRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDistributeRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DistributeRecordLogic {
	return &DistributeRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分发到目标数据源（真实写入目标表）
func (l *DistributeRecordLogic) DistributeRecord(in *mdm.DistributeReq) (*mdm.DistributeReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.DistributeReply{}, nil
}

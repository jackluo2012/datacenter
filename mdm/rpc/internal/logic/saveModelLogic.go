package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveModelLogic {
	return &SaveModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 主数据模型
func (l *SaveModelLogic) SaveModel(in *mdm.SaveModelReq) (*mdm.ModelReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.ModelReply{}, nil
}

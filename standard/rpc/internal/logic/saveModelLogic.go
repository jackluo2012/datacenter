package logic

import (
	"context"

	"datacenter/standard/rpc/internal/svc"
	"datacenter/standard/rpc/standard"

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

// 标准模型
func (l *SaveModelLogic) SaveModel(in *standard.SaveModelReq) (*standard.ModelReply, error) {
	// todo: add your logic here and delete this line

	return &standard.ModelReply{}, nil
}

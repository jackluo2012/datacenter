package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRecordLogic {
	return &DeleteRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteRecordLogic) DeleteRecord(in *mdm.IdReq) (*mdm.IdReq, error) {
	// todo: add your logic here and delete this line

	return &mdm.IdReq{}, nil
}

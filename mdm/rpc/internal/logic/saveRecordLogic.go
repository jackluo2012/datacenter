package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveRecordLogic {
	return &SaveRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 主数据记录
func (l *SaveRecordLogic) SaveRecord(in *mdm.SaveRecordReq) (*mdm.RecordReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.RecordReply{}, nil
}

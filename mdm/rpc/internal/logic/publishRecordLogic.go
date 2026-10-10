package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishRecordLogic {
	return &PublishRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发布记录（draft→published）
func (l *PublishRecordLogic) PublishRecord(in *mdm.IdReq) (*mdm.RecordReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.RecordReply{}, nil
}

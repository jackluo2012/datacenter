package logic

import (
	"context"

	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRecordsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRecordsLogic {
	return &ListRecordsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRecordsLogic) ListRecords(in *mdm.RecordListReq) (*mdm.RecordListReply, error) {
	// todo: add your logic here and delete this line

	return &mdm.RecordListReply{}, nil
}

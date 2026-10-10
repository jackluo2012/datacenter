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

// ListRecords 分页查询主数据记录（modelId 可选，keyword 模糊匹配 biz_code）
func (l *ListRecordsLogic) ListRecords(in *mdm.RecordListReq) (*mdm.RecordListReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	list, total, err := l.svcCtx.RecordModel.FindList(l.ctx, in.Beid, in.ModelId, in.Keyword, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	reply := &mdm.RecordListReply{Total: total}
	for _, rec := range list {
		reply.List = append(reply.List, toPbRecord(rec))
	}
	return reply, nil
}

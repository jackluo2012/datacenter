package logic

import (
	"context"
	"errors"

	"datacenter/mdm/model"
	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRecordLogic {
	return &GetRecordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetRecord 按主键查询主数据记录
func (l *GetRecordLogic) GetRecord(in *mdm.IdReq) (*mdm.RecordReply, error) {
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}
	rec, err := l.svcCtx.RecordModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &mdm.RecordReply{Record: toPbRecord(rec)}, nil
}

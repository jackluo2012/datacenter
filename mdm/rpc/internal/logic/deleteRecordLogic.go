package logic

import (
	"context"
	"errors"

	"datacenter/mdm/model"
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

// DeleteRecord 删除主数据记录（dc_mdm_distribution 分发留痕保留，供审计回溯）
func (l *DeleteRecordLogic) DeleteRecord(in *mdm.IdReq) (*mdm.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}

	//记录须存在
	rec, err := l.svcCtx.RecordModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	if err := l.svcCtx.RecordModel.Delete(l.ctx, rec.Id, in.Beid); err != nil {
		return nil, err
	}
	l.Infof("[DeleteRecord] 记录 id=%d biz_code=%s beid=%d 已删除", rec.Id, rec.BizCode, in.Beid)
	return &mdm.IdReq{Id: rec.Id, Beid: in.Beid}, nil
}

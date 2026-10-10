package logic

import (
	"context"
	"errors"

	"datacenter/mdm/model"
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

// PublishRecord 发布记录：draft→published（版本不变）；
// 已 published 再发布视为重新发布（version+1）；archived 拒绝
func (l *PublishRecordLogic) PublishRecord(in *mdm.IdReq) (*mdm.RecordReply, error) {
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}

	//记录须存在
	rec, err := l.svcCtx.RecordModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	if rec.Status == RecordStatusArchived {
		return nil, ErrRecordArchived
	}

	//已发布再发布 = 重新发布，版本+1；草稿首发版本不变
	bumpVersion := rec.Status == RecordStatusPublished
	if err := l.svcCtx.RecordModel.UpdateStatus(l.ctx, rec.Id, in.Beid, RecordStatusPublished, bumpVersion); err != nil {
		return nil, err
	}
	rec.Status = RecordStatusPublished
	if bumpVersion {
		rec.Version++
	}
	l.Infof("[PublishRecord] 记录 id=%d biz_code=%s 已发布 version=%d republish=%v beid=%d",
		rec.Id, rec.BizCode, rec.Version, bumpVersion, in.Beid)
	return &mdm.RecordReply{Record: toPbRecord(rec)}, nil
}

package logic

import (
	"context"
	"errors"

	"datacenter/mdm/model"
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

// SaveRecord 新增/更新主数据记录：
//   - id=0 新建：data_json 按模型 fields_json 逐字段校验收敛（required/type/多余字段拒绝），
//     biz_code 为空时号段自动生成，非空时模型内查重，落库 status=draft、version=1；
//   - id>0 更新：version+1（编码可改，仍须查重），归档记录禁止更新
func (l *SaveRecordLogic) SaveRecord(in *mdm.SaveRecordReq) (*mdm.RecordReply, error) {
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}
	if in.ModelId <= 0 {
		return nil, ErrInvalidParams
	}

	//模型须存在
	mdl, err := l.svcCtx.ModelModel.FindOne(l.ctx, in.ModelId, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}

	//按模型字段定义校验收敛 data_json（required 缺失/类型收敛/多余字段拒绝）
	fields, err := ParseFieldsJson(mdl.FieldsJson)
	if err != nil {
		return nil, err
	}
	dataJson, err := ValidateDataJson(fields, in.DataJson)
	if err != nil {
		return nil, err
	}

	//id=0 新建，id>0 更新（version+1）
	if in.Id == 0 {
		bizCode := trimSpace(in.BizCode)
		if bizCode == "" {
			//自动生成：号段取号（prefix + 日期段 + 6 位流水）
			bizCode, _, err = genBizCode(l.ctx, l.svcCtx, mdl, in.Beid)
			if err != nil {
				return nil, err
			}
		} else {
			if len(bizCode) > bizCodeMaxLen {
				return nil, ErrInvalidParams
			}
			//显式编码查重（模型内唯一）
			if _, err := l.svcCtx.RecordModel.FindByBizCode(l.ctx, mdl.Id, in.Beid, bizCode); err == nil {
				return nil, ErrDupBizCode
			} else if !errors.Is(err, model.ErrNotFound) {
				return nil, err
			}
		}
		row := &model.RecordRow{
			ModelId:  mdl.Id,
			BizCode:  bizCode,
			DataJson: dataJson,
			Status:   RecordStatusDraft,
			Version:  1,
			Beid:     in.Beid,
		}
		id, err := l.svcCtx.RecordModel.Insert(l.ctx, row)
		if err != nil {
			if errors.Is(err, model.ErrDupBizCode) {
				return nil, ErrDupBizCode
			}
			return nil, err
		}
		row.Id = id
		l.Infof("[SaveRecord] 新建记录 id=%d model=%d biz_code=%s beid=%d", id, mdl.Id, bizCode, in.Beid)
		return &mdm.RecordReply{Record: toPbRecord(row)}, nil
	}

	//更新：记录须存在，归档后禁止更新
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

	//编码可改：非空且变化时按模型内唯一查重
	bizCode := rec.BizCode
	if newCode := trimSpace(in.BizCode); newCode != "" && newCode != rec.BizCode {
		if len(newCode) > bizCodeMaxLen {
			return nil, ErrInvalidParams
		}
		if dup, err := l.svcCtx.RecordModel.FindByBizCode(l.ctx, rec.ModelId, in.Beid, newCode); err == nil && dup.Id != rec.Id {
			return nil, ErrDupBizCode
		} else if err != nil && !errors.Is(err, model.ErrNotFound) {
			return nil, err
		}
		bizCode = newCode
	}

	//更新内容并自增版本
	newVer, err := l.svcCtx.RecordModel.Update(l.ctx, &model.RecordRow{
		Id:       rec.Id,
		BizCode:  bizCode,
		DataJson: dataJson,
		Version:  rec.Version,
		Beid:     in.Beid,
	})
	if err != nil {
		if errors.Is(err, model.ErrDupBizCode) {
			return nil, ErrDupBizCode
		}
		return nil, err
	}
	rec.BizCode = bizCode
	rec.DataJson = dataJson
	rec.Version = newVer
	l.Infof("[SaveRecord] 更新记录 id=%d model=%d biz_code=%s version=%d beid=%d", rec.Id, rec.ModelId, bizCode, newVer, in.Beid)
	return &mdm.RecordReply{Record: toPbRecord(rec)}, nil
}

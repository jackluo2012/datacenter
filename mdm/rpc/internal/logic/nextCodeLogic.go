package logic

import (
	"context"
	"errors"
	"time"

	"datacenter/mdm/model"
	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
)

type NextCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNextCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NextCodeLogic {
	return &NextCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 号段默认配置（proto 未暴露码段配置字段，首取自动初始化）：
// 前缀取模型编码大写、日期段精确到日、步长 1
const (
	defaultDatePattern = "20060102"
	defaultStep        = 1
)

// NextCode 号段模式生成下一个业务编码：
// 事务内 update current_value=current_value+step 后回读（InnoDB 行锁串行化并发取号），
// 编码 = prefix + date_pattern（Go 格式）+ 6 位补零流水
func (l *NextCodeLogic) NextCode(in *mdm.IdReq) (*mdm.SegmentReply, error) {
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}

	//模型须存在
	mdl, err := l.svcCtx.ModelModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}

	bizCode, current, err := genBizCode(l.ctx, l.svcCtx, mdl, in.Beid)
	if err != nil {
		return nil, err
	}
	l.Infof("[NextCode] model=%d(%s) current=%d code=%s", mdl.Id, mdl.ModelCode, current, bizCode)
	return &mdm.SegmentReply{BizCode: bizCode, CurrentValue: current}, nil
}

// genBizCode 码段取号并拼装业务编码（NextCode 与 SaveRecord 自动编码共用）：
// 码段不存在时按默认配置初始化（前缀=模型编码大写、日期段 20060102、步长 1），
// 再事务内行锁推进流水取号
func genBizCode(ctx context.Context, sc *svc.ServiceContext, mdl *model.MdmModelRow, beid int64) (string, int64, error) {
	if _, err := sc.SegmentModel.FindOrCreateByModel(ctx, mdl.Id, beid,
		upper(mdl.ModelCode), defaultDatePattern, defaultStep); err != nil {
		return "", 0, err
	}
	//号段取号：事务内 update 推进 + 回读（行锁串行化）
	row, err := sc.SegmentModel.NextValue(ctx, mdl.Id)
	if err != nil {
		return "", 0, err
	}
	if row.Step <= 0 {
		return "", 0, ErrSegmentStepInvalid
	}
	//编码 = prefix + 日期段 + 6 位补零流水
	return BuildCode(row.Prefix, row.DatePattern, time.Now(), row.CurrentValue), row.CurrentValue, nil
}

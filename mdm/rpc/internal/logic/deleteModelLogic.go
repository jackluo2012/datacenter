package logic

import (
	"context"
	"errors"

	"datacenter/mdm/model"
	"datacenter/mdm/rpc/internal/svc"
	"datacenter/mdm/rpc/mdm"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteModelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteModelLogic {
	return &DeleteModelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteModel 删除主数据模型：模型下存在主数据记录时拒绝；
// 删除时同一事务内级联清理编码码段（物理数据表由外部生命周期管理，此处不删）
func (l *DeleteModelLogic) DeleteModel(in *mdm.IdReq) (*mdm.IdReq, error) {
	if in.Id <= 0 {
		return nil, ErrInvalidParams
	}
	if in.Beid <= 0 {
		return nil, ErrBeidRequired
	}

	//模型须存在
	exist, err := l.svcCtx.ModelModel.FindOne(l.ctx, in.Id, in.Beid)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}

	//存在记录即拒绝删除
	count, err := l.svcCtx.RecordModel.CountByModel(l.ctx, exist.Id, in.Beid)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrModelInUse
	}

	//单事务：先删码段再删模型
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := l.svcCtx.SegmentModel.DeleteByModel(ctx, session, exist.Id, in.Beid); err != nil {
			return err
		}
		return l.svcCtx.ModelModel.Delete(ctx, exist.Id, in.Beid)
	})
	if err != nil {
		return nil, err
	}
	l.Infof("[DeleteModel] 模型 id=%d code=%s beid=%d 已删除（含码段）", exist.Id, exist.ModelCode, in.Beid)
	return &mdm.IdReq{Id: exist.Id, Beid: in.Beid}, nil
}

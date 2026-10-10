package logic

import (
	"context"

	"datacenter/catalog/model"
	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyAssetLogic {
	return &ApplyAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 申请使用资产（写入 pending 状态的申请单）
func (l *ApplyAssetLogic) ApplyAsset(in *catalog.ApplyAssetReq) (*catalog.ApplyReply, error) {
	// 参数校验
	if in.AssetId <= 0 || in.Beid <= 0 {
		return nil, ErrInvalidParam
	}
	// 校验资产存在（带平台隔离）
	if _, err := l.svcCtx.AssetModel.FindOne(in.AssetId, in.Beid); err != nil {
		if err == model.ErrNotFound {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	ret, err := l.svcCtx.ApplyModel.Insert(model.AssetApply{
		AssetId:   in.AssetId,
		Reason:    in.Reason,
		Applicant: in.Applicant,
		Status:    ApplyStatusPending,
		Beid:      in.Beid,
	})
	if err != nil {
		return nil, err
	}
	id, err := ret.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &catalog.ApplyReply{
		Apply: &catalog.AssetApply{
			Id:        id,
			AssetId:   in.AssetId,
			Reason:    in.Reason,
			Applicant: in.Applicant,
			Status:    ApplyStatusPending,
			Beid:      in.Beid,
		},
	}, nil
}

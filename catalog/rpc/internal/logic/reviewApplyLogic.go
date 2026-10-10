package logic

import (
	"context"

	"datacenter/catalog/model"
	"datacenter/catalog/rpc/catalog"
	"datacenter/catalog/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReviewApplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewApplyLogic {
	return &ReviewApplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审批：仅 pending 状态可审批，approve→approved / reject→rejected，记录审批人
func (l *ReviewApplyLogic) ReviewApply(in *catalog.ReviewApplyReq) (*catalog.ApplyReply, error) {
	// 参数校验
	if in.Id <= 0 || in.Beid <= 0 || in.Reviewer == "" {
		return nil, ErrInvalidParam
	}
	// 查询申请单（带平台隔离）
	apply, err := l.svcCtx.ApplyModel.FindOne(in.Id, in.Beid)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, ErrApplyNotFound
		}
		return nil, err
	}
	// 状态机：仅 pending 可审批
	nextStatus, err := nextApplyStatus(apply.Status, in.Approve)
	if err != nil {
		return nil, err
	}
	// 落库（条件更新防并发重复审批）
	affected, err := l.svcCtx.ApplyModel.UpdateStatus(in.Id, in.Beid, nextStatus, in.Reviewer)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		// 并发下已被他人审批
		return nil, ErrApplyAlreadyReviewed
	}
	apply.Status = nextStatus
	apply.Reviewer = in.Reviewer
	return &catalog.ApplyReply{Apply: applyToPb(apply)}, nil
}

// applyToPb model 转 pb
func applyToPb(a *model.AssetApply) *catalog.AssetApply {
	return &catalog.AssetApply{
		Id:        a.Id,
		AssetId:   a.AssetId,
		Reason:    a.Reason,
		Applicant: a.Applicant,
		Status:    a.Status,
		Reviewer:  a.Reviewer,
		Beid:      a.Beid,
		CreatedAt: a.CreatedAt.Unix(),
	}
}

package logic

import "errors"

// 申请状态取值
const (
	ApplyStatusPending  = "pending"  // 待审批
	ApplyStatusApproved = "approved" // 已通过
	ApplyStatusRejected = "rejected" // 已驳回
)

// 资产可见性取值
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

var (
	// ErrCategoryNotFound 分类不存在
	ErrCategoryNotFound = errors.New("分类不存在")
	// ErrAssetNotFound 资产不存在
	ErrAssetNotFound = errors.New("资产不存在")
	// ErrApplyNotFound 申请不存在
	ErrApplyNotFound = errors.New("申请不存在")
	// ErrApplyAlreadyReviewed 申请已审批（仅 pending 状态可审批）
	ErrApplyAlreadyReviewed = errors.New("该申请已审批，不能重复审批")
	// ErrInvalidParam 请求参数不合法
	ErrInvalidParam = errors.New("请求参数不合法")
)

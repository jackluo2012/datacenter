package logic

import "fmt"

// nextApplyStatus 审批状态机（纯函数，便于单元测试）。
// 规则：仅 pending 状态可审批；approve=true 通过（approved），否则驳回（rejected）；
// 已审批（approved/rejected）再审批返回 ErrApplyAlreadyReviewed。
func nextApplyStatus(current string, approve bool) (string, error) {
	switch current {
	case ApplyStatusPending:
		if approve {
			return ApplyStatusApproved, nil
		}
		return ApplyStatusRejected, nil
	case ApplyStatusApproved, ApplyStatusRejected:
		return "", ErrApplyAlreadyReviewed
	default:
		return "", fmt.Errorf("非法的申请状态: %s", current)
	}
}

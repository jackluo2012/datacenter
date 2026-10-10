package logic

import (
	"errors"
	"testing"
)

// 聚合方式白名单：6 个合法取值全部通过
func TestIsValidAggTypeWhitelist(t *testing.T) {
	for _, agg := range []string{"sum", "count", "avg", "max", "min", "distinct"} {
		if !IsValidAggType(agg) {
			t.Fatalf("IsValidAggType(%q) = false, want true", agg)
		}
	}
}

// 聚合方式白名单外的取值一律拒绝（含大小写变体、空白、空串）
func TestIsValidAggTypeRejectsInvalid(t *testing.T) {
	for _, agg := range []string{"", "SUM", "Sum", "AVG", "median", "top", " sum", "sum "} {
		if IsValidAggType(agg) {
			t.Fatalf("IsValidAggType(%q) = true, want false", agg)
		}
	}
}

// 指标类型校验：仅 atom/derive 合法，其余（空串/大小写变体/近似词）拒绝
func TestIsValidMetricType(t *testing.T) {
	for _, mt := range []string{"atom", "derive"} {
		if !IsValidMetricType(mt) {
			t.Fatalf("IsValidMetricType(%q) = false, want true", mt)
		}
	}
	for _, mt := range []string{"", "ATOM", "DERIVE", "atomic", "derived", "atom "} {
		if IsValidMetricType(mt) {
			t.Fatalf("IsValidMetricType(%q) = true, want false", mt)
		}
	}
}

// 原子指标口径校验：biz_field/agg_type 必填且 agg_type 必须在白名单内
func TestValidateMetricSpecAtom(t *testing.T) {
	if err := validateMetricSpec("atom", "pay_amount", "sum", ""); err != nil {
		t.Fatalf("合法原子指标校验失败: %v", err)
	}
	if err := validateMetricSpec("atom", "", "sum", ""); !errors.Is(err, ErrBizFieldRequired) {
		t.Fatalf("缺 biz_field 应报 ErrBizFieldRequired, got %v", err)
	}
	if err := validateMetricSpec("atom", "pay_amount", "", ""); !errors.Is(err, ErrAggTypeRequired) {
		t.Fatalf("缺 agg_type 应报 ErrAggTypeRequired, got %v", err)
	}
	if err := validateMetricSpec("atom", "pay_amount", "median", ""); !errors.Is(err, ErrInvalidAggType) {
		t.Fatalf("白名单外 agg_type 应报 ErrInvalidAggType, got %v", err)
	}
}

// 派生指标口径校验：expr 必填（biz_field/agg_type 不参与校验）
func TestValidateMetricSpecDerive(t *testing.T) {
	if err := validateMetricSpec("derive", "", "", "sum(pay_amount) where status=1 by day"); err != nil {
		t.Fatalf("合法派生指标校验失败: %v", err)
	}
	if err := validateMetricSpec("derive", "pay_amount", "sum", ""); !errors.Is(err, ErrExprRequired) {
		t.Fatalf("缺 expr 应报 ErrExprRequired, got %v", err)
	}
}

package logic

import "testing"

// TestLevelRank 密级数字化与未知编码 fail-closed
func TestLevelRank(t *testing.T) {
	cases := []struct {
		code string
		want int
	}{
		{"L1", 1}, {"L2", 2}, {"L3", 3}, {"L4", 4}, {"L9", 9},
		{"L0", 0},  //数字越界
		{"L10", 0}, //两位数不匹配白名单
		{"l3", 0},  //小写不通过
		{"M1", 0},  //前缀非法
		{"", 0},    //空编码
	}
	for _, c := range cases {
		if got := LevelRank(c.code); got != c.want {
			t.Fatalf("LevelRank(%q) = %d, want %d", c.code, got, c.want)
		}
	}
}

// TestDecidePerm 密级比较：数据密级 ≤ 许可密级放行，未知密级一律拒绝
func TestDecidePerm(t *testing.T) {
	cases := []struct {
		name       string
		grantLevel string
		dataLevel  string
		wantAllow  bool
	}{
		{"L2 许可访问 L3 数据拒绝", "L2", "L3", false},
		{"L4 许可访问 L1 数据放行", "L4", "L1", true},
		{"同级放行", "L2", "L2", true},
		{"L4 许可不可覆盖 L9 自定义密级", "L4", "L9", false},
		{"未知数据密级拒绝", "L2", "X1", false},
		{"数据密级为空拒绝", "L2", "", false},
		{"未知许可密级拒绝", "L0", "L1", false},
		{"许可密级为空拒绝", "", "L1", false},
		{"L9 最高许可全覆盖", "L9", "L8", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			allowed, reason := DecidePerm(c.grantLevel, c.dataLevel)
			if allowed != c.wantAllow {
				t.Fatalf("DecidePerm(%q, %q) allowed = %v, want %v（reason=%q）",
					c.grantLevel, c.dataLevel, allowed, c.wantAllow, reason)
			}
			if !c.wantAllow && reason == "" {
				t.Fatalf("DecidePerm(%q, %q) 拒绝时应返回原因", c.grantLevel, c.dataLevel)
			}
		})
	}
}

// TestGrantExpired 授权过期判断：0 永久，等于过期时刻视为已过期
func TestGrantExpired(t *testing.T) {
	cases := []struct {
		name     string
		expireAt int64
		now      int64
		want     bool
	}{
		{"永久授权不过期", 0, 1000, false},
		{"未到过期时间", 1001, 1000, false},
		{"恰在过期时刻", 1000, 1000, true},
		{"已过过期时间", 999, 1000, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := GrantExpired(c.expireAt, c.now); got != c.want {
				t.Fatalf("GrantExpired(%d, %d) = %v, want %v", c.expireAt, c.now, got, c.want)
			}
		})
	}
}

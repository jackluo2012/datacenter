package logic

import "testing"

// TestValidSubject 授权主体格式：type:id 且 type ∈ user/app/role，id 段非空
func TestValidSubject(t *testing.T) {
	valid := []string{"user:1", "app:app_key", "role:admin", "user:10086"}
	for _, s := range valid {
		if !ValidSubject(s) {
			t.Fatalf("ValidSubject(%q) = false, want true", s)
		}
	}
	invalid := []string{
		"",        //整体为空
		"user",    //缺 id 段
		"user:",   //id 段为空
		"user:  ", //id 段纯空白
		"guest:1", //类型不在白名单
		"USER:1",  //类型大小写敏感
		":1",      //类型为空
	}
	for _, s := range invalid {
		if ValidSubject(s) {
			t.Fatalf("ValidSubject(%q) = true, want false", s)
		}
	}
}

// TestValidLevelCode 分级编码白名单 ^L[1-9]$
func TestValidLevelCode(t *testing.T) {
	for _, c := range []string{"L1", "L2", "L3", "L4", "L9"} {
		if !ValidLevelCode(c) {
			t.Fatalf("ValidLevelCode(%q) = false, want true", c)
		}
	}
	for _, c := range []string{"", "L0", "L10", "l1", "M1", "公开", " L1"} {
		if ValidLevelCode(c) {
			t.Fatalf("ValidLevelCode(%q) = true, want false", c)
		}
	}
}

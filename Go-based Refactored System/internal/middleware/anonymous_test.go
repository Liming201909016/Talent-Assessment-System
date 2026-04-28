package middleware

import "testing"

// ============================================================
// 回归测试 — FB-033 移动端预览页 401 异常
// 对应：docs/regression-tests.md FB-033
//
// 用户反馈：移动端打开 MBTI 预览页报"系统接口401异常"
// 根因：preview.vue 调用 /exam/api/tester/idNumber/{idNumber}
//       未在 anonymousPrefixes 列表中，考生无 token → 401
// ============================================================

func TestBugFB033_TesterIdNumberIsAnonymous(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"FB-033: tester/idNumber/xxx 必须匿名（preview 页要查考生信息）",
			"/exam/api/tester/idNumber/411111199001011234", true},
		{"FB-033: tester/idNumber 带 query 参数",
			"/exam/api/tester/idNumber/test?examId=x", true},

		// 现有规则不能被破坏
		{"login 必须匿名", "/login", true},
		{"captchaImage 必须匿名", "/captchaImage", true},
		{"exam/exam/detail 必须匿名（preview 用）", "/exam/api/exam/exam/detail", true},
		{"mbti/paper-detail 必须匿名", "/exam/api/mbti/paper-detail", true},
		{"tester/login 必须匿名", "/exam/api/tester/login", true},

		// 反例：管理类接口必须需要 token
		{"tester/list 不能匿名", "/exam/api/tester/list", false},
		{"qu/save 不能匿名", "/exam/api/qu/qu/save", false},
		{"system/user/list 不能匿名", "/system/user/list", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAnonymous(tt.path)
			if got != tt.want {
				t.Errorf("IsAnonymous(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

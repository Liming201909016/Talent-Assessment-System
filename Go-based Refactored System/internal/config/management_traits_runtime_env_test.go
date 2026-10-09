package config

import "testing"

// TestBugFB216_ManagementTraitsTestRuntimeEnvironmentGate
// 对应：docs/regression-tests.md #FB-216
// 复现：正式服务器无独立环境总门禁，TEST routes 与 expiry worker 会随服务器启动。
// 期望：两个可信显式环境必须归一化后同为 local、staging 或 production；其他值全部关闭。
func TestBugFB216_ManagementTraitsTestRuntimeEnvironmentGate(t *testing.T) {
	tests := []struct {
		name, reportEnv, testEnv string
		want                     bool
	}{
		{"both local", "local", "local", true},
		{"both staging", "staging", "staging", true},
		{"trim lowercase local", " LOCAL ", " local\t", true},
		{"trim lowercase staging", "StAgInG", " STAGING ", true},
		{"both production", "production", "production", true},
		{"trim lowercase production", " ProDuction ", "production\t", true},
		{"both empty", "", "", false},
		{"report only", "staging", "", false},
		{"test only", "", "staging", false},
		{"mismatch allowed names", "local", "staging", false},
		{"prod", "prod", "prod", false},
		{"malformed suffix", "staging-production", "staging-production", false},
		{"malformed internal whitespace", "sta ging", "sta ging", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("REPORT_EFFECTIVE_ENV", tt.reportEnv)
			t.Setenv("MNG_TEST_REPORT_ENV", tt.testEnv)
			if got := ManagementTraitsTestRuntimeEnabled(); got != tt.want {
				t.Fatalf("ManagementTraitsTestRuntimeEnabled()=%t, want %t", got, tt.want)
			}
		})
	}
}

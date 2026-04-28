package handler

import (
	"encoding/json"
	"testing"
)

// ============================================================
// FB-019 回归测试：qu.Paging 必须支持 params.repoId（单数）
// 对应：docs/regression-tests.md FB-019
// ============================================================

// TestBugFB019_QuPagingReq_AcceptsParamsRepoId
// 前端发送 {params: {repoId: "xxx"}} 时必须能正确解析
// 历史 bug：之前只支持 params.repoIds（复数数组）和顶层 repoId，遗漏 params.repoId（嵌套单数）
func TestBugFB019_QuPagingReq_AcceptsParamsRepoId(t *testing.T) {
	tests := []struct {
		name           string
		json           string
		wantTopRepoID  string
		wantParRepoID  string
		wantParRepoIds []string
	}{
		{
			name:          "顶层 repoId（旧格式）",
			json:          `{"current":1,"size":10,"repoId":"REPO-A"}`,
			wantTopRepoID: "REPO-A",
		},
		{
			name:          "params.repoId（前端实际发送格式 — FB-019）",
			json:          `{"current":1,"size":10,"params":{"repoId":"REPO-B"}}`,
			wantParRepoID: "REPO-B",
		},
		{
			name:           "params.repoIds 数组格式",
			json:           `{"current":1,"size":10,"params":{"repoIds":["REPO-C"]}}`,
			wantParRepoIds: []string{"REPO-C"},
		},
		{
			name:           "三种共存：优先顶层，其次嵌套单数，最后数组",
			json:           `{"current":1,"size":10,"repoId":"TOP","params":{"repoId":"PAR","repoIds":["ARR"]}}`,
			wantTopRepoID:  "TOP",
			wantParRepoID:  "PAR",
			wantParRepoIds: []string{"ARR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req quPagingReq
			if err := json.Unmarshal([]byte(tt.json), &req); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if req.RepoID != tt.wantTopRepoID {
				t.Errorf("top RepoID: want %q, got %q", tt.wantTopRepoID, req.RepoID)
			}
			if req.Params.RepoID != tt.wantParRepoID {
				t.Errorf("Params.RepoID: want %q, got %q", tt.wantParRepoID, req.Params.RepoID)
			}
			if len(req.Params.RepoIds) != len(tt.wantParRepoIds) {
				t.Errorf("Params.RepoIds length: want %d, got %d", len(tt.wantParRepoIds), len(req.Params.RepoIds))
			}
		})
	}
}

// TestBugFB019_RepoIDResolution
// 验证业务层选择 repoID 的优先级：顶层 > 嵌套单数 > 嵌套数组首项
// 模拟 handler 中的真实逻辑
func TestBugFB019_RepoIDResolution(t *testing.T) {
	resolve := func(req *quPagingReq) string {
		repoID := req.RepoID
		if repoID == "" {
			repoID = req.Params.RepoID
		}
		if repoID == "" && len(req.Params.RepoIds) > 0 {
			repoID = req.Params.RepoIds[0]
		}
		return repoID
	}

	tests := []struct {
		name     string
		req      quPagingReq
		expected string
	}{
		{"全空", quPagingReq{}, ""},
		{"仅顶层", quPagingReq{RepoID: "TOP"}, "TOP"},
		{"仅嵌套单数", func() quPagingReq { var r quPagingReq; r.Params.RepoID = "P"; return r }(), "P"},
		{"仅嵌套数组", func() quPagingReq { var r quPagingReq; r.Params.RepoIds = []string{"A1", "A2"}; return r }(), "A1"},
		{"顶层 + 嵌套单数（顶层优先）", func() quPagingReq { r := quPagingReq{RepoID: "TOP"}; r.Params.RepoID = "P"; return r }(), "TOP"},
		{"嵌套单数 + 嵌套数组（单数优先）", func() quPagingReq {
			var r quPagingReq
			r.Params.RepoID = "P"
			r.Params.RepoIds = []string{"A"}
			return r
		}(), "P"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolve(&tt.req); got != tt.expected {
				t.Errorf("want %q, got %q", tt.expected, got)
			}
		})
	}
}

package handler

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/talent-assessment/refactored/internal/service"
)

func TestManagementTraitsTestReportsEnvironmentPolicy(t *testing.T) {
	h := &ManagementTraitsRuntimeHandler{}
	t.Setenv("MNG_TEST_REPORT_DIR", t.TempDir())
	t.Setenv("MNG_TEST_CONTENT_PATH", filepath.Join("..", "..", "configs", "export-templates", "management-traits-002-test-content-v1.xlsx"))
	for _, env := range []string{"", "prod", "sta ging"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("REPORT_EFFECTIVE_ENV", env)
			t.Setenv("MNG_TEST_REPORT_ENV", env)
			if _, _, _, err := h.testReportInputs(false); !errors.Is(err, service.ErrManagementTraitsRuntimeClosed) {
				t.Fatal("unapproved environment reached content loader", err)
			}
		})
	}
	for _, env := range []string{"local", "staging", "production"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("REPORT_EFFECTIVE_ENV", env)
			t.Setenv("MNG_TEST_REPORT_ENV", env)
			_, content, _, err := h.testReportInputs(false)
			if err != nil || content.RuleCount() != 205 {
				t.Fatal("approved TEST environment rejected", err)
			}
		})
	}
}

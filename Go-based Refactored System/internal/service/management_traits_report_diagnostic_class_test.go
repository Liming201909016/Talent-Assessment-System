package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
)

type managementDiagnosticSecretError struct{}

func (managementDiagnosticSecretError) Error() string {
	panic("diagnostics must never call an unknown Error method")
}

// MT-REPORT-DIAG-01: classification must use real error identities/types;
// unknown, nested and wrapped secret-bearing errors have no textual fallback.
func TestBugMTReportDiag_AllowlistedClassesNeverFormatErrors(t *testing.T) {
	for _, x := range []struct {
		name, class string
		err         error
	}{
		{"unknown", "unknown", managementDiagnosticSecretError{}},
		{"nested", "unknown", errors.Join(errors.New(managementDiagnosticSecret), managementDiagnosticSecretError{})},
		{"wrapped", "unknown", fmt.Errorf("%s: %w", managementDiagnosticSecret, errors.New(managementDiagnosticSecret))},
		{"deadline", "context_deadline", context.DeadlineExceeded},
		{"canceled", "context_canceled", context.Canceled},
		{"permission", "os_permission", &os.PathError{Op: managementDiagnosticSecret, Path: managementDiagnosticSecret, Err: os.ErrPermission}},
		{"missing-path", "os_path", &os.PathError{Path: managementDiagnosticSecret, Err: os.ErrNotExist}},
		{"missing-sentinel", "os_not_exist", os.ErrNotExist},
		{"exists", "os_exists", os.ErrExist},
		{"path", "os_path", &os.PathError{Path: managementDiagnosticSecret, Err: errors.New(managementDiagnosticSecret)}},
		{"mysql", "mysql_1062", &mysql.MySQLError{Number: 1062, Message: managementDiagnosticSecret}},
		{"mysql-other", "mysql_65535", &mysql.MySQLError{Number: 65535, Message: managementDiagnosticSecret}},
		{"exit", "exec_exit", &exec.ExitError{Stderr: []byte(managementDiagnosticSecret)}},
		{"start", "exec_start", &exec.Error{Name: managementDiagnosticSecret, Err: exec.ErrNotFound}},
		{"validation", "validation", ErrManagementTraitsRuntimeInvalid},
		{"closed", "runtime_closed", ErrManagementTraitsRuntimeClosed},
	} {
		t.Run(x.name, func(t *testing.T) {
			out := managementDiagnosticCapture(t)
			wrapped := errors.Join(errors.New(managementDiagnosticSecret), x.err)
			d := &managementTraitsReportDiagnostic{stage: "render"}
			d.failure(wrapped)
			d.emit()
			if d.class != x.class || strings.Contains(out.String(), managementDiagnosticSecret) || !strings.Contains(out.String(), `"class":"`+x.class+`"`) {
				t.Fatal("unsafe or incorrect diagnostic class")
			}
		})
	}
	t.Run("reject-arbitrary-stage", func(t *testing.T) {
		out := managementDiagnosticCapture(t)
		d := &managementTraitsReportDiagnostic{stage: managementDiagnosticSecret}
		d.failure(errors.New(managementDiagnosticSecret))
		d.emit()
		if out.Len() != 0 {
			t.Fatal("arbitrary stage emitted")
		}
	})
	t.Run("nil-no-diagnostic", func(t *testing.T) {
		out := managementDiagnosticCapture(t)
		var d *managementTraitsReportDiagnostic
		d.failure(errors.New(managementDiagnosticSecret))
		d.emit()
		if out.Len() != 0 {
			t.Fatal("download/helper emitted generation log")
		}
	})
}

func TestBugMTReportDiag_RealFileFailureClass(t *testing.T) {
	root := filepath.Join(t.TempDir(), "regular.pdf")
	if os.WriteFile(root, []byte("old"), 0600) != nil {
		t.Fatal("fixture creation failed")
	}
	err := os.MkdirAll(root, 0700)
	var path *os.PathError
	if err == nil || !errors.As(err, &path) {
		t.Fatal("expected actual typed path failure")
	}
	if managementTraitsReportErrorClass(err) != "os_path" {
		t.Fatal("typed path failure misclassified")
	}
	d := &managementTraitsReportDiagnostic{stage: "write"}
	_, returned := writeManagementTraitsTestPDF(root, "22222222-2222-4222-8222-222222222222", append([]byte("%PDF-"), make([]byte, 2048)...), d)
	if returned != ErrManagementTraitsRuntimeInvalid {
		t.Fatal("write sentinel changed")
	}
	if d.class != "os_path" {
		t.Fatal("actual write path cause was lost")
	}
}

package libreofficepdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestClientConvertWritesValidPDFAndCleansWorkspace(t *testing.T) {
	runner := &fakeRunner{pdf: append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("x"), 2048)...)}
	client := newClient("libreoffice", runner)

	pdf, err := client.Convert(context.Background(), "report.docx", []byte("docx"))
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !bytes.Equal(pdf, runner.pdf) {
		t.Fatal("converted PDF bytes changed")
	}
	if runner.command != "libreoffice" || !containsArgument(runner.args, "--headless") || !containsArgument(runner.args, "--convert-to") {
		t.Fatalf("unexpected command: %s %v", runner.command, runner.args)
	}
	if runner.outDir == "" {
		t.Fatal("conversion output directory was not captured")
	}
	if _, statErr := os.Stat(runner.outDir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("temporary conversion workspace remains: %v", statErr)
	}
}

func TestClientConvertRejectsNonPDFAndCleansWorkspace(t *testing.T) {
	runner := &fakeRunner{pdf: []byte("not a pdf")}
	client := newClient("libreoffice", runner)

	if _, err := client.Convert(context.Background(), "report.docx", []byte("docx")); err == nil {
		t.Fatal("non-PDF output was accepted")
	}
	if _, statErr := os.Stat(runner.outDir); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("temporary conversion workspace remains after invalid output: %v", statErr)
	}
}

func TestClientConvertDoesNotExposeCommandOutput(t *testing.T) {
	runner := &fakeRunner{runErr: errors.New("exit status 1"), output: []byte("secret-from-command-output")}
	client := newClient("libreoffice", runner)

	_, err := client.Convert(context.Background(), "report.docx", []byte("docx"))
	if err == nil {
		t.Fatal("command failure was accepted")
	}
	if strings.Contains(err.Error(), "secret-from-command-output") {
		t.Fatalf("command output leaked in error: %v", err)
	}
}

func TestClientConvertRejectsUnsafeFileName(t *testing.T) {
	client := newClient("libreoffice", &fakeRunner{})
	if _, err := client.Convert(context.Background(), "../report.docx", []byte("docx")); err == nil {
		t.Fatal("path traversal file name was accepted")
	}
}

func TestClientConvertQueueHonorsContextCancellation(t *testing.T) {
	conversionSlot <- struct{}{}
	defer func() { <-conversionSlot }()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := newClient("libreoffice", &fakeRunner{})
	if _, err := client.Convert(ctx, "report.docx", []byte("docx")); err == nil || !strings.Contains(err.Error(), "排队超时") {
		t.Fatalf("cancelled queue error=%v", err)
	}
}

func TestNewClientUsesWaitableWindowsConsoleExecutable(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific LibreOffice launcher")
	}
	client := NewClient("")
	if !strings.EqualFold(filepath.Ext(client.executable), ".com") {
		t.Fatalf("Windows executable=%q, want soffice.com", client.executable)
	}
}

func TestLocalFileURLUsesLibreOfficeCompatibleWindowsURI(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific file URI")
	}
	if got := localFileURL(`C:\Temp\phase1 profile`); got != "file:///C:/Temp/phase1%20profile" {
		t.Fatalf("file URI=%q", got)
	}
}

type fakeRunner struct {
	command string
	args    []string
	outDir  string
	pdf     []byte
	output  []byte
	runErr  error
}

func (r *fakeRunner) Run(_ context.Context, command string, args ...string) ([]byte, error) {
	r.command = command
	r.args = append([]string(nil), args...)
	for index, argument := range args {
		if argument == "--outdir" && index+1 < len(args) {
			r.outDir = args[index+1]
		}
	}
	if r.runErr != nil {
		return r.output, r.runErr
	}
	if r.outDir == "" || len(args) == 0 {
		return nil, errors.New("missing conversion arguments")
	}
	inputPath := args[len(args)-1]
	pdfName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath)) + ".pdf"
	if err := os.WriteFile(filepath.Join(r.outDir, pdfName), r.pdf, 0o600); err != nil {
		return nil, err
	}
	return r.output, nil
}

func containsArgument(arguments []string, expected string) bool {
	for _, argument := range arguments {
		if argument == expected {
			return true
		}
	}
	return false
}

// MT-REPORT-DIAG-01: retain typed causes, not raw command output/text, with
// original public messages and the same isolated-workspace cleanup.
func TestBugMTReportDiag_ConverterRetainsSafeFailureIdentity(t *testing.T) {
	secret := "INJECTED_CONVERTER_SECRET_PATH_OUTPUT"
	for _, x := range []struct {
		name, class, message string
		cause                error
	}{
		{"unknown", "lo_command", "LibreOffice转换PDF失败", errors.New(secret)},
		{"permission", "lo_command", "LibreOffice转换PDF失败", &os.PathError{Op: secret, Path: secret, Err: os.ErrPermission}},
		{"exit", "lo_command", "LibreOffice转换PDF失败", &exec.ExitError{Stderr: []byte(secret)}},
		{"start", "lo_command", "LibreOffice转换PDF失败", &exec.Error{Name: secret, Err: exec.ErrNotFound}},
	} {
		t.Run(x.name, func(t *testing.T) {
			runner := &fakeRunner{runErr: x.cause, output: []byte(secret)}
			_, err := newClient("libreoffice", runner).Convert(context.Background(), "report.docx", []byte("docx"))
			var classified *ConversionError
			if !errors.As(fmt.Errorf("wrapped: %w", err), &classified) || classified.DiagnosticClass() != x.class || !errors.Is(err, x.cause) || err.Error() != x.message || strings.Contains(err.Error(), secret) {
				t.Fatal("converter cause or fixed public message changed")
			}
			if _, err := os.Stat(runner.outDir); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("converter workspace remains")
			}
		})
	}
	t.Run("queue-canceled", func(t *testing.T) {
		conversionSlot <- struct{}{}
		defer func() { <-conversionSlot }()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := newClient("libreoffice", &fakeRunner{}).Convert(ctx, "report.docx", []byte("docx"))
		var classified *ConversionError
		if !errors.As(err, &classified) || classified.DiagnosticClass() != "lo_queue" || !errors.Is(err, context.Canceled) || err.Error() != "LibreOffice转换排队超时" {
			t.Fatal("queue classification changed behavior")
		}
	})
	t.Run("non-pdf", func(t *testing.T) {
		runner := &fakeRunner{pdf: []byte(secret)}
		_, err := newClient("libreoffice", runner).Convert(context.Background(), "report.docx", []byte("docx"))
		var classified *ConversionError
		if !errors.As(err, &classified) || classified.DiagnosticClass() != "lo_output_invalid" || err.Error() != "LibreOffice返回的文件不是有效PDF" || strings.Contains(err.Error(), secret) {
			t.Fatal("invalid output leaked")
		}
		if _, err := os.Stat(runner.outDir); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("invalid-output workspace remains")
		}
	})
}

type diagnosticConverterRunner struct {
	wait      bool
	workspace string
}

func (r *diagnosticConverterRunner) Run(ctx context.Context, _ string, args ...string) ([]byte, error) {
	for i, arg := range args {
		if arg == "--outdir" {
			r.workspace = args[i+1]
		}
	}
	if r.wait {
		<-ctx.Done()
		return []byte("INJECTED_SECRET_OUTPUT"), ctx.Err()
	}
	return nil, nil
}

func TestBugMTReportDiag_ConverterLocalCategories(t *testing.T) {
	for _, x := range []struct{ name, class string }{
		{"config", "lo_config"}, {"input-size", "lo_input_size"}, {"input-name", "lo_input_name"},
		{"workspace", "lo_workspace"}, {"missing-output", "lo_output_open"}, {"command-deadline", "lo_command"},
	} {
		t.Run(x.name, func(t *testing.T) {
			runner := &diagnosticConverterRunner{wait: x.name == "command-deadline"}
			client := newClient("libreoffice", runner)
			ctx := context.Background()
			name, docx := "report.docx", []byte("docx")
			switch x.name {
			case "config":
				client = nil
			case "input-size":
				docx = nil
			case "input-name":
				name = "../INJECTED_SECRET_PATH.docx"
			case "workspace":
				missing := filepath.Join(t.TempDir(), "nonexistent")
				t.Setenv("TMP", missing)
				t.Setenv("TEMP", missing)
				t.Setenv("TMPDIR", missing)
			case "command-deadline":
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 20*time.Millisecond)
				defer cancel()
			}
			_, err := client.Convert(ctx, name, docx)
			var classified *ConversionError
			if !errors.As(err, &classified) || classified.DiagnosticClass() != x.class || strings.Contains(err.Error(), "INJECTED_SECRET") {
				t.Fatal("local converter category unavailable or leaked")
			}
			if x.name == "command-deadline" && (!errors.Is(err, context.DeadlineExceeded) || err.Error() != "LibreOffice转换PDF超时") {
				t.Fatal("command deadline cause/message lost")
			}
			if runner.workspace != "" {
				if _, err := os.Stat(runner.workspace); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("local converter workspace remains")
				}
			}
		})
	}
}

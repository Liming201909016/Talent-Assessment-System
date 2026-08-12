package libreofficepdf

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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

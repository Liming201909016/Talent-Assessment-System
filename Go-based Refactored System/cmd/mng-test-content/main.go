package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/talent-assessment/refactored/internal/service"
)

func main() {
	source := flag.String("source", "", "exact customer XLSX (read-only)")
	output := flag.String("output", "", "independent TEST-only runtime copy")
	flag.Parse()
	if err := copyContent(*source, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func copyContent(source, output string) error {
	if source == "" || output == "" {
		return fmt.Errorf("source and output required")
	}
	a, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	b, err := filepath.Abs(output)
	if err != nil || a == b {
		return fmt.Errorf("invalid output")
	}
	raw, err := os.ReadFile(a)
	if err != nil {
		return err
	}
	content, err := service.LoadManagementTraitsTestContent(raw)
	if err != nil || content.RuleCount() != 205 {
		return fmt.Errorf("exact TEST content rejected")
	}
	if old, err := os.ReadFile(b); err == nil {
		if !bytes.Equal(old, raw) {
			return fmt.Errorf("refusing to overwrite existing content")
		}
	} else {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(b), 0750); err != nil {
			return err
		}
		f, err := os.OpenFile(b, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		_, e := f.Write(raw)
		closeErr := f.Close()
		if e != nil {
			return e
		}
		if closeErr != nil {
			return closeErr
		}
	}
	fmt.Printf("TEST_ONLY_CONTENT_SHA=%s rules=%d\n", content.SourceSHA(), content.RuleCount())
	return nil
}

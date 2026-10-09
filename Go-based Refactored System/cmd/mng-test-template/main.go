package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/talent-assessment/refactored/internal/handler"
)

func main() {
	source := flag.String("source", "", "SHA-locked candidate DOCX")
	output := flag.String("output", "", "new test-only DOCX; different existing files are never overwritten")
	flag.Parse()
	if err := build(*source, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func build(source, output string) error {
	if source == "" || output == "" {
		return fmt.Errorf("source and output are required")
	}
	a, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	b, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if a == b {
		return fmt.Errorf("source must not be the output")
	}
	raw, err := os.ReadFile(a)
	if err != nil {
		return err
	}
	data, err := handler.BuildManagementTraitsTestTemplate(raw)
	if err != nil {
		return err
	}
	if old, err := os.ReadFile(b); err == nil {
		if !bytes.Equal(old, data) {
			return fmt.Errorf("refusing to overwrite existing template")
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
		_, writeErr := f.Write(data)
		closeErr := f.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	fmt.Printf("TEST_TEMPLATE_SHA256=%x bytes=%d\n", sha256.Sum256(data), len(data))
	return nil
}

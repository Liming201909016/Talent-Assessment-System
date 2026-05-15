package handler

import (
	"errors"
	"fmt"
	"testing"
)

// TestIsRetryableReportErr_BusinessErrors verifies business errors are NOT retryable
func TestIsRetryableReportErr_BusinessErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"试卷不存在", errors.New("试卷不存在或未关联考生")},
		{"不支持的 repoCode", fmt.Errorf("不支持的 repoCode: %s", "999")},
		{"MBTI 拒绝", errors.New("MBTI 报告由现有 docx→PDF 流程生成，不走 chromedp")},
		{"PDF 太小", fmt.Errorf("生成的 PDF 太小 (%d 字节)，疑似失败", 100)},
		{"paperId 为空", errors.New("paperId 为空")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if isRetryableReportErr(c.err) {
				t.Errorf("expected NOT retryable for business error: %v", c.err)
			}
		})
	}
}

// TestIsRetryableReportErr_InfraErrors verifies infrastructure errors ARE retryable
func TestIsRetryableReportErr_InfraErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"chromedp render", errors.New("chromedp render: page crashed")},
		{"context deadline", errors.New("context deadline exceeded")},
		{"wait ready timeout", errors.New("pdfgen: wait ready timeout (window.__reportReady === true)")},
		{"pool slot timeout", errors.New("pdfgen: wait pool slot timeout: context deadline exceeded")},
		{"pdfcpu merge", errors.New("pdfcpu merge: invalid syntax")},
		{"file IO", errors.New("open /tmp/x: no such file or directory")},
		{"network", errors.New("connection refused")},
		{"generic", errors.New("some unexpected error")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !isRetryableReportErr(c.err) {
				t.Errorf("expected retryable for infra error: %v", c.err)
			}
		})
	}
}

// TestIsRetryableReportErr_Nil verifies nil error returns false (not retryable since nothing to retry)
func TestIsRetryableReportErr_Nil(t *testing.T) {
	if isRetryableReportErr(nil) {
		t.Error("nil error should not be retryable")
	}
}

// TestIsRetryableReportErr_Wrapped verifies %w wrapped errors are still classified
func TestIsRetryableReportErr_Wrapped(t *testing.T) {
	innerBiz := errors.New("试卷不存在")
	wrappedBiz := fmt.Errorf("layer1: %w", innerBiz)
	if isRetryableReportErr(wrappedBiz) {
		t.Error("wrapped business error should still be classified as not retryable")
	}

	innerInfra := errors.New("context deadline exceeded")
	wrappedInfra := fmt.Errorf("chromedp: %w", innerInfra)
	if !isRetryableReportErr(wrappedInfra) {
		t.Error("wrapped infra error should still be classified as retryable")
	}
}

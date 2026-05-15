// Package pdfgen 通过 chromedp 调用本机 Chromium，把前端报告页渲染为 PDF。
//
// 用途：管理端"批量生成报告" / "单个生成报告" 触发后端生成，
// 替代前端 html2canvas+jsPDF 的慢速方案。
//
// 资源限制：服务器仅 2vCPU/4GB，pool size 默认 2，并发渲染上限 = poolSize。
package pdfgen

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// Pool 维护一组 chromedp 浏览器上下文，复用以避免频繁启动 Chrome 开销。
type Pool struct {
	chromePath string
	size       int
	semaphore  chan struct{}
	allocCtx   context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	closed     bool
}

// NewPool 启动一个 chromedp 浏览器实例池。
// chromePath 为空时使用 chromedp 默认（自动查找）。
func NewPool(chromePath string, size int) (*Pool, error) {
	if size <= 0 {
		size = 1
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-extensions", true),
		// 禁用所有缓存，确保每次都拉最新前端
		chromedp.Flag("disable-application-cache", true),
		chromedp.Flag("disable-cache", true),
		chromedp.Flag("disk-cache-size", "0"),
		chromedp.Flag("media-cache-size", "0"),
		// FB-045: 提高字体抗锯齿和图像渲染质量
		chromedp.Flag("font-render-hinting", "max"),
		chromedp.Flag("enable-font-antialiasing", true),
		// FB-045: viewport 用 2x DPR，让所有 PNG/canvas/字体在 PDF 中清晰度翻倍
		chromedp.Flag("force-device-scale-factor", "2"),
	)
	if chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	p := &Pool{
		chromePath: chromePath,
		size:       size,
		semaphore:  make(chan struct{}, size),
		allocCtx:   allocCtx,
		cancel:     cancel,
	}
	slog.Info("[pdfgen] pool initialized", "size", size, "chromePath", chromePath)
	return p, nil
}

// Close 关闭池及底层浏览器进程。
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	if p.cancel != nil {
		p.cancel()
	}
}

// GeneratePDF 渲染指定 URL 并返回 PDF 字节。
//
// readyJS：在 Chrome 端运行的"就绪表达式"，返回 true 时认为页面渲染完毕。
// 例如：`window.__reportReady === true`
//
// 超时由 ctx 控制；如果 readyJS 在超时内未返回 true，则强制超时失败。
//
// 返回 incomplete=true 表示前端通过 window.__reportIncomplete 标记
// 数据未完全加载（部分 dict 失败/超时），调用方可选择重试或写日志。
func (p *Pool) GeneratePDF(ctx context.Context, url, readyJS, reportType, headerTitle string) (pdf []byte, incomplete bool, err error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, false, errors.New("pdfgen pool is closed")
	}
	p.mu.Unlock()

	// 限流：等待 pool 槽位
	select {
	case p.semaphore <- struct{}{}:
		defer func() { <-p.semaphore }()
	case <-ctx.Done():
		return nil, false, fmt.Errorf("pdfgen: wait pool slot timeout: %w", ctx.Err())
	}

	// 每次新建一个 tab（不长期保留），避免内存累积
	taskCtx, cancelTask := chromedp.NewContext(p.allocCtx)
	defer cancelTask()

	var pdfBytes []byte
	var incompleteFlag bool
	pollDelay := 200 * time.Millisecond

	runErr := chromedp.Run(taskCtx,
		chromedp.Navigate(url),
		// 主文档加载完成
		chromedp.WaitReady("body", chromedp.ByQuery),
		// 轮询 readyJS 直到 true 或 ctx 超时
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadline, ok := ctx.Deadline()
			if !ok {
				deadline = time.Now().Add(30 * time.Second)
			}
			for {
				if time.Now().After(deadline) {
					return fmt.Errorf("pdfgen: wait ready timeout (%s)", readyJS)
				}
				var ready bool
				if err := chromedp.Evaluate(readyJS, &ready).Do(ctx); err == nil && ready {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(pollDelay):
				}
			}
		}),
		// 读取前端 incomplete 标志（best-effort，错误忽略）
		chromedp.ActionFunc(func(ctx context.Context) error {
			_ = chromedp.Evaluate(`window.__reportIncomplete === true`, &incompleteFlag).Do(ctx)
			return nil
		}),
		// 打印为 PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			data, err := printPDF(ctx, reportType, headerTitle)
			if err != nil {
				return err
			}
			pdfBytes = data
			return nil
		}),
	)
	if runErr != nil {
		return nil, false, fmt.Errorf("pdfgen render: %w", runErr)
	}
	if incompleteFlag {
		slog.Warn("[pdfgen] page rendered with incomplete data", "url", url)
	}
	return pdfBytes, incompleteFlag, nil
}

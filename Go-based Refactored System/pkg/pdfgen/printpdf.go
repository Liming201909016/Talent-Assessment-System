package pdfgen

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdffont "github.com/pdfcpu/pdfcpu/pkg/font"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	pdfcputypes "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// CJK 字体安装（lazy + 一次性）
// 通过环境变量 PDFGEN_CJK_FONT 指定 .ttc 或 .ttf 路径。
// 例：PDFGEN_CJK_FONT=/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc
var (
	cjkFontOnce sync.Once
	cjkFontName string // 安装成功后可用的字体名（如 NotoSansCJKsc-Regular），空 = 不可用
)

func ensureCJKFont() string {
	cjkFontOnce.Do(func() {
		fontPath := os.Getenv("PDFGEN_CJK_FONT")
		if fontPath == "" {
			return
		}
		if _, err := os.Stat(fontPath); err != nil {
			fmt.Printf("[pdfgen] CJK font not found: %s err=%v\n", fontPath, err)
			return
		}
		// 设置 pdfcpu 用户字体目录（默认 ~/.config/pdfcpu/fonts），用 /tmp 避免 root home 权限问题
		fontDir := os.Getenv("PDFGEN_FONT_DIR")
		if fontDir == "" {
			fontDir = filepath.Join(os.TempDir(), "pdfcpu-fonts")
		}
		_ = os.MkdirAll(fontDir, 0o755)
		pdffont.UserFontDir = fontDir
		// 直接调用底层 API（pdfapi.InstallFonts 错误被 CLI 日志吞掉）
		var instErr error
		switch strings.ToLower(filepath.Ext(fontPath)) {
		case ".ttc":
			instErr = pdffont.InstallTrueTypeCollection(fontDir, fontPath)
		case ".ttf", ".otf":
			instErr = pdffont.InstallTrueTypeFont(fontDir, fontPath)
		default:
			instErr = fmt.Errorf("unsupported font ext: %s", filepath.Ext(fontPath))
		}
		if instErr != nil {
			// .ttc 包含多个 sub-font 时，部分 sub-font 解析失败（如 EOF）
			// 但前面的 sub-font 可能已成功写入 .gob 文件 → 检查目录非空就继续
			entries, _ := os.ReadDir(fontDir)
			hasGob := false
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".gob") {
					hasGob = true
					break
				}
			}
			if !hasGob {
				fmt.Printf("[pdfgen] InstallTrueType failed (no font installed): %v\n", instErr)
				return
			}
			fmt.Printf("[pdfgen] InstallTrueType partial: %v (continuing with installed fonts)\n", instErr)
		}
		if err := pdffont.LoadUserFonts(); err != nil {
			fmt.Printf("[pdfgen] LoadUserFonts failed: %v\n", err)
			return
		}
		fonts, err := pdfapi.ListFonts()
		if err != nil {
			fmt.Printf("[pdfgen] ListFonts failed: %v\n", err)
			return
		}
		// 解析 ListFonts 输出（含 "Corefonts:"/"Userfonts(...):" 标签 + 字体名缩进 + 可能的 " (N glyphs)" 后缀）
		extractName := func(s string) string {
			s = strings.TrimSpace(s)
			if i := strings.Index(s, "("); i > 0 {
				s = strings.TrimSpace(s[:i])
			}
			return s
		}
		coreFonts := map[string]bool{
			"Courier": true, "Courier-Bold": true, "Courier-BoldOblique": true, "Courier-Oblique": true,
			"Helvetica": true, "Helvetica-Bold": true, "Helvetica-BoldOblique": true, "Helvetica-Oblique": true,
			"Symbol": true, "Times-Bold": true, "Times-BoldItalic": true, "Times-Italic": true, "Times-Roman": true,
			"ZapfDingbats": true,
		}
		// 优先 SC/CJK/WQY，其次任意 non-core 用户字体
		for _, f := range fonts {
			name := extractName(f)
			low := strings.ToLower(name)
			if low == "" || strings.HasPrefix(low, "corefonts") || strings.HasPrefix(low, "userfonts") {
				continue
			}
			if strings.Contains(low, "cjk") || strings.Contains(low, "wenquanyi") || strings.Contains(low, "wqy") || strings.Contains(low, "noto") {
				cjkFontName = name
				break
			}
		}
		if cjkFontName == "" {
			for _, f := range fonts {
				name := extractName(f)
				low := strings.ToLower(name)
				if low == "" || strings.HasPrefix(low, "corefonts") || strings.HasPrefix(low, "userfonts") {
					continue
				}
				if !coreFonts[name] {
					cjkFontName = name
					break
				}
			}
		}
		fmt.Printf("[pdfgen] CJK font installed: %s (selected=%q from %d fonts: %v)\n", fontPath, cjkFontName, len(fonts), fonts)
	})
	return cjkFontName
}

// HTML 页眉模板：与 baseline 几乎贴左上角
// FB-036: 之前用 inline SVG 在 chromedp HF 模板里渲染成空方框，改用 CSS border 三角更可靠
// FB-038: 减小 header 内 margin（0.2/0.6 → 0.15/0.4cm），配合 marginTop 0.55→0.7in 防止与正文重叠
// FB-049: header 图标位置稍下移（top:0 → top:0.3cm），离顶边留呼吸空间
// FB-039: 直接用 baseline 的原 PNG 图标 (base64)，001 绿色三角折角 / 002 浅蓝双圆，与 baseline 视觉完全一致
// FB-040: baseline 图标紧贴页面左上角(0,0)，去掉 padding-left + margin-top，让 img 完全顶左顶上
// FB-044: Chrome HF 模板默认 body 有 padding/margin，需 <style> 强制重置；img 用 fixed 定位贴 0,0
//
//	占位符顺序：%s=icon base64 png, %d=icon width px, %d=icon height px, %s=title text
const headerHTMLTpl = `<style>html,body{margin:0 !important;padding:0 !important;border:0 !important;}body>div{margin:0 !important;padding:0 !important;}</style>
<div style="width:100%%;font-family:'Noto Sans CJK SC','SimSun',sans-serif;position:relative;margin:0;padding:0;height:100%%;">
<img src="data:image/png;base64,%s" style="width:%dpx;height:%dpx;position:absolute;left:0;top:0.3cm;margin:0;padding:0;display:block;" />
<span style="position:absolute;right:0.2cm;top:0.55cm;font-size:8pt;color:#888888;letter-spacing:0;font-weight:400;">%s</span>
</div>`

// HTML 页脚模板：留空（实际页码用 pdfcpu 后处理添加，避免 chromedp 内置 pageNumber 跳号）
const footerHTML = `<div></div>`

// buildHeaderHTML 根据报告类型生成对应的 header（不同图标 + 标题）
// reportType: "001" → 绿色三角图标 / "002" → 浅蓝双圆图标 / 默认 → 001
func buildHeaderHTML(reportType, headerTitle string) string {
	icon := headerIcon001Base64
	w, h := 20, 22
	if reportType == "002" {
		icon = headerIcon002Base64
		w, h = 21, 38
	}
	return fmt.Sprintf(headerHTMLTpl, icon, w, h, headerTitle)
}

// printPDF 用 Chrome DevTools Protocol 的 Page.printToPDF 输出 PDF。
// 两阶段渲染（封面页无页眉页脚 + 后续页有）然后用 pdfcpu 合并。
//
// 阶段 1：仅打印第 1 页，无 header/footer，边距小（封面）
// 阶段 2：打印第 2 页起，加 header/footer
// 合并：用 pdfcpu Merge
func printPDF(ctx context.Context, reportType, headerTitle string) ([]byte, error) {
	if headerTitle == "" {
		headerTitle = "职业心理测评报告"
	}
	headerHTML := buildHeaderHTML(reportType, headerTitle)
	// 阶段 1：封面页
	cover, err := printRange(ctx, "1", false, headerHTML)
	if err != nil {
		return nil, fmt.Errorf("print cover: %w", err)
	}
	// 阶段 2：内容页（前端 .one-page 用 page-break-after:always 后会有空白第 2 页）
	// 跳过封面 + 跳过空白页，从第 3 页起才是真正内容
	content, err := printRange(ctx, "3-", true, headerHTML)
	if err != nil {
		// 内容页失败 → 仅返回封面（不报错，至少有封面）
		// 可能是只有 1 页报告（无内容页）
		return cover, nil
	}
	// 合并两份 PDF
	merged, err := mergePDFs([][]byte{cover, content})
	if err != nil {
		// 合并失败 → 返回不带页眉页脚的整份（兜底）
		return printRange(ctx, "", false, headerHTML)
	}
	// 合并后用 pdfcpu 给非首页加页码（封面不加，从内容第 1 页起编号 1）
	stamped, err := stampPageNumbers(merged)
	if err != nil {
		fmt.Printf("[pdfgen] stampPageNumbers failed: %v\n", err)
		return merged, nil
	}
	return stamped, nil
}

// printRange 打印指定页范围。
// pageRanges: "" = 全部, "1" = 仅首页, "2-" = 第 2 页起
// withHF: true = 加页眉页脚, false = 不加
func printRange(ctx context.Context, pageRanges string, withHF bool, headerHTML string) ([]byte, error) {
	var data []byte
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		p := page.PrintToPDF().WithPrintBackground(true)
		if withHF {
			p = p.WithDisplayHeaderFooter(true).
				WithHeaderTemplate(headerHTML).
				WithFooterTemplate(footerHTML).
				WithMarginTop(0.7).
				WithMarginBottom(1.1).
				WithMarginLeft(0.4).
				WithMarginRight(0.4)
		} else {
			// 封面：四边边距全 = 0，让封面图填满整页（与 baseline 一致）
			p = p.WithDisplayHeaderFooter(false).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0)
		}
		p = p.WithPaperWidth(8.27).
			WithPaperHeight(11.69).
			WithPreferCSSPageSize(false)
		if pageRanges != "" {
			p = p.WithPageRanges(pageRanges)
		}
		buf, _, err := p.Do(ctx)
		if err != nil {
			return err
		}
		data = buf
		return nil
	}))
	return data, err
}

// mergePDFs 用 pdfcpu 合并多份 PDF 字节流。
func mergePDFs(parts [][]byte) ([]byte, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("no parts to merge")
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	readers := make([]io.ReadSeeker, 0, len(parts))
	for _, b := range parts {
		readers = append(readers, bytes.NewReader(b))
	}
	conf := pdfmodel.NewDefaultConfiguration()
	out := &bytes.Buffer{}
	if err := pdfapi.MergeRaw(readers, out, false, conf); err != nil {
		return nil, fmt.Errorf("pdfcpu merge: %w", err)
	}
	return out.Bytes(), nil
}

// stampPageNumbers 给 PDF 加底部居中页码"第 N 页"。
// 第 1 页（封面）跳过，从第 2 页起依次显示"第 1 页, 第 2 页, ..."（与 baseline jsPDF 行为一致）。
func stampPageNumbers(pdfBytes []byte) ([]byte, error) {
	conf := pdfmodel.NewDefaultConfiguration()
	conf.Unit = pdfcputypes.POINTS

	// 先获取总页数
	pageCount, err := pdfapi.PageCount(bytes.NewReader(pdfBytes), conf)
	if err != nil {
		return nil, fmt.Errorf("pdfcpu PageCount: %w", err)
	}

	// 单份就跳过
	if pageCount <= 1 {
		return pdfBytes, nil
	}

	// 选字体：CJK 可用则用之（显示中文），否则降级为纯数字（拉丁字体）
	fontName := ensureCJKFont()
	useCJK := fontName != ""
	if !useCJK {
		fontName = "Helvetica"
	}

	// 为每一页（从第 2 页起）准备一个独立水印（页码不同）
	wmMap := make(map[int]*pdfmodel.Watermark, pageCount)
	for i := 2; i <= pageCount; i++ {
		var text string
		if useCJK {
			text = fmt.Sprintf("第 %d 页", i-1)
		} else {
			text = fmt.Sprintf("%d", i-1)
		}
		// position:bc 底居中，offset:dx dy 让水印向上偏移 28pt（0.99cm，接近 baseline 1.13cm）
		desc := fmt.Sprintf("position:bc, offset:0 28, scalefactor:1 abs, fontname:%s, points:9, fillcolor:#555555, opacity:1, rotation:0", fontName)
		wm, err := pdfapi.TextWatermark(text, desc, true, false, pdfcputypes.POINTS)
		if err != nil {
			return nil, fmt.Errorf("pdfcpu TextWatermark page %d: %w", i, err)
		}
		wmMap[i] = wm
	}

	out := &bytes.Buffer{}
	if err := pdfapi.AddWatermarksMap(bytes.NewReader(pdfBytes), out, wmMap, conf); err != nil {
		return nil, fmt.Errorf("pdfcpu AddWatermarksMap: %w", err)
	}
	return out.Bytes(), nil
}

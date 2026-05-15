package handler

import (
	"log/slog"
	"os"
	"os/exec"
)

// compressPDF 使用 ghostscript 压缩 PDF 文件（原地替换）。
// 压缩失败时保留原文件不影响业务。
//
// FB-046: 之前用 -dPDFSETTINGS=/ebook 会把所有图像重采样到 150 DPI 并 JPEG 有损压缩，
// 导致 echarts canvas / PNG 图标在 PDF 中明显模糊。
// 改用自定义参数：
//   - 流压缩（zip）保留 → 省 30-50% 大小
//   - 字体子集化保留 → 省额外大小
//   - 关闭图像重采样 → 图像 100% 无损
//   - 关闭有损 JPEG → 图像保持 chromedp 原始 PNG/Flate 编码
func compressPDF(path string) {
	tmp := path + ".gs.tmp"
	cmd := exec.Command("gs",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dNOPAUSE", "-dBATCH", "-dQUIET",
		// 字体子集化（默认开，显式标注）
		"-dSubsetFonts=true",
		"-dEmbedAllFonts=true",
		// FB-046: 关闭所有图像重采样（保持原始分辨率）
		"-dDownsampleColorImages=false",
		"-dDownsampleGrayImages=false",
		"-dDownsampleMonoImages=false",
		// FB-046: 不强制有损 JPEG，让 gs 自动选 Flate（无损）
		"-dAutoFilterColorImages=false",
		"-dAutoFilterGrayImages=false",
		"-dColorImageFilter=/FlateEncode",
		"-dGrayImageFilter=/FlateEncode",
		// 流压缩（默认开，显式标注）
		"-dCompressFonts=true",
		"-dCompressStreams=true",
		"-sOutputFile="+tmp,
		path,
	)
	if err := cmd.Run(); err != nil {
		slog.Warn("pdf-compress: gs failed", "path", path, "error", err)
		os.Remove(tmp)
		return
	}
	// 仅当压缩后更小时才替换
	origInfo, err1 := os.Stat(path)
	compInfo, err2 := os.Stat(tmp)
	if err1 != nil || err2 != nil || compInfo.Size() >= origInfo.Size() {
		slog.Info("pdf-compress: skip (no gain)", "path", path)
		os.Remove(tmp)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		slog.Warn("pdf-compress: rename failed", "error", err)
		os.Remove(tmp)
		return
	}
	slog.Info("pdf-compress: done",
		"path", path,
		"before", origInfo.Size(),
		"after", compInfo.Size(),
		"ratio", float64(compInfo.Size())/float64(origInfo.Size()),
	)
}

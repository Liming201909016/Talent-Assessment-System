package handler

import (
	"log/slog"
	"os"
	"os/exec"
)

// compressPDF 使用 ghostscript 压缩 PDF 文件（原地替换）。
// 压缩失败时保留原文件不影响业务。
func compressPDF(path string) {
	tmp := path + ".gs.tmp"
	cmd := exec.Command("gs",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/ebook",
		"-dNOPAUSE", "-dBATCH", "-dQUIET",
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

// Diag: run mbti template replacement standalone, output docx to /tmp/diag-replaced.docx
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

func stripXmlTags(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, "")
}

func findLabelEndInXml(xmlStr, label string) int {
	labelPos := 0
	labelRunes := []rune(label)
	i := 0
	for i < len(xmlStr) && labelPos < len(labelRunes) {
		if xmlStr[i] == '<' {
			closeIdx := strings.Index(xmlStr[i:], ">")
			if closeIdx < 0 {
				return -1
			}
			i += closeIdx + 1
			continue
		}
		r, sz := utf8.DecodeRuneInString(xmlStr[i:])
		if r == labelRunes[labelPos] {
			labelPos++
			i += sz
		} else {
			return -1
		}
	}
	if labelPos < len(labelRunes) {
		return -1
	}
	if strings.HasPrefix(xmlStr[i:], "</w:t>") {
		i += len("</w:t>")
	}
	return i
}

func locateLabelEnd(content, label string) int {
	if idx := strings.Index(content, label+"</w:t>"); idx >= 0 {
		return idx + len(label) + len("</w:t>")
	}
	labelRunes := []rune(label)
	if len(labelRunes) < 2 {
		return -1
	}
	firstChar := string(labelRunes[0])
	pos := 0
	for {
		p := strings.Index(content[pos:], firstChar)
		if p < 0 {
			return -1
		}
		absPos := pos + p
		lookback := absPos - 50
		if lookback < 0 {
			lookback = 0
		}
		if !strings.Contains(content[lookback:absPos], "<w:t") {
			pos = absPos + 1
			continue
		}
		tail := content[absPos:]
		if len(tail) > 1000 {
			tail = tail[:1000]
		}
		if strings.HasPrefix(stripXmlTags(tail), label) {
			if endIdx := findLabelEndInXml(tail, label); endIdx > 0 {
				return absPos + endIdx
			}
		}
		pos = absPos + 1
	}
}

func replaceDocxDate(content, dateStr string) string {
	// 简化版：跳过
	return content
}

func replaceDocumentFields(data []byte, fields map[string]string, dateStr string) []byte {
	content := string(data)
	for label, value := range fields {
		labelEnd := locateLabelEnd(content, label)
		if labelEnd < 0 {
			fmt.Fprintf(os.Stderr, "label %q NOT FOUND\n", label)
			continue
		}
		if value == "" {
			pStart := strings.LastIndex(content[:labelEnd], "<w:p ")
			pStart2 := strings.LastIndex(content[:labelEnd], "<w:p>")
			if pStart2 > pStart {
				pStart = pStart2
			}
			if pStart < 0 {
				continue
			}
			pEnd := strings.Index(content[labelEnd:], "</w:p>")
			if pEnd < 0 {
				continue
			}
			pEnd = labelEnd + pEnd + len("</w:p>")
			content = content[:pStart] + content[pEnd:]
			continue
		}
		afterLabel := content[labelEnd:]
		reT := regexp.MustCompile(`(<w:t[^>]*>)([^<]*)(</w:t>)`)
		matchT := reT.FindStringSubmatch(afterLabel)
		if matchT != nil {
			newT := matchT[1] + value + matchT[3]
			rebuilt := strings.Replace(afterLabel, matchT[0], newT, 1)
			rebuilt = strings.Replace(rebuilt, `<w:u w:val="single"/>`, "", 1)
			content = content[:labelEnd] + rebuilt
		}
	}
	content = replaceDocxDate(content, dateStr)
	return []byte(content)
}

func main() {
	templatePath := os.Args[1]
	outPath := os.Args[2]

	r, err := zip.OpenReader(templatePath)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	fields := map[string]string{
		"姓名：":   "懂",
		"年龄：":   "30",
		"性别：":   "女",
		"单位：":   "",
		"岗位：":   "",
		"联系方式：": "12345678992",
		"测评时长：": "5分钟",
	}

	for _, f := range r.File {
		rc, _ := f.Open()
		data, _ := io.ReadAll(rc)
		rc.Close()
		if f.Name == "word/document.xml" {
			data = replaceDocumentFields(data, fields, "2026年5月19日")
		}
		method := zip.Deflate
		if strings.HasSuffix(f.Name, "/") || len(data) == 0 {
			method = zip.Store
		}
		header := &zip.FileHeader{Name: f.Name, Method: method}
		writer, _ := w.CreateHeader(header)
		writer.Write(data)
	}
	w.Close()
	os.WriteFile(outPath, buf.Bytes(), 0o644)
	fmt.Println("wrote", outPath, len(buf.Bytes()), "bytes")
}

package handler

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	"github.com/talent-assessment/refactored/internal/service"
	"github.com/talent-assessment/refactored/pkg/response"
	"github.com/xuri/excelize/v2"
)

const phase1WordTemplateFileName = "competency-phase1-report.docx"

type phase1WordTemplateContract struct {
	SchemaVersion     string `json:"schemaVersion"`
	ContentControls   int    `json:"contentControls"`
	RegisteredFields  int    `json:"registeredFields"`
	UsedFields        int    `json:"usedFields"`
	Charts            int    `json:"charts"`
	BusinessCharts    int    `json:"businessCharts"`
	EmbeddedWorkbooks int    `json:"embeddedWorkbooks"`
	ExternalLinks     int    `json:"externalLinks"`
	VisibleTokens     int    `json:"visibleTokens"`
}

type phase1WordTemplateInfo struct {
	Exists          bool   `json:"exists"`
	FileName        string `json:"fileName"`
	Size            int64  `json:"size"`
	ModTime         string `json:"modTime"`
	SHA256          string `json:"sha256"`
	Valid           bool   `json:"valid"`
	ValidationError string `json:"validationError,omitempty"`
	phase1WordTemplateContract
}

func (h *CompetencyReportHandler) templateIdentity(c *gin.Context) (*model.LoginUser, bool) {
	value, ok := c.Get("loginUser")
	if !ok {
		response.AjaxUnauthorized(c, "")
		return nil, false
	}
	login, ok := value.(*model.LoginUser)
	if !ok || !canPublishCompetencyExam(login) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "msg": "无权管理胜任力报告模板"})
		return nil, false
	}
	return login, true
}

func (h *CompetencyReportHandler) phase1WordTemplatePath() (string, error) {
	if h == nil || h.examH == nil || h.examH.cfg == nil {
		return "", errors.New("一期Word报告模板未配置")
	}
	path := strings.TrimSpace(h.examH.cfg.Phase1WordReport.TemplatePath)
	if path == "" || !strings.EqualFold(filepath.Ext(path), ".docx") {
		return "", errors.New("一期Word报告模板路径无效")
	}
	return filepath.Clean(path), nil
}

func (h *CompetencyReportHandler) Phase1TemplateInfo(c *gin.Context) {
	if _, ok := h.templateIdentity(c); !ok {
		return
	}
	path, err := h.phase1WordTemplatePath()
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	h.templateMu.RLock()
	defer h.templateMu.RUnlock()
	info, err := readPhase1WordTemplateInfo(path)
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	response.Rest(c, info)
}

func (h *CompetencyReportHandler) DownloadPhase1Template(c *gin.Context) {
	if _, ok := h.templateIdentity(c); !ok {
		return
	}
	path, err := h.phase1WordTemplatePath()
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	h.templateMu.RLock()
	defer h.templateMu.RUnlock()
	file, err := os.Open(path)
	if err != nil {
		response.RestErr(c, "一期Word报告模板不存在")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		response.RestErr(c, "读取一期Word报告模板失败")
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", "attachment; filename="+phase1WordTemplateFileName+"; filename*=UTF-8''"+encodeRFC5987FileName(phase1WordTemplateFileName))
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if _, err := io.Copy(c.Writer, file); err != nil {
		return
	}
}

func (h *CompetencyReportHandler) UploadPhase1Template(c *gin.Context) {
	if _, ok := h.templateIdentity(c); !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxPhase1WordTemplateBytes+(1<<20))
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.RestErr(c, "请选择一期Word报告模板")
		return
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".docx") {
		response.RestErr(c, "只支持.docx格式")
		return
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxPhase1WordTemplateBytes {
		response.RestErr(c, "模板文件大小必须在20MB以内")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.RestErr(c, "读取上传模板失败")
		return
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxPhase1WordTemplateBytes+1))
	file.Close()
	if readErr != nil || len(data) == 0 || len(data) > maxPhase1WordTemplateBytes {
		response.RestErr(c, "读取上传模板失败")
		return
	}
	contract, err := validatePhase1WordTemplateUpload(data)
	if err != nil {
		response.RestErr(c, "模板校验失败："+err.Error())
		return
	}
	path, err := h.phase1WordTemplatePath()
	if err != nil {
		response.RestErr(c, err.Error())
		return
	}
	h.templateMu.Lock()
	backup, err := installPhase1WordTemplate(path, data, time.Now())
	h.templateMu.Unlock()
	if err != nil {
		response.RestErr(c, "保存模板失败")
		return
	}
	info, err := readPhase1WordTemplateInfo(path)
	if err != nil {
		response.RestErr(c, "读取已保存模板失败")
		return
	}
	info.phase1WordTemplateContract = contract
	response.Rest(c, gin.H{"template": info, "backupFile": filepath.Base(backup)})
}

func readPhase1WordTemplateInfo(path string) (phase1WordTemplateInfo, error) {
	info := phase1WordTemplateInfo{FileName: phase1WordTemplateFileName}
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return info, nil
	}
	if err != nil || stat.IsDir() {
		return info, errors.New("读取一期Word报告模板失败")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return info, errors.New("读取一期Word报告模板失败")
	}
	digest := sha256.Sum256(data)
	info.Exists = true
	info.Size = stat.Size()
	info.ModTime = stat.ModTime().Format("2006-01-02 15:04:05")
	info.SHA256 = hex.EncodeToString(digest[:])
	contract, validationErr := validatePhase1WordTemplateUpload(data)
	if validationErr != nil {
		info.ValidationError = validationErr.Error()
		return info, nil
	}
	info.Valid = true
	info.phase1WordTemplateContract = contract
	return info, nil
}

func validatePhase1WordTemplateUpload(data []byte) (phase1WordTemplateContract, error) {
	contract := phase1WordTemplateContract{}
	if len(data) == 0 || len(data) > maxPhase1WordTemplateBytes {
		return contract, errors.New("模板文件大小无效")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return contract, errors.New("DOCX文件结构无效")
	}
	parts := make(map[string][]byte, 32)
	for _, file := range reader.File {
		if file.Name != "[Content_Types].xml" && file.Name != "word/document.xml" && file.Name != "word/_rels/document.xml.rels" && file.Name != phase1ChartWorkbookPath && !strings.HasPrefix(file.Name, "word/charts/chart") && !strings.HasPrefix(file.Name, "word/charts/_rels/chart") {
			continue
		}
		rc, openErr := file.Open()
		if openErr != nil {
			return contract, errors.New("读取DOCX内容失败")
		}
		body, readErr := io.ReadAll(io.LimitReader(rc, maxPhase1WordTemplateBytes+1))
		rc.Close()
		if readErr != nil || len(body) > maxPhase1WordTemplateBytes {
			return contract, errors.New("读取DOCX内容失败")
		}
		parts[file.Name] = body
	}
	document, ok := parts["word/document.xml"]
	if !ok {
		return contract, errors.New("模板缺少Word正文")
	}
	tags := wordContentControlTagPattern.FindAllSubmatch(document, -1)
	seen := make(map[string]int, len(tags))
	for _, match := range tags {
		seen[string(match[1])]++
	}
	registry := phase1TemplateFieldRegistry()
	definitions := make(map[string]phase1TemplateFieldDefinition, len(registry))
	for _, field := range registry {
		definitions[field.Key] = field
	}
	for tag, count := range seen {
		definition, exists := definitions[tag]
		if !exists {
			return contract, fmt.Errorf("模板包含未支持的内容控件Tag：%s", tag)
		}
		if count > 1 && !definition.Repeatable {
			return contract, fmt.Errorf("内容控件Tag重复：%s", tag)
		}
	}
	missing := make([]string, 0)
	for _, field := range registry {
		if field.Required && seen[field.Key] == 0 {
			missing = append(missing, field.Key)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return contract, fmt.Errorf("缺少内容控件Tag：%s", strings.Join(missing, ", "))
	}
	visibleTokens := wordTemplateTokenPattern.FindAll(document, -1)
	if len(visibleTokens) > 0 {
		return contract, fmt.Errorf("模板存在可见占位符：%s", visibleTokens[0])
	}
	resolvedCharts, err := resolvePhase1BusinessChartParts(parts)
	if err != nil {
		return contract, err
	}
	contract.SchemaVersion = phase1WordTemplateSchemaV1
	if len(resolvedCharts) == 0 {
		for _, definition := range phase1TemplateChartRegistry() {
			chart, exists := parts[definition.LegacyPart]
			if !exists {
				return contract, fmt.Errorf("模板缺少图表：%s", filepath.Base(definition.LegacyPart))
			}
			if _, err := replaceWordChartValues(chart, make([]float64, definition.ValueCount)); err != nil {
				return contract, fmt.Errorf("图表数据点无效：%s", filepath.Base(definition.LegacyPart))
			}
		}
	} else {
		contract.SchemaVersion = phase1WordTemplateSchemaV2
		if err := validatePhase1V2ContentTypes(parts["[Content_Types].xml"]); err != nil {
			return contract, err
		}
		for index, definition := range phase1TemplateChartRegistry() {
			part := resolvedCharts[definition.Key]
			if _, err := replaceWordChartValues(parts[part], make([]float64, definition.ValueCount)); err != nil {
				return contract, fmt.Errorf("图表数据点无效：%s", definition.Key)
			}
			for _, formula := range phase1V2ChartFormulas(index) {
				if !strings.Contains(string(parts[part]), "ChartData!"+formula) {
					return contract, fmt.Errorf("图表数据区域无效：%s", definition.Key)
				}
			}
			relPart := "word/charts/_rels/" + filepath.Base(part) + ".rels"
			relations := string(parts[relPart])
			if !strings.Contains(relations, `relationships/package`) || !strings.Contains(relations, `../embeddings/competency-phase1-chart-data.xlsx`) || strings.Contains(relations, `TargetMode="External"`) {
				return contract, fmt.Errorf("图表内嵌工作簿关系无效：%s", definition.Key)
			}
		}
		if err := validatePhase1V2Workbook(parts[phase1ChartWorkbookPath], registry); err != nil {
			return contract, err
		}
	}
	externalLinks := 0
	for name, body := range parts {
		if strings.HasPrefix(name, "word/charts/_rels/") {
			externalLinks += strings.Count(string(body), `TargetMode="External"`)
		}
	}
	if contract.SchemaVersion == phase1WordTemplateSchemaV2 && externalLinks != 0 {
		return contract, errors.New("V2模板不得包含外部Excel链接")
	}
	contract.ContentControls = len(tags)
	contract.RegisteredFields = len(registry)
	contract.UsedFields = len(seen)
	contract.Charts = 12
	contract.BusinessCharts = len(resolvedCharts)
	if len(parts[phase1ChartWorkbookPath]) > 0 {
		contract.EmbeddedWorkbooks = 1
	}
	contract.ExternalLinks = externalLinks
	contract.VisibleTokens = len(visibleTokens)
	return contract, nil
}

func validatePhase1V2ContentTypes(contentTypes []byte) error {
	content := string(contentTypes)
	firstOverride := strings.Index(content, "<Override")
	lastDefault := strings.LastIndex(content, "<Default")
	if firstOverride < 0 || lastDefault < 0 || lastDefault > firstOverride {
		return errors.New("V2模板Content Types顺序无效")
	}
	workbookOverride := `<Override PartName="/word/embeddings/competency-phase1-chart-data.xlsx" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"/>`
	if !strings.Contains(content, workbookOverride) {
		return errors.New("V2模板缺少内嵌工作簿Content Type")
	}
	if strings.Contains(content, `<Default Extension="xlsx"`) {
		return errors.New("V2模板不得使用全局xlsx Content Type")
	}
	return nil
}

func phase1V2ChartFormulas(index int) []string {
	switch index {
	case 0:
		return []string{"$B$2:$B$3", "$C$2:$C$3"}
	case 1:
		return []string{"$C$1", "$B$4:$B$13", "$C$4:$C$13", "$D$1", "$D$4:$D$13"}
	default:
		row := index + 2
		return []string{fmt.Sprintf("$C$%d:$D$%d", row, row)}
	}
}

func requiredPhase1WordTemplateTags() []string {
	tags := make([]string, 0, 49)
	for _, field := range phase1TemplateFieldRegistry() {
		if field.Required {
			tags = append(tags, field.Key)
		}
	}
	return tags
}

func validatePhase1V2Workbook(workbook []byte, registry []phase1TemplateFieldDefinition) error {
	if len(workbook) == 0 {
		return errors.New("V2模板缺少内嵌图表工作簿")
	}
	book, err := excelize.OpenReader(bytes.NewReader(workbook))
	if err != nil {
		return errors.New("V2模板内嵌图表工作簿无效")
	}
	defer book.Close()
	if value, _ := book.GetCellValue("FieldDictionary", "B1"); value != phase1WordTemplateSchemaV2 {
		return errors.New("V2模板FieldDictionary契约版本无效")
	}
	rows, err := book.GetRows("FieldDictionary")
	if err != nil {
		return errors.New("V2模板FieldDictionary无效")
	}
	dictionaryKeys := make(map[string]bool, len(rows))
	for _, row := range rows {
		if len(row) > 0 {
			dictionaryKeys[row[0]] = true
		}
	}
	for _, field := range registry {
		if !dictionaryKeys[field.Key] {
			return fmt.Errorf("V2模板FieldDictionary缺少业务键：%s", field.Key)
		}
	}
	chartKeys := append([]string{"group.general_ability", "group.psychological_quality"}, service.NormalizePhase1CompetencyConfiguration().DimensionIDs...)
	for index, expected := range chartKeys {
		cell := fmt.Sprintf("A%d", index+2)
		if value, _ := book.GetCellValue("ChartData", cell); value != expected {
			return fmt.Errorf("V2模板ChartData业务键无效：%s", cell)
		}
	}
	return nil
}

func installPhase1WordTemplate(target string, data []byte, now time.Time) (string, error) {
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	backupDir := filepath.Join(dir, "backups")
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return "", err
	}
	backup := filepath.Join(backupDir, phase1WordTemplateFileName+"."+now.Format("20060102_150405_000")+".bak")
	var current []byte
	if existing, err := os.ReadFile(target); err == nil {
		current = append([]byte(nil), existing...)
		if err := os.WriteFile(backup, current, 0o600); err != nil {
			return "", err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	} else {
		backup = ""
	}
	temporary, err := os.CreateTemp(dir, ".phase1-word-template-*.docx")
	if err != nil {
		return "", err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return "", err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		if runtime.GOOS == "windows" && len(current) > 0 {
			_ = os.WriteFile(target, current, 0o644)
		}
		return "", err
	}
	return backup, nil
}

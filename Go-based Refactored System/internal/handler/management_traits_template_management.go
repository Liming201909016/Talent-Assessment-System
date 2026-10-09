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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/service"
)

const managementTraitsSharedTemplateFileName = "management-traits-00501-00502-shared.docx"
const maxManagementTraitsTemplateBytes = 20 << 20
const opcRelationshipNS = "http://schemas.openxmlformats.org/package/2006/relationships"

type managementTraitsTemplateInfo struct {
	Exists          bool     `json:"exists"`
	FileName        string   `json:"fileName"`
	Size            int64    `json:"size"`
	ModTime         string   `json:"modTime"`
	SHA256          string   `json:"sha256"`
	Valid           bool     `json:"valid"`
	ValidationError string   `json:"validationError,omitempty"`
	ProductCodes    []string `json:"productCodes"`
	ContentControls int      `json:"contentControls"`
	BusinessCharts  int      `json:"businessCharts"`
	NumericLabels   int      `json:"numericLabels"`
	ExternalLinks   int      `json:"externalLinks"`
}

func (h *ManagementTraitsRuntimeHandler) managementTraitsTemplatePath() (string, error) {
	path := strings.TrimSpace(os.Getenv("MNG_TEST_TEMPLATE_PATH"))
	if path == "" {
		path = filepath.Join("configs", "export-templates", "management-traits-002-test-only-v2.docx")
	}
	if !strings.EqualFold(filepath.Ext(path), ".docx") {
		return "", errors.New("005共用报告模板路径无效")
	}
	return filepath.Clean(path), nil
}

func (h *ManagementTraitsRuntimeHandler) managementTraitsTemplateSHAs() (map[string]bool, error) {
	path, err := h.managementTraitsTemplatePath()
	if err != nil {
		return nil, err
	}
	h.templateMu.RLock()
	defer h.templateMu.RUnlock()
	paths := []string{path}
	backupDir := filepath.Join(filepath.Dir(path), "backups")
	if entries, readErr := os.ReadDir(backupDir); readErr == nil {
		prefix := filepath.Base(path) + "."
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), ".bak") {
				paths = append(paths, filepath.Join(backupDir, entry.Name()))
			}
		}
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return nil, readErr
	}
	allowed := make(map[string]bool, len(paths))
	for _, candidate := range paths {
		data, readErr := os.ReadFile(candidate)
		if readErr != nil {
			return nil, readErr
		}
		digest := sha256.Sum256(data)
		allowed[hex.EncodeToString(digest[:])] = true
	}
	return allowed, nil
}

func managementTraitsTemplateError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": message, "success": false, "data": nil})
}

func managementTraitsTemplateProbeData() service.ManagementTraitsTestReportData {
	data := service.ManagementTraitsTestReportData{
		Schema:           "mng-test-report-data-v1",
		TestOnly:         true,
		RunID:            "template-probe-run",
		PaperID:          "template-probe-paper",
		ExamID:           "template-probe-exam",
		ContentSourceSHA: "b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c",
		Participant:      service.ManagementTraitsTestReportParticipant{Name: "模板校验", Gender: "0", Telephone: "18800000000", Affiliation: "校验单位", Post: "校验岗位"},
		SubmittedAt:      "2026-10-09 12:00",
		Overall: service.ManagementTraitsTestReportOverall{
			Score: "50.00", Level: "合格", Diagnosis: "模板校验文案", Advice: []string{"校验建议一", "校验建议二", "校验建议三"}, ChartScore: "50.000000000000",
		},
		Dimensions: make([]service.ManagementTraitsTestReportDimension, 0, 13),
		Modules:    make([]service.ManagementTraitsTestReportModule, 0, 4),
		Highest:    make([]service.ManagementTraitsTestReportSelection, 0, 3),
		Lowest:     make([]service.ManagementTraitsTestReportSelection, 0, 3),
	}
	for _, dimension := range service.ManagementTraitsDimensions() {
		data.Dimensions = append(data.Dimensions, service.ManagementTraitsTestReportDimension{Key: dimension.Key, Name: dimension.Name, Score: "50.00", Norm: dimension.Norm.FloatString(2), Level: "合格", Diagnosis: "模板校验评价", Advice: "模板校验建议", ChartScore: "50.000000000000", ChartNorm: dimension.Norm.FloatString(12)})
	}
	for _, key := range []string{"self", "interpersonal", "task", "development"} {
		data.Modules = append(data.Modules, service.ManagementTraitsTestReportModule{Key: key, Score: "50.00", ChartScore: "50.000000000000"})
	}
	for _, dimension := range data.Dimensions[:3] {
		selection := service.ManagementTraitsTestReportSelection{Key: dimension.Key, Name: dimension.Name, Text: "模板校验摘要"}
		data.Highest = append(data.Highest, selection)
		data.Lowest = append(data.Lowest, selection)
	}
	return data
}

func validateManagementTraitsTemplate(data []byte) (managementTraitsTemplateInfo, error) {
	info := managementTraitsTemplateInfo{FileName: managementTraitsSharedTemplateFileName, ProductCodes: []string{"00501", "00502"}}
	if len(data) == 0 || len(data) > maxManagementTraitsTemplateBytes {
		return info, errors.New("模板文件大小无效")
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return info, errors.New("DOCX文件结构无效")
	}
	for _, file := range reader.File {
		lowerName := strings.ToLower(file.Name)
		if strings.Contains(lowerName, "vbaproject") || strings.HasPrefix(lowerName, "word/embeddings/") {
			return info, errors.New("模板不得包含宏或嵌入对象")
		}
		if !strings.HasSuffix(lowerName, ".rels") {
			continue
		}
		rc, openErr := file.Open()
		if openErr != nil {
			return info, errors.New("读取DOCX关系失败")
		}
		body, readErr := io.ReadAll(io.LimitReader(rc, maxManagementTraitsTemplateBytes+1))
		closeErr := rc.Close()
		if readErr != nil || closeErr != nil || len(body) > maxManagementTraitsTemplateBytes {
			return info, errors.New("读取DOCX关系失败")
		}
		tree, parseErr := mngWordTree(body)
		if parseErr != nil {
			return info, errors.New("DOCX关系结构无效")
		}
		for _, relationship := range tree.all(opcRelationshipNS, "Relationship") {
			if strings.EqualFold(relationship.attr("TargetMode"), "External") {
				return info, fmt.Errorf("模板包含外部关系：%s", file.Name)
			}
		}
	}
	digest := sha256.Sum256(data)
	sha := hex.EncodeToString(digest[:])
	if _, err := renderManagementTraitsTestWordWithSHA(data, sha, managementTraitsTemplateProbeData()); err != nil {
		return info, errors.New("模板控件或图表契约不兼容")
	}
	info.SHA256 = sha
	info.Valid = true
	info.ContentControls = 90
	info.BusinessCharts = 6
	info.NumericLabels = 5
	return info, nil
}

func readManagementTraitsTemplateInfo(path string) (managementTraitsTemplateInfo, error) {
	info := managementTraitsTemplateInfo{FileName: managementTraitsSharedTemplateFileName, ProductCodes: []string{"00501", "00502"}}
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return info, nil
	}
	if err != nil || stat.IsDir() {
		return info, errors.New("读取005共用报告模板失败")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return info, errors.New("读取005共用报告模板失败")
	}
	validated, validationErr := validateManagementTraitsTemplate(data)
	info.Exists = true
	info.Size = stat.Size()
	info.ModTime = stat.ModTime().Format("2006-01-02 15:04:05")
	digest := sha256.Sum256(data)
	info.SHA256 = hex.EncodeToString(digest[:])
	if validationErr != nil {
		info.ValidationError = validationErr.Error()
		return info, nil
	}
	validated.Exists = true
	validated.Size = info.Size
	validated.ModTime = info.ModTime
	return validated, nil
}

func (h *ManagementTraitsRuntimeHandler) ManagementTraitsTemplateInfo(c *gin.Context) {
	if !managementTraitsRuntimeAdmin(c, "management-traits:report:manage-template") {
		return
	}
	path, err := h.managementTraitsTemplatePath()
	if err != nil {
		managementTraitsTemplateError(c, err.Error())
		return
	}
	h.templateMu.RLock()
	info, err := readManagementTraitsTemplateInfo(path)
	h.templateMu.RUnlock()
	if err != nil {
		managementTraitsTemplateError(c, err.Error())
		return
	}
	managementTraitsRuntimeRespond(c, info, nil)
}

func (h *ManagementTraitsRuntimeHandler) DownloadManagementTraitsTemplate(c *gin.Context) {
	if !managementTraitsRuntimeAdmin(c, "management-traits:report:manage-template") {
		return
	}
	path, err := h.managementTraitsTemplatePath()
	if err != nil {
		managementTraitsTemplateError(c, err.Error())
		return
	}
	h.templateMu.RLock()
	defer h.templateMu.RUnlock()
	file, err := os.Open(path)
	if err != nil {
		managementTraitsTemplateError(c, "005共用报告模板不存在")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		managementTraitsTemplateError(c, "读取005共用报告模板失败")
		return
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", "attachment; filename="+managementTraitsSharedTemplateFileName+"; filename*=UTF-8''"+encodeRFC5987FileName(managementTraitsSharedTemplateFileName))
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	_, _ = io.Copy(c.Writer, file)
}

func (h *ManagementTraitsRuntimeHandler) UploadManagementTraitsTemplate(c *gin.Context) {
	if !managementTraitsRuntimeAdmin(c, "management-traits:report:manage-template") {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxManagementTraitsTemplateBytes+(1<<20))
	fileHeader, err := c.FormFile("file")
	if err != nil {
		managementTraitsTemplateError(c, "请选择00501/00502共用报告模板")
		return
	}
	if !strings.EqualFold(filepath.Ext(fileHeader.Filename), ".docx") || fileHeader.Size <= 0 || fileHeader.Size > maxManagementTraitsTemplateBytes {
		managementTraitsTemplateError(c, "只支持20MB以内的.docx文件")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		managementTraitsTemplateError(c, "读取上传模板失败")
		return
	}
	data, readErr := io.ReadAll(io.LimitReader(file, maxManagementTraitsTemplateBytes+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil || len(data) == 0 || len(data) > maxManagementTraitsTemplateBytes {
		managementTraitsTemplateError(c, "读取上传模板失败")
		return
	}
	if _, err := validateManagementTraitsTemplate(data); err != nil {
		managementTraitsTemplateError(c, "模板校验失败："+err.Error())
		return
	}
	path, err := h.managementTraitsTemplatePath()
	if err != nil {
		managementTraitsTemplateError(c, err.Error())
		return
	}
	h.templateMu.Lock()
	backup, err := installPhase1WordTemplate(path, data, time.Now())
	var info managementTraitsTemplateInfo
	if err == nil {
		info, err = readManagementTraitsTemplateInfo(path)
	}
	h.templateMu.Unlock()
	if err != nil {
		managementTraitsTemplateError(c, "保存005共用报告模板失败")
		return
	}
	managementTraitsRuntimeRespond(c, gin.H{"template": info, "backupFile": filepath.Base(backup)}, nil)
}

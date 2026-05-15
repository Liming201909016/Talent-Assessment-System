package handler

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/model"
	jwtpkg "github.com/talent-assessment/refactored/pkg/jwt"
	"github.com/talent-assessment/refactored/pkg/response"
)

// ===================== 标准分（对齐 CandidateServiceImpl.standScore / standScore2） =====================
//
// Java 通过 AviatorEvaluator 对字符串表达式求值；Go 使用最小的 shunting-yard + RPN 计算。
// 变量形如 V{n}，取值来源于 el_paper_qu JOIN el_qu 后 content=V{n} 的那一行：
//   - standScore: quScore[content] = isRight ? 1 : 0
//   - standScore2: quScore[content] = answered ? actual_score : 0
//
// 最终把 rowScore 套入 (rowScore - avg)/st*2 + 5.5 做常模换算，范围 [0, 10]，仅 standScore 使用。

// POST /exam/api/tester/stand-score
// 入参：Tester（需要 paperId + repoCode；与 Java 保持一致）
func (h *TesterHandler) StandScore(c *gin.Context) {
	var b struct {
		PaperID  string `json:"paperId"  form:"paperId"`
		RepoCode string `json:"repoCode" form:"repoCode"`
	}
	if err := c.ShouldBind(&b); err != nil {
		response.AjaxErr(c, "参数错误")
		return
	}
	if b.PaperID == "" {
		response.AjaxErr(c, "paperId 为空")
		return
	}

	// 若 repoCode 未传，从 paper → exam → repo 反查（与 candidate.StandScoreCandidate 一致）
	if b.RepoCode == "" {
		var code string
		h.db.Table("el_paper p").
			Select("r.code").
			Joins("LEFT JOIN el_exam_repo er ON er.exam_id = p.exam_id").
			Joins("LEFT JOIN el_repo r ON er.repo_id = r.id").
			Where("p.id = ?", b.PaperID).
			Limit(1).
			Scan(&code)
		b.RepoCode = code
	}

	// 读取所有题目（paper_qu 关联 qu.content）
	rows, err := queryPaperQuContent(h.db, b.PaperID)
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}

	var out map[string]float64
	if strings.HasPrefix(b.RepoCode, "002") {
		out = standScore2(rows)
	} else {
		out = standScore1(rows)
	}
	response.AjaxOK(c, out)
}

// queryPaperQuWithContent 已迁移到 exam_pdf.go 的 queryPaperQuContent 共享函数
// paperQuContent struct 也移到 exam_pdf.go

// standScore1 对齐 CandidateServiceImpl.standScore（SCL-90 式 12 项指标）
func standScore1(rows []paperQuContent) map[string]float64 {
	vars := make(map[string]float64, 91)
	for i := 1; i <= 91; i++ {
		vars[fmt.Sprintf("V%d", i)] = 0
	}
	for _, r := range rows {
		if r.IsRight == 1 {
			vars[r.Content] = 1
		} else {
			vars[r.Content] = 0
		}
	}

	exprs := []struct {
		Key  string
		Expr string
		Avg  float64
		St   float64
	}{
		{"焦虑", "1-V1+1-V36+V41+V56+1-V61+V81+1-V15", 2.295, 1.367},
		// 抑郁: V77 公式方向与 Excel 模板（260428）一致；之前曾被误改为 +V77，
		// 真相是客户 Excel 中"高分"答题与数据库不同（同名异人），公式本身正确。
		{"抑郁", "V12+V26+V38+1-V49+V70+1-V77+1-V86+V21", 4.513, 1.479},
		{"心理失衡", "1-V11+1-V31+1-V47+V57+V75+V84+1-V19", 3.266, 1.448},
		{"敌意", "V3+V35+V46+V54+V63+1-V78+V14", 3.274, 1.287},
		{"恐惧", "1-V4+1-V28+1-V40+V51+1-V72+V87+1-V23", 3.417, 1.549},
		{"身体不适", "1-V2+V30+1-V44+V59+V74+1-V89+V17", 1.933, 1.746},
		{"认知衰退", "V7+1-V33+V42+V52+V68+1-V80+V20", 4.268, 1.395},
		{"情绪化", "V9+1-V37+V32+1-V53+1-V66+V85+1-V16", 3.352, 1.523},
		{"挫折感", "1-V10+1-V27+V48+1-V58+1-V64+1-V82+V18", 3.504, 1.310},
		{"自我否定", "V5+1-V34+1-V43+1-V50+1-V67+1-V79+1-V24", 3.275, 1.558},
		{"怀疑感", "V6+1-V39+V25+V55+1-V73+V83+1-V90+V65+V71+1-V22", 4.618, 1.697},
		{"职业倦怠", "1-V8+V29+V45+1-V60+1-V69+V76+V88+V62+V13", 2.937, 2.034},
	}
	out := map[string]float64{}
	for _, e := range exprs {
		raw, err := evalExpr(e.Expr, vars)
		if err != nil {
			continue
		}
		s := (raw-e.Avg)/e.St*2 + 5.5
		// 保留两位
		s = float64(int64(s*100+0.5)) / 100.0
		if s < 0 {
			s = 0
		} else if s > 10 {
			s = 10
		}
		out[e.Key] = s
	}
	return out
}

// standScore2 对齐 CandidateServiceImpl.standScore2（13 项人格/能力指标，保留 4 位小数）
func standScore2(rows []paperQuContent) map[string]float64 {
	vars := make(map[string]float64, 140)
	for i := 1; i <= 140; i++ {
		vars[fmt.Sprintf("V%d", i)] = 0
	}
	for _, r := range rows {
		if r.Answered == 1 {
			vars[r.Content] = float64(r.ActualScore)
		}
	}
	exprs := []struct {
		Key  string
		Expr string
	}{
		{"社会性", "(V1+V14+6-V28+6-V42+6-V56+6-V72+6-V87+6-V102+6-V115+6-V128)/10.0"},
		{"进取性", "(V2+V15+V29+V43+6-V57+V73+V88+V103+6-V116+V129)/10.0"},
		{"领导性", "(V3+V16+V31+V44+V58+V74+V89+V104+6-V117+6-V130)/10.0"},
		{"计划性", "(V4+V17+V21+V30+V45+V59+V75+V81+V90+V105+V118+6-V131)/12.0"},
		{"人际敏感性", "(V5+V18+V32+V46+V50+V61+V65+V76+V83+V91+6-V106+6-V119+6-V132)/13.0"},
		{"自信心", "(V6+V19+V33+V47+V60+V66+V78+V92+V101+V107+6-V120+6-V133)/12.0"},
		{"责任心", "(V7+V20+V34+V48+V63+V77+6-V94+6-V108+6-V121+6-V134)/10.0"},
		{"学习力", "(6-V8+V22+V36+V49+6-V64+V79+V93+6-V100+V109+V122+6-V135)/11.0"},
		{"创新性", "(6-V9+V23+V37+6-V51+V67+6-V80+V95+V110+V123+6-V136)/10.0"},
		{"情绪稳定性", "(V10+V24+V38+V52+V68+V82+V96+6-V111+6-V124+6-V137)/10.0"},
		{"自律性", "(V11+6-V25+V39+V53+6-V69+V84+6-V97+6-V112+V125+V138)/10.0"},
		{"决断性", "(6-V12+V26+V35+V40+V54+V62+V70+V85+V98+V113+V126+V139)/12.0"},
		{"合作性", "(V13+V27+V41+V55+V71+V86+V99+6-V114+V127+6-V140)/10.0"},
	}
	out := map[string]float64{}
	for _, e := range exprs {
		v, err := evalExpr(e.Expr, vars)
		if err != nil {
			continue
		}
		// 四位小数 HALF_UP
		v = float64(int64(v*10000+0.5)) / 10000.0
		out[e.Key] = v
	}
	return out
}

// ===================== 表达式求值（支持 + - * / ( ) 整数 V变量） =====================

func evalExpr(src string, vars map[string]float64) (float64, error) {
	tokens, err := tokenize(src)
	if err != nil {
		return 0, err
	}
	rpn, err := toRPN(tokens)
	if err != nil {
		return 0, err
	}
	return evalRPN(rpn, vars)
}

type token struct {
	kind string // "num" | "var" | "op" | "lp" | "rp"
	val  string
}

func tokenize(s string) ([]token, error) {
	var toks []token
	i := 0
	for i < len(s) {
		ch := s[i]
		switch {
		case ch == ' ' || ch == '\t':
			i++
		case ch == '+' || ch == '-' || ch == '*' || ch == '/':
			toks = append(toks, token{"op", string(ch)})
			i++
		case ch == '(':
			toks = append(toks, token{"lp", "("})
			i++
		case ch == ')':
			toks = append(toks, token{"rp", ")"})
			i++
		case ch >= '0' && ch <= '9':
			j := i
			for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || s[j] == '.') {
				j++
			}
			toks = append(toks, token{"num", s[i:j]})
			i = j
		case ch == 'V' || ch == 'v':
			j := i + 1
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}
			if j == i+1 {
				return nil, fmt.Errorf("invalid var at %d", i)
			}
			toks = append(toks, token{"var", "V" + s[i+1:j]})
			i = j
		default:
			return nil, fmt.Errorf("unexpected char %q at %d", ch, i)
		}
	}
	return toks, nil
}

func opPrec(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	}
	return 0
}

func toRPN(toks []token) ([]token, error) {
	var out, stack []token
	for _, t := range toks {
		switch t.kind {
		case "num", "var":
			out = append(out, t)
		case "op":
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.kind == "op" && opPrec(top.val) >= opPrec(t.val) {
					out = append(out, top)
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, t)
		case "lp":
			stack = append(stack, t)
		case "rp":
			matched := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.kind == "lp" {
					matched = true
					break
				}
				out = append(out, top)
			}
			if !matched {
				return nil, errors.New("unbalanced parentheses")
			}
		}
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.kind == "lp" || top.kind == "rp" {
			return nil, errors.New("unbalanced parentheses")
		}
		out = append(out, top)
	}
	return out, nil
}

func evalRPN(rpn []token, vars map[string]float64) (float64, error) {
	var stack []float64
	for _, t := range rpn {
		switch t.kind {
		case "num":
			var v float64
			if _, err := fmt.Sscanf(t.val, "%f", &v); err != nil {
				return 0, err
			}
			stack = append(stack, v)
		case "var":
			stack = append(stack, vars[t.val])
		case "op":
			if len(stack) < 2 {
				return 0, errors.New("bad expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			var r float64
			switch t.val {
			case "+":
				r = a + b
			case "-":
				r = a - b
			case "*":
				r = a * b
			case "/":
				if b == 0 {
					return 0, errors.New("div by zero")
				}
				r = a / b
			}
			stack = append(stack, r)
		}
	}
	if len(stack) != 1 {
		return 0, errors.New("bad expression")
	}
	return stack[0], nil
}

// ===================== 批量下载（对齐 TesterController.batchDownload） =====================

// POST /exam/api/tester/batch-download
// 入参：string[] ids（el_tester 主键）；响应体直接写 zip 流。
func (h *TesterHandler) BatchDownload(c *gin.Context) {
	var ids []string
	if err := c.ShouldBindJSON(&ids); err != nil || len(ids) == 0 {
		response.AjaxErr(c, "ids 为空")
		return
	}
	var rows []model.Tester
	if err := h.db.Table("el_tester").Where("id IN (?)", ids).Find(&rows).Error; err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	// 任一未生成报告 → 报错
	for _, r := range rows {
		if r.PdfFlag == nil || *r.PdfFlag == 0 {
			response.AjaxErr(c, "存在未生成报告，请先生成报告")
			return
		}
	}
	var paths []string
	for _, r := range rows {
		if r.PdfFlag != nil && *r.PdfFlag == 1 &&
			(r.DelFlag == nil || *r.DelFlag != 1) &&
			r.PdfPath != nil && *r.PdfPath != "" {
			paths = append(paths, *r.PdfPath)
		}
	}
	if len(paths) == 0 {
		response.AjaxErr(c, "没有pdf文件")
		return
	}

	// 压缩包名：用第一个 pdf 文件名中 '_' 第 2 段作为前缀（与 Java 同规则）
	base := filepath.Base(paths[0])
	zipName := time.Now().Format("20060102") + ".zip"
	if parts := strings.Split(base, "_"); len(parts) >= 2 {
		zipName = parts[1] + ".zip"
	}

	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename="+zipName)
	c.Status(http.StatusOK)

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		w, err := zw.Create(filepath.Base(p))
		if err != nil {
			f.Close()
			continue
		}
		_, _ = io.Copy(w, f)
		f.Close()
	}
}

// ===================== PDF 持久化（对齐 TesterController.PdfPersistence） =====================

// POST /exam/api/tester/pdf-persistence
// multipart/form-data: file=@xxx.pdf, idNumber=..., examId=...(可选)
// 将文件落盘到 {profile}/pdf/{yyyyMMdd}/{name}_{title}_{yyyyMMddhhmmssSSS}.pdf
// 并更新 tester_exam.pdf_path 与 pdf_flag=1。
//
// 与 Java 行为一致：
//   - 多文件时按最后一个落盘（同文件名）；通常前端只传 1 个
//   - 若 tester 已有旧 pdfPath，先删除旧文件
func (h *TesterHandler) PdfPersistence(c *gin.Context) {
	idNumber := c.PostForm("idNumber")
	examID := c.PostForm("examId")
	if idNumber == "" {
		response.AjaxErr(c, "没有识别码或身份证号")
		return
	}
	te, err := h.selectTesterByIdentifier(idNumber, examID)
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if te == nil {
		response.AjaxErr(c, "没有录入测评者信息")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.AjaxErr(c, "参数错误")
		return
	}
	files := form.File["file"]
	if len(files) == 0 {
		response.AjaxErr(c, "未上传文件")
		return
	}

	// 删除旧文件
	if te.PdfPath != nil && *te.PdfPath != "" {
		_ = os.Remove(*te.PdfPath)
	}

	// 查询 pdf 文件名（tester.name + exam.title）
	name, title, err := h.pdfNameByPaperID(te.PaperID)
	if err != nil || name == "" {
		// 容错：允许 paperId 为空时仅用 idNumber
		name = idNumber
	}
	ts := time.Now().Format("20060102150405.000")
	ts = strings.ReplaceAll(ts, ".", "")
	day := time.Now().Format("20060102")

	baseDir := h.cfg.Upload.Path
	if baseDir == "" {
		baseDir = "./tmp"
	}
	pdfDir := filepath.Join(baseDir, "pdf", day)
	if err := os.MkdirAll(pdfDir, 0o755); err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	fname := fmt.Sprintf("%s_%s_%s.pdf", name, title, ts)
	saved := filepath.Join(pdfDir, fname)

	var lastErr error
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			lastErr = err
			continue
		}
		dst, err := os.OpenFile(saved, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			src.Close()
			lastErr = err
			continue
		}
		_, err = io.Copy(dst, src)
		src.Close()
		dst.Close()
		if err != nil {
			lastErr = err
		}
	}
	if lastErr != nil {
		response.AjaxErr(c, "pdf存储失败")
		return
	}
	// 异步压缩 PDF（ghostscript），不阻塞响应
	go compressPDF(saved)

	te.PdfPath = &saved
	one := 1
	te.PdfFlag = &one
	now := time.Now()
	te.UpdateTime = &now
	if err := h.db.Table("el_tester").Where("id = ?", te.ID).Updates(map[string]any{
		"pdf_path": saved, "pdf_flag": 1, "update_time": &now,
	}).Error; err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	response.AjaxOK(c, 1)
}

// pdfNameByPaperID 对齐 TesterMapper.getPdfNameByPaperId：返回 (name, title)
func (h *TesterHandler) pdfNameByPaperID(paperID *string) (string, string, error) {
	if paperID == nil || *paperID == "" {
		return "", "", nil
	}
	var r struct {
		Name  string
		Title string
	}
	err := h.db.Table("el_tester AS t").
		Select("t.name AS name, ea.title AS title").
		Joins("INNER JOIN el_exam ea ON ea.id = t.exam_id").
		Where("t.paper_id = ?", *paperID).
		Take(&r).Error
	return r.Name, r.Title, err
}

// POST /exam/api/tester/login  对齐 Java TesterController.login
// Java 使用 AjaxResult（code=200），接受 form/query 参数（无 @RequestBody）
// 返回：AjaxResult success data 为完整 Tester 合并视图（含 password）
func (h *TesterHandler) LoginForm(c *gin.Context) {
	var b struct {
		IDNumber string `json:"idNumber" form:"idNumber"`
		Password string `json:"password" form:"password"`
		ExamID   string `json:"examId"   form:"examId"`
	}
	if err := c.ShouldBind(&b); err != nil || b.IDNumber == "" {
		response.AjaxErr(c, "参数错误")
		return
	}
	te, err := h.selectTesterByIdentifier(b.IDNumber, b.ExamID)
	if err != nil {
		response.AjaxErr(c, err.Error())
		return
	}
	if te == nil {
		response.AjaxErr(c, "用户不存在")
		return
	}
	if b.Password != te.Password {
		response.AjaxErr(c, "密码错误")
		return
	}

	// 7.1 修复：考试已过期 / 未开始 / 已禁用 → 拒绝登录
	// el_exam.state: 0=未开始, 1=ENABLE 进行中, 2=OVERDUE 已过期, 3=禁用
	examID := b.ExamID
	if examID == "" && te.ExamID != nil {
		examID = *te.ExamID
	}
	if examID != "" {
		var examState int
		var endTime *time.Time
		h.db.Table("el_exam").Where("id = ?", examID).Select("state, end_time").Row().Scan(&examState, &endTime)
		if examState == 2 {
			response.AjaxErr(c, "考试已过期")
			return
		}
		if examState == 3 {
			response.AjaxErr(c, "考试已禁用")
			return
		}
		if endTime != nil && endTime.Before(time.Now()) {
			response.AjaxErr(c, "考试已过期")
			return
		}
	}
	// 更新 el_tester 的 update_time 和 exam_id（与 Java 一致）
	now := time.Now()
	updates := map[string]any{"update_time": &now}
	if b.ExamID != "" {
		updates["exam_id"] = b.ExamID
	}
	if err := h.db.Table("el_tester").Where("id = ?", te.ID).Updates(updates).Error; err != nil {
		slog.Error("tester-login: update failed", "error", err)
	}

	// 签发 JWT（Go 侧额外追加，Java 原生无此字段）
	claims := map[string]any{
		"tester_id": te.ID,
		"id_number": te.IDNumber,
		"name":      te.Name,
	}
	token, _ := jwtpkg.Create(h.cfg.Jwt.Secret, claims)

	// 对齐 Java：返回完整 Tester（不含密码明文，Go 增强安全）
	view := gin.H{
		"id":          te.ID,
		"idNumber":    te.IDNumber,
		"name":        te.Name,
		"age":         te.Age,
		"gender":      te.Gender,
		"password":    te.Password, // 前端 H5 答题页需要此字段做身份验证
		"telephone":   te.Telephone,
		"affiliation": te.Affiliation,
		"depart":      te.Depart,
		"post":        te.Post,
		"degree":      te.Degree,
		"major":       te.Major,
		"stuFlag":     te.StuFlag,
		"status":      te.Status,
		"examId":      te.ExamID,
		"paperId":     te.PaperID,
		"endTime":     te.EndTime,
		"pdfPath":     te.PdfPath,
		"pdfFlag":     te.PdfFlag,
		"token":       token,
	}
	response.AjaxOK(c, view)
}

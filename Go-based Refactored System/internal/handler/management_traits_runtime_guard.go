package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/service"
	"gorm.io/gorm"
)

type managementTraitsLegacyRoute struct {
	method, path, kind string
	identity           bool
}

// Exact method/path inventory, not repo-prefix or source-question protection.
// Read endpoints that expose mutable legacy scores, credentials or PDF pointers
// are closed for new scope rather than silently recalculating/falling back.
var managementTraitsLegacyRoutes = []managementTraitsLegacyRoute{
	{"POST", "/exam/api/exam/exam/save", "exam", false},
	{"POST", "/exam/api/exam/exam/state", "exam", false},
	{"POST", "/exam/api/exam/exam/delete", "exam", false},
	{"POST", "/exam/api/exam/exam/pdf-team", "exam", false},
	{"POST", "/exam/api/exam/exam/pdf-upload", "exam", false},
	{"POST", "/exam/api/exam/exam/generate-report", "paper", false},
	{"POST", "/exam/api/exam/exam/export-raw-data", "exam", false},
	{"GET", "/exam/api/exam/exam/export-raw-data", "exam", false},
	{"POST", "/exam/api/exam/exam/export-raw-answers", "exam", false},
	{"GET", "/exam/api/exam/exam/export-raw-answers", "exam", false},
	{"GET", "/exam/api/exam/exam/answer-detail", "paper", false},
	{"POST", "/exam/api/paper/paper/save", "paper", false},
	{"POST", "/exam/api/paper/paper/delete", "paper", false},
	{"POST", "/exam/api/paper/paper/create-paper", "paper", false},
	{"POST", "/exam/api/paper/paper/paging", "collection", false},
	{"POST", "/exam/api/paper/paper/detail", "paper", false},
	{"POST", "/exam/api/paper/paper/paper-detail", "paper", false},
	{"POST", "/exam/api/paper/paper/paperQu-detail", "paper", false},
	{"POST", "/exam/api/paper/paper/qu-detail", "paper", false},
	{"POST", "/exam/api/paper/paper/fill-answer", "paper", false},
	{"POST", "/exam/api/paper/paper/hand-exam", "paper", false},
	{"POST", "/exam/api/paper/paper/paper-result", "paper", false},
	{"POST", "/exam/api/paper/paper/training", "paper", false},
	{"POST", "/exam/api/paper/paper/show_pdf", "paper", false},
	{"POST", "/exam/api/paper/paper/review-paper", "paper", false},
	{"POST", "/exam/api/paper/paper/stand-score", "paper", false},
	{"POST", "/exam/api/candidate/save", "candidate", true},
	{"POST", "/exam/api/candidate/update", "candidate", false},
	{"PUT", "/exam/api/candidate", "candidate", false},
	{"DELETE", "/exam/api/candidate/:ids", "candidate", false},
	{"DELETE", "/exam/api/candidate/logistic/:ids", "candidate", false},
	{"DELETE", "/exam/api/candidate/logicDeletePdfByIds/:ids", "both", false},
	{"PUT", "/exam/api/candidate/logicDeletePdfByIds/:ids", "both", false},
	{"POST", "/exam/api/candidate/end-time", "candidate", false},
	{"POST", "/exam/api/candidate/pdf-persistence", "candidate", false},
	{"POST", "/exam/api/candidate/pdf-upload", "both", false},
	{"POST", "/exam/api/candidate/info", "candidate", false},
	{"POST", "/exam/api/candidate/tester-info", "candidate", false},
	{"POST", "/exam/api/candidate/tester-list", "collection", false},
	{"POST", "/exam/api/candidate/stand-score", "candidate", false},
	{"GET", "/exam/api/candidate/team-score", "exam", false},
	{"POST", "/exam/api/candidate/batch-download", "candidate", false},
	{"POST", "/exam/api/tester/login", "tester", true},
	{"POST", "/exam/api/tester", "tester", false},
	{"PUT", "/exam/api/tester", "tester", false},
	{"DELETE", "/exam/api/tester/:ids", "tester", false},
	{"DELETE", "/exam/api/tester/logistic/:ids", "tester", false},
	{"POST", "/exam/api/tester/end-time", "tester", false},
	{"POST", "/exam/api/tester/pdf-persistence", "tester", false},
	{"POST", "/exam/api/tester/importData", "tester", false},
	{"POST", "/exam/api/tester/stand-score", "tester", false},
	{"POST", "/exam/api/tester/batch-download", "tester", false},
	{"POST", "/exam/api/tester/export", "exam", false},
	{"GET", "/exam/api/tester", "collection", false},
	{"GET", "/exam/api/tester/list", "collection", false},
	{"GET", "/exam/api/tester/tester-list", "collection", false},
	{"GET", "/exam/api/tester/:id", "tester", false},
	{"GET", "/exam/api/tester/idNumber/:idNumber", "tester", false},
	{"POST", "/exam/api/mbti/paper-detail", "paper", false},
	{"POST", "/exam/api/mbti/fill-answer", "paper", false},
	{"POST", "/exam/api/mbti/submit", "paper", false},
	{"POST", "/exam/api/mbti/score", "paper", false},
	{"POST", "/exam/api/mbti/generate-report", "paper", false},
	{"POST", "/exam/api/mbti/download-report", "paper", false},
	{"POST", "/exam/api/mbti/batch-download-simple", "paper", false},
}

func managementTraitsLegacyMatch(method, path string) (managementTraitsLegacyRoute, bool) {
	for _, route := range managementTraitsLegacyRoutes {
		if route.method != method {
			continue
		}
		a, b := strings.Split(route.path, "/"), strings.Split(path, "/")
		if len(a) != len(b) {
			continue
		}
		matched := true
		for i := range a {
			if strings.HasPrefix(a[i], ":") {
				matched = matched && b[i] != ""
			} else {
				matched = matched && a[i] == b[i]
			}
		}
		if matched {
			return route, true
		}
	}
	return managementTraitsLegacyRoute{}, false
}

func managementTraitsLegacyWrites(route managementTraitsLegacyRoute) bool {
	if route.kind == "collection" {
		return false
	}
	if route.method == "GET" {
		return false
	}
	switch route.path {
	case "/exam/api/exam/exam/pdf-upload", "/exam/api/exam/exam/export-raw-data", "/exam/api/exam/exam/export-raw-answers",
		"/exam/api/paper/paper/detail", "/exam/api/paper/paper/paper-detail", "/exam/api/paper/paper/paperQu-detail", "/exam/api/paper/paper/qu-detail",
		"/exam/api/paper/paper/paper-result", "/exam/api/paper/paper/training", "/exam/api/paper/paper/show_pdf", "/exam/api/paper/paper/stand-score",
		"/exam/api/candidate/pdf-upload", "/exam/api/candidate/info", "/exam/api/candidate/tester-info", "/exam/api/candidate/tester-list", "/exam/api/candidate/stand-score", "/exam/api/candidate/batch-download",
		"/exam/api/tester/stand-score", "/exam/api/tester/batch-download", "/exam/api/tester/export",
		"/exam/api/mbti/paper-detail", "/exam/api/mbti/score", "/exam/api/mbti/download-report":
		return false
	}
	return true
}

// Strict only for scope identifiers; the existing nonnew DTO remains unchanged.
// Reject duplicate/case aliases instead of inspecting one ID and writing another.
func managementTraitsLegacyJSON(raw []byte) (map[string]json.RawMessage, []string, error) {
	if !utf8.Valid(raw) {
		return nil, nil, errors.New("invalid scope body")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	tok, err := d.Token()
	if err != nil {
		return nil, nil, err
	}
	if tok == json.Delim('[') {
		ids := make([]string, 0)
		for d.More() {
			var id string
			if err := d.Decode(&id); err != nil {
				return nil, nil, err
			}
			ids = append(ids, id)
		}
		if _, err := d.Token(); err != nil {
			return nil, nil, err
		}
		if _, err := d.Token(); err != io.EOF {
			return nil, nil, errors.New("trailing body")
		}
		return nil, ids, nil
	}
	if tok != json.Delim('{') {
		return nil, nil, errors.New("scope object required")
	}
	fields := make(map[string]json.RawMessage)
	for d.More() {
		tok, err = d.Token()
		if err != nil {
			return nil, nil, err
		}
		name, ok := tok.(string)
		if !ok {
			return nil, nil, errors.New("scope key required")
		}
		key := strings.ToLower(name)
		if _, seen := fields[key]; seen {
			return nil, nil, errors.New("duplicate scope key")
		}
		var v json.RawMessage
		if err := d.Decode(&v); err != nil {
			return nil, nil, err
		}
		fields[key] = v
	}
	if _, err := d.Token(); err != nil {
		return nil, nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, nil, errors.New("trailing body")
	}
	return fields, nil, nil
}

func managementTraitsLegacyTargets(c *gin.Context, route managementTraitsLegacyRoute) (service.ManagementTraitsLegacyScopeRequest, error) {
	var out service.ManagementTraitsLegacyScopeRequest
	// Collection filters must follow the actual binder, not unused query/body
	// aliases that would disguise an unfiltered legacy response as scoped.
	collectionExam := ""
	if route.kind == "collection" && c.Request.Method == "GET" {
		collectionExam = c.Query("examId")
	}
	values := make(map[string][]string)
	add := func(key string, vals ...string) {
		key = strings.ToLower(key)
		for _, v := range vals {
			if v != "" {
				values[key] = append(values[key], v)
			}
		}
	}
	for k, v := range c.Request.URL.Query() {
		add(k, v...)
	}
	if c.Request.Body != nil {
		// Bounded read and byte-for-byte restoration, including multipart files.
		raw, err := io.ReadAll(io.LimitReader(c.Request.Body, (64<<20)+1))
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))
		if err != nil || len(raw) > 64<<20 {
			return out, errors.New("scope body too large")
		}
		defer func() { c.Request.Body = io.NopCloser(bytes.NewReader(raw)) }()
		media, _, _ := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if len(bytes.TrimSpace(raw)) > 0 {
			switch media {
			case "application/x-www-form-urlencoded":
				form, err := url.ParseQuery(string(raw))
				if err != nil {
					return out, err
				}
				for k, v := range form {
					add(k, v...)
				}
			case "multipart/form-data":
				if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
					return out, err
				}
				for k, v := range c.Request.MultipartForm.Value {
					add(k, v...)
				}
			default:
				fields, ids, err := managementTraitsLegacyJSON(raw)
				if err != nil {
					return out, err
				}
				add("ids", ids...)
				if route.kind == "collection" && c.Request.Method == "POST" {
					filter := fields
					if route.path == "/exam/api/paper/paper/paging" {
						filter = nil
						if v, ok := fields["params"]; ok && string(v) != "null" {
							var err error
							filter, _, err = managementTraitsLegacyJSON(v)
							if err != nil {
								return out, err
							}
						}
					}
					if v, ok := filter["examid"]; ok && string(v) != "null" {
						if json.Unmarshal(v, &collectionExam) != nil {
							return out, errors.New("invalid collection scope")
						}
					}
				}
				for _, key := range []string{"id", "examid", "paperid", "idnumber", "telephone", "file", "changeexamid"} {
					if v, ok := fields[key]; ok {
						if string(v) == "null" {
							continue
						}
						var s string
						if json.Unmarshal(v, &s) != nil {
							return out, errors.New("invalid scope identifier")
						}
						add(key, s)
					}
				}
				if v, ok := fields["ids"]; ok {
					var ids []string
					if string(v) == "null" || json.Unmarshal(v, &ids) != nil {
						return out, errors.New("invalid scope ids")
					}
					for _, id := range ids {
						if id == "" {
							return out, errors.New("empty scope id")
						}
					}
					add("ids", ids...)
				}
				if route.path == "/exam/api/paper/paper/review-paper" {
					if v, ok := fields["qulist"]; ok {
						var questions []json.RawMessage
						if json.Unmarshal(v, &questions) != nil {
							return out, errors.New("invalid question IDs")
						}
						for _, question := range questions {
							nested, _, err := managementTraitsLegacyJSON(question)
							if err != nil {
								return out, err
							}
							v, ok := nested["id"]
							if !ok {
								return out, errors.New("missing question ID")
							}
							var id string
							if string(v) == "null" || json.Unmarshal(v, &id) != nil || id == "" {
								return out, errors.New("invalid question ID")
							}
							out.PaperQuestionIDs = append(out.PaperQuestionIDs, id)
						}
					}
				}
			}
		}
	}
	pattern, path := strings.Split(route.path, "/"), strings.Split(c.Request.URL.Path, "/")
	if len(pattern) == len(path) {
		for i, p := range pattern {
			if strings.HasPrefix(p, ":") {
				add(p[1:], path[i])
			}
		}
	}
	out.ExamIDs = append(out.ExamIDs, values["examid"]...)
	out.ExamIDs = append(out.ExamIDs, values["changeexamid"]...)
	if route.kind == "collection" {
		out.AllLegacy = collectionExam == ""
		if collectionExam != "" {
			out.ExamIDs = append(out.ExamIDs, collectionExam)
		}
	}
	out.PaperIDs = append(out.PaperIDs, values["paperid"]...)
	ids := append(append([]string{}, values["id"]...), values["ids"]...)
	expanded := make([]string, 0, len(ids))
	for _, id := range ids {
		expanded = append(expanded, splitCsv(id)...)
	}
	switch route.kind {
	case "exam":
		out.ExamIDs = append(out.ExamIDs, expanded...)
	case "paper":
		out.PaperIDs = append(out.PaperIDs, expanded...)
	case "candidate":
		out.CandidateIDs = append(out.CandidateIDs, expanded...)
	case "tester":
		out.TesterIDs = append(out.TesterIDs, expanded...)
	case "both":
		out.CandidateIDs = append(out.CandidateIDs, expanded...)
		out.TesterIDs = append(out.TesterIDs, expanded...)
	}
	if route.kind == "candidate" {
		out.CandidateTelephones = append(out.CandidateTelephones, values["telephone"]...)
	}
	if route.kind == "tester" {
		out.TesterIdentifiers = append(out.TesterIdentifiers, values["idnumber"]...)
		out.TesterIDs = append(out.TesterIDs, values["idnumber"]...)
	}
	if strings.HasSuffix(route.path, "/pdf-upload") && route.kind == "both" {
		out.PDFPaths = append(out.PDFPaths, values["file"]...)
	}
	if len(values["examid"]) == 1 {
		out.IdentifierExamID = values["examid"][0]
	}
	return out, nil
}

// Install BEFORE JWT. Protected legacy writes are rejected regardless of JWT;
// only the two identity admission helpers may issue new participant credentials.
// The shared read lease covers probes, all handlers, response/file streaming and
// synchronous old compression. Freeze must hold its exclusive gate through tx.
func ManagementTraitsLegacyScopeGuard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		route, matched := managementTraitsLegacyMatch(c.Request.Method, c.Request.URL.Path)
		if !matched {
			c.Next()
			return
		}
		// Renderer children only read. Do not queue them behind a freeze that
		// already waits for their parent report writer's lease to finish.
		if managementTraitsLegacyWrites(route) {
			lease := service.LockManagementTraitsLegacyMutation()
			defer lease.Release()
			c.Set("managementTraitsLegacyLease", lease)
		}
		target, err := managementTraitsLegacyTargets(c, route)
		if c.Request.MultipartForm != nil {
			defer c.Request.MultipartForm.RemoveAll()
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误", "success": false})
			return
		}
		identity := route.identity && len(target.ExamIDs) > 0
		if route.path == "/exam/api/candidate/save" && len(target.ExamIDs) == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "examId 为空", "success": false})
			return
		}
		if identity {
			target = service.ManagementTraitsLegacyScopeRequest{}
		}
		prepared := false
		switch route.method + " " + route.path {
		case "POST /exam/api/exam/exam/save", "POST /exam/api/tester", "PUT /exam/api/tester", "POST /exam/api/tester/importData", "GET /exam/api/tester/list", "GET /exam/api/tester":
			// The exception never applies to unfiltered collections, paper IDs,
			// candidate writes, credential login, or frozen profiles.
			exactExam := len(target.ExamIDs) > 0
			for _, id := range target.ExamIDs {
				exactExam = exactExam && id == target.ExamIDs[0]
			}
			target.DraftPreparation = !target.AllLegacy && exactExam && len(target.PaperIDs) == 0
			target.DraftPreparationUsed = &prepared
		}
		protected, err := service.CheckManagementTraitsLegacyScope(c.Request.Context(), db, target)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"code": 1, "msg": "管理特质保护校验失败，请稍后重试", "success": false})
			return
		}
		if protected {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 1, "msg": "管理特质新版实体禁止通过旧接口写入或读取，请使用专用入口", "success": false})
			return
		}
		if prepared {
			c.Set("managementTraitsDraftPreparation", true)
		}
		c.Next()
	}
}

// Runs AFTER normal JWT, not after the handler. Role-only is not authorization.
func ManagementTraitsDraftPreparationGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if value, exists := c.Get("managementTraitsDraftPreparation"); exists && value == true && !managementTraitsRuntimeAdmin(c, "management-traits:draft:prepare") {
			return
		}
		c.Next()
	}
}

// Acquire the child ownership BEFORE returning the parent HTTP response. Fork
// extends an existing request lease even when an exclusive freeze is queued.
func managementTraitsLegacyBackground(c *gin.Context, work func()) {
	var release func()
	if value, ok := c.Get("managementTraitsLegacyLease"); ok {
		if lease, ok := value.(*service.ManagementTraitsLegacyLease); ok {
			release = lease.Fork()
		}
	}
	if release == nil {
		lease := service.LockManagementTraitsLegacyMutation()
		release = lease.Release
	}
	go func() { defer release(); work() }()
}

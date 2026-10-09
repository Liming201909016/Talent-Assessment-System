package handler

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talent-assessment/refactored/internal/service"
)

func TestManagementTraitsWriteExpiredCredentialHTTPMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, wrapped := range []bool{false, true} {
		err := service.ErrManagementTraitsRuntimeToken
		if wrapped {
			err = fmt.Errorf("private transaction context: %w", err)
		}
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		managementTraitsRuntimeRespond(c, gin.H{"answered": 140}, err)
		resumeHandlerStatus(t, w, 401, "private transaction context", err.Error())
		if !strings.Contains(w.Body.String(), "管理特质身份验证失败") {
			t.Fatal("authentication message changed")
		}
	}
}

// The real Gin participant parser admits a signed credential, then a real GORM
// locking query waits until it expires. The response must be 401 with rollback,
// not generic 409, a successful retry or an internal-error disclosure.
func TestBugManagementTraitsWriteExpiryRealHTTPTransaction(t *testing.T) {
	for _, operation := range []string{"fill-answer", "submit"} {
		t.Run(operation, func(t *testing.T) {
			v := resumeHandlerFixtureNew(t, 20*time.Minute)
			db, mock := identityHandlerDB(t)
			r, cfg, _ := resumeHandlerRouter(t, db)
			resumeHandlerRuntimeSchema(t, mock)
			expires := time.Now().Add(2 * time.Second).Unix()
			claims := service.ManagementTraitsRuntimeClaims{Purpose: service.ManagementTraitsRuntimePaperPurpose, ParticipantType: "candidate", ParticipantID: "existing", ExamID: "exam", PaperID: "paper", ExpiresAt: expires}
			token, err := service.CreateManagementTraitsRuntimeToken(cfg.Jwt.Secret, claims, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			// Delay in the driver, not callbacks on the original GORM registry:
			// the runtime deliberately owns an isolated registry on the same pool.
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_paper.*FOR UPDATE").WithArgs("paper", 1).
				WillDelayFor(time.Until(time.Unix(expires, 0)) + 100*time.Millisecond).
				WillReturnRows(identityHandlerModelRows(t, v.paper))
			mock.ExpectQuery("SELECT .*el_mng_paper_snapshot").WithArgs("paper", 1).WillReturnRows(identityHandlerModelRows(t, v.snapshot))
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle").WithArgs("bundle", 1).WillReturnRows(identityHandlerModelRows(t, v.bundle))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile").WithArgs("exam", 1).WillReturnRows(identityHandlerModelRows(t, v.profile))
			mock.ExpectQuery("SELECT 'candidate'.*UNION ALL SELECT 'tester'").WithArgs("paper", "paper").WillReturnRows(resumeHandlerOwnerRows(v))
			mock.ExpectQuery("SELECT .*el_mng_paper_question_snapshot").WithArgs("paper", 141).WillReturnRows(identityHandlerModelRows(t, v.questions...))
			mock.ExpectQuery("SELECT .*el_paper_qu").WithArgs("paper", 141).WillReturnRows(identityHandlerModelRows(t, v.legacy...))
			mock.ExpectQuery("SELECT .*el_mng_definition_bundle.*LOCK IN SHARE MODE").WithArgs("bundle").WillReturnRows(identityHandlerModelRows(t, v.bundle))
			mock.ExpectRollback()
			body := `{"paperId":"paper","submitType":"manual"}`
			if operation == "fill-answer" {
				body = `{"paperId":"paper","paperQuestionId":"pq1","optionId":"o1-3"}`
			}
			w := resumeHandlerRequest(r, "/participant/"+operation, body, "", token)
			resumeHandlerStatus(t, w, 401, token, cfg.Jwt.Secret)
			resumeHandlerSQLDone(t, mock)
		})
	}
}

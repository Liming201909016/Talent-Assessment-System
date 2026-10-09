package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

func TestManagementTraits005SourceIsolation(t *testing.T) {
	for _, pair := range [][2]string{{"00201", "00501"}, {"00202", "00502"}} {
		t.Run(pair[1], func(t *testing.T) {
			rows := managementRuntimeSourceRows()
			old, mapping, err := BuildManagementTraitsRuntimeCurrentSource(pair[0], rows)
			if err != nil {
				t.Fatal(err)
			}
			mc, _ := CanonicalManagementTraitsManifest(old)
			pc, _ := CanonicalManagementTraitsMapping(old, mapping)
			t.Logf("historical %s manifest=%s mapping=%s", pair[0], mc.SHA256, pc.SHA256)
			golden := map[string][2]string{
				"00201": {"639f42caf091d6572b4e4b4df61af18d3295a4dc19b812d0e23f6cfcfe64619f", "f7b97a615d0134706285f3ef65edf0c4eb8a05993d151390ea5e8ebec2b6b3f6"},
				"00202": {"1e7128bf52fce6cdad541a653965b1149814afd7b76a820e7622d25a7591e88e", "569a16b609ddf0595cf3403f02f11500620a255f71c294b1994a079689a9c936"},
			}
			if mc.SHA256 != golden[pair[0]][0] || pc.SHA256 != golden[pair[0]][1] {
				t.Fatal("historical canonical bytes changed")
			}
			for i := range rows {
				rows[i].RelationID = pair[1] + "-" + rows[i].RelationID
				rows[i].QuestionID = pair[1] + "-" + rows[i].QuestionID
				rows[i].OptionID = pair[1] + "-" + rows[i].OptionID
			}
			newManifest, newMapping, err := BuildManagementTraitsRuntimeCurrentSource(pair[1], rows)
			if err != nil {
				t.Fatalf("independent 005 source rejected: %v", err)
			}
			if newManifest.Questionnaire != old.Questionnaire || newManifest.Versions.Product == old.Versions.Product || newManifest.Versions.Question == old.Versions.Question {
				t.Fatal("005 audience or namespace not independent")
			}
			if validateManagementTraitsRuntimeVersions(newManifest.Questionnaire, newManifest.Versions) != nil {
				t.Fatal("005 runtime policy rejected")
			}
			nc, _ := CanonicalManagementTraitsManifest(newManifest)
			np, _ := CanonicalManagementTraitsMapping(newManifest, newMapping)
			if nc.SHA256 == mc.SHA256 || np.SHA256 == pc.SHA256 {
				t.Fatal("005 reused historical hashes")
			}
			if len(newMapping.Questions) != 140 || newMapping.Questions[0].SourceQuestionID != pair[1]+"-q1" {
				t.Fatal("source IDs were substituted")
			}
			if validateManagementTraitsRuntimeVersions(old.Questionnaire, newManifest.Versions) != nil {
				t.Fatal("same approved audience rejected")
			}
			opposite := ManagementTraitsQuestionnaireLeader
			if old.Questionnaire == opposite {
				opposite = ManagementTraitsQuestionnaireStaff
			}
			if validateManagementTraitsRuntimeVersions(opposite, newManifest.Versions) == nil {
				t.Fatal("cross audience accepted")
			}
		})
	}
}

func TestManagementTraits00502QuestionVersionPreservesFrozenV1(t *testing.T) {
	questionnaire, current, err := managementTraitsRuntimeVersions("00502")
	if err != nil || current.Question != "mng-00502-db-current-v2" {
		t.Fatalf("current 00502 question version=%q err=%v", current.Question, err)
	}
	legacy := current
	legacy.Question = "mng-00502-db-current-v1"
	if validateManagementTraitsRuntimeVersions(questionnaire, legacy) != nil {
		t.Fatal("already frozen 00502 v1 was invalidated")
	}
}

func TestManagementTraits005FreezeUnmarkedProductRejected(t *testing.T) {
	for _, code := range []string{"00201", "00202", "00501", "00502"} {
		t.Run(code, func(t *testing.T) {
			db, mock := managementRuntimeDB(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*el_exam.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, model.Exam{ID: "exam", AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25, RequiredFields: "name"}))
			mock.ExpectQuery("SELECT .*el_mng_exam_profile").WillReturnRows(sqlmock.NewRows([]string{"exam_id"}))
			mock.ExpectQuery("SELECT COUNT.*information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT count.*el_paper").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("SELECT .*el_exam_repo").WillReturnRows(managementRuntimeModelRows(t, model.ExamRepo{ID: "link", ExamID: "exam", RepoID: "synthetic-repo", RadioCount: 140}))
			mock.ExpectQuery("SELECT .*el_repo").WillReturnRows(managementRuntimeModelRows(t, model.Repo{ID: "synthetic-repo", Code: code, RadioCount: 140}))
			mock.ExpectRollback()
			s := NewManagementTraitsRuntimeService(db, "synthetic-only", 1<<20)
			if p, err := s.freezeProfileTransaction(context.Background(), "exam", time.Now); err == nil || p.ExamID != "" {
				t.Fatal("unmarked product upgraded")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

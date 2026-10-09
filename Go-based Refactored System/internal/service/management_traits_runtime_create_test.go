package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
	"gorm.io/gorm"
)

func TestManagementTraitsCreateFrozenPaperFourPairs(t *testing.T) {
	for _, code := range []string{"00201", "00202"} {
		for _, kind := range []string{"candidate", "tester"} {
			t.Run(code+"/"+kind, func(t *testing.T) {
				manifest, mapping, err := BuildManagementTraitsRuntimeCurrentSource(code, managementRuntimeSourceRows())
				if err != nil {
					t.Fatal(err)
				}
				mc, _ := CanonicalManagementTraitsManifest(manifest)
				pc, _ := CanonicalManagementTraitsMapping(manifest, mapping)
				now := time.Date(2026, 10, 3, 8, 0, 0, 0, time.UTC)
				frozen := now.Add(-time.Minute)
				fc, _ := managementTraitsCanonicalBytes(managementTraitsRuntimeFieldContract{Schema: "mng-candidate-fields-v1", RequiredFields: []string{"name", "telephone"}, TimePolicy: "candidate-only-personal-25-minutes-v1", Source: "current-source-not-client-leader-full-not-historical", RepoCode: code, CapturedAt: frozen, ManifestSHA: mc.SHA256, MappingSHA: pc.SHA256})
				bundle := model.ManagementTraitsDefinitionBundle{ID: "bundle", ProductVersion: manifest.Versions.Product, QuestionVersion: manifest.Versions.Question, ScoringVersion: manifest.Versions.Scoring, NormVersion: manifest.Versions.Norm, Questionnaire: manifest.Questionnaire, ScoringManifest: string(mc.JSON), ScoringManifestSHA: mc.SHA256, Status: "candidate-current-source", CreatedAt: frozen}
				profile := model.ManagementTraitsExamProfile{ExamID: "exam", BundleID: bundle.ID, MappingSnapshot: string(pc.JSON), MappingSHA: pc.SHA256, FieldContract: string(fc.JSON), TotalTimeMinutes: 25, FrozenAt: &frozen, CreatedAt: frozen}
				owner := managementTraitsRuntimeOwner{Kind: kind, ID: "owner", ExamID: "exam", Name: "frozen owner", Telephone: "13800000000", Status: "0"}
				exam := model.Exam{ID: "exam", State: 0, AssessmentType: "legacy", ScoringMode: "legacy", TotalTime: 25}
				order := make([]string, 140)
				for i, q := range mapping.Questions {
					order[139-i] = q.SourceQuestionID
				}
				r, err := buildManagementTraitsRuntimePaper(profile, bundle, exam, owner, "paper", now, order, 1<<20)
				if err != nil {
					t.Fatal(err)
				}
				if len(r.Questions) != 140 || len(r.LegacyQuestions) != 140 || len(r.Buckets) != 700 || !r.Snapshot.LimitTime.Equal(now.Add(25*time.Minute)) || r.Paper.State != 1 {
					t.Fatal("incomplete frozen paper")
				}
				if _, err := ValidateManagementTraitsStoredInput(bundle, r.Snapshot, r.Questions, 1<<20); err != nil {
					t.Fatal(err)
				}
				if r.Questions[0].Number != 140 || r.Questions[0].DisplayOrder != 1 || r.Questions[0].RawAnswer != nil || r.Questions[0].Content != mapping.Questions[139].Content {
					t.Fatal("V/source/order/empty answer changed")
				}
				for _, bad := range []func(*model.ManagementTraitsDefinitionBundle, *model.Exam, []string){
					func(b *model.ManagementTraitsDefinitionBundle, e *model.Exam, o []string) { b.Status = "retired" },
					func(b *model.ManagementTraitsDefinitionBundle, e *model.Exam, o []string) {
						b.Status = "review-revoked"
					},
					func(b *model.ManagementTraitsDefinitionBundle, e *model.Exam, o []string) { e.EndTime = &now },
					func(b *model.ManagementTraitsDefinitionBundle, e *model.Exam, o []string) { o[0] = o[1] },
				} {
					b, e, o := bundle, exam, append([]string{}, order...)
					bad(&b, &e, o)
					if _, err := buildManagementTraitsRuntimePaper(profile, b, e, owner, "paper", now, o, 1<<20); err == nil {
						t.Fatal("invalid new paper admitted")
					}
				}
			})
		}
	}
}

func TestManagementTraitsCreateWriterAtomicRollback(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `el_paper`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `el_paper_qu`").WillReturnError(errors.New("private snapshot failure"))
	mock.ExpectRollback()
	err := db.Transaction(func(tx *gorm.DB) error {
		return persistManagementTraitsRuntimePaper(tx, managementTraitsRuntimePaperRecords{Paper: model.Paper{ID: "paper"}, LegacyQuestions: []model.PaperQu{{ID: "pq"}}})
	})
	if err == nil {
		t.Fatal("failed write committed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestManagementTraitsPublicRuntimeSchemaBeforeTransaction(t *testing.T) {
	db, mock := managementRuntimeDB(t)
	mock.ExpectQuery("SELECT table_name AS table_name, engine AS engine FROM information_schema.tables").WillReturnRows(sqlmock.NewRows([]string{"table_name", "engine"}))
	s := NewManagementTraitsRuntimeService(db, "test", 1<<20)
	claims := ManagementTraitsRuntimeClaims{Purpose: ManagementTraitsRuntimeParticipantPurpose, ParticipantType: "candidate", ParticipantID: "owner", ExamID: "exam", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	if _, err := s.CreatePaper(context.Background(), claims); !errors.Is(err, ErrManagementTraitsRuntimeClosed) {
		t.Fatal("schema-absent create started", err)
	}
	if _, err := s.FreezeProfile(context.Background(), "exam"); !errors.Is(err, ErrManagementTraitsRuntimeClosed) {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

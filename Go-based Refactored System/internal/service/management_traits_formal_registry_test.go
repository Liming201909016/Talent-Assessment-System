package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/talent-assessment/refactored/internal/model"
)

func formalRegistryFixture(t *testing.T) (model.ManagementTraitsFormalVersion, []model.ManagementTraitsFormalApproval) {
	t.Helper()
	f, _, _, profile, _ := managementRuntimeLoadFixture(t, false)
	v := model.ManagementTraitsFormalVersion{ID: "89ff90b9-692a-47d4-a342-57d4b1a1a001", VersionCode: "local-v1", Environment: "local", RepoCode: "00201", BundleID: f.Bundle.ID, SourceSnapshot: formalSourceSnapshot(f.Bundle, profile), ContentSHA: strings.Repeat("a", 64), WorkbookSHA: managementTraitsTestWorkbookSHA, TemplateSHA: strings.Repeat("b", 64), BindingSHA: strings.Repeat("c", 64), AssetKey: "local-v1", State: "draft", Epoch: 1, CreatedBy: 7, CreatedAt: time.Now().Truncate(time.Microsecond)}
	v.IdentitySHA = formalVersionIdentity(v)
	a := make([]model.ManagementTraitsFormalApproval, 0, 2)
	for _, kind := range []string{"content", "psychometrics"} {
		a = append(a, model.ManagementTraitsFormalApproval{ID: kind, VersionID: v.ID, IdentitySHA: v.IdentitySHA, Kind: kind, ActorID: 7, CreatedAt: v.CreatedAt})
	}
	return v, a
}

func TestManagementTraitsFormalReadiness(t *testing.T) {
	v, a := formalRegistryFixture(t)
	for _, tc := range []struct {
		name      string
		mutate    func(*model.ManagementTraitsFormalVersion)
		approvals int
		assets    bool
		active    bool
		reason    string
	}{
		{"draft", func(*model.ManagementTraitsFormalVersion) {}, 0, true, false, "dual_approval_required"},
		{"single", func(*model.ManagementTraitsFormalVersion) {}, 1, true, false, "dual_approval_required"},
		{"same_actor_dual", func(*model.ManagementTraitsFormalVersion) {}, 2, true, false, "activation_required"},
		{"active", func(v *model.ManagementTraitsFormalVersion) { v.State = "active" }, 2, true, true, ""},
		{"missing_template", func(*model.ManagementTraitsFormalVersion) {}, 2, false, false, "formal_assets_not_ready"},
		{"revoked", func(v *model.ManagementTraitsFormalVersion) { v.State = "revoked" }, 2, true, false, "version_revoked"},
		{"production", func(v *model.ManagementTraitsFormalVersion) {
			v.Environment = "production"
			v.IdentitySHA = formalVersionIdentity(*v)
		}, 2, true, false, "environment_closed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x := v
			tc.mutate(&x)
			got := formalReadiness(x, a[:tc.approvals], tc.assets)
			if got.VersionReady != tc.active || got.BlockedReason != tc.reason || got.CanGenerate || got.GenerationBlockedReason != "formal_report_pipeline_not_installed" {
				t.Fatalf("readiness: %+v", got)
			}
			if got.Approvals == nil {
				t.Fatal("nil approvals")
			}
		})
	}
	bad := append([]model.ManagementTraitsFormalApproval{}, a...)
	bad[1].IdentitySHA = strings.Repeat("f", 64)
	if formalReadiness(v, bad, true).CanActivate {
		t.Fatal("foreign approval accepted")
	}
}

func TestManagementTraitsFormalApprovalTransaction(t *testing.T) {
	for _, kind := range []string{"content", "psychometrics"} {
		t.Run(kind, func(t *testing.T) {
			v, a := formalRegistryFixture(t)
			db, m := managementRuntimeDB(t)
			s := &ManagementTraitsFormalRegistry{db: db, environment: "local"}
			m.ExpectBegin()
			m.ExpectQuery("SELECT .*el_mng_formal_version.*FOR UPDATE").WithArgs(v.ID, 1).WillReturnRows(managementRuntimeModelRows(t, v))
			m.ExpectQuery("SELECT .*el_mng_formal_approval").WithArgs(v.ID, 3).WillReturnRows(sqlmock.NewRows([]string{"id"}))
			m.ExpectExec("INSERT INTO `el_mng_formal_approval`").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectExec("INSERT INTO `el_mng_formal_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectCommit()
			got, err := s.change(context.Background(), v.ID, v.IdentitySHA, 1, "approve", kind, "", 7, func(model.ManagementTraitsFormalVersion) bool { return true })
			if err != nil || len(got.Approvals) != 1 || got.Approvals[0].ActorID != a[0].ActorID || got.CanGenerate {
				t.Fatalf("approval: %+v %v", got, err)
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestManagementTraitsFormalActivationRevokeAndRollback(t *testing.T) {
	for _, action := range []string{"activate", "revoke"} {
		for _, auditFail := range []bool{false, true} {
			t.Run(action+map[bool]string{false: "_commit", true: "_rollback"}[auditFail], func(t *testing.T) {
				v, a := formalRegistryFixture(t)
				if action == "revoke" {
					v.State = "active"
				}
				db, m := managementRuntimeDB(t)
				s := &ManagementTraitsFormalRegistry{db: db, environment: "local"}
				m.ExpectBegin()
				m.ExpectQuery("SELECT .*el_mng_formal_version.*FOR UPDATE").WithArgs(v.ID, 1).WillReturnRows(managementRuntimeModelRows(t, v))
				m.ExpectQuery("SELECT .*el_mng_formal_approval").WithArgs(v.ID, 3).WillReturnRows(managementRuntimeModelRows(t, a[0], a[1]))
				m.ExpectExec("UPDATE `el_mng_formal_version`.*WHERE id = .*epoch =").WillReturnResult(sqlmock.NewResult(0, 1))
				if auditFail {
					m.ExpectExec("INSERT INTO `el_mng_formal_audit`").WillReturnError(errors.New("private DB detail"))
					m.ExpectRollback()
				} else {
					m.ExpectExec("INSERT INTO `el_mng_formal_audit`").WillReturnResult(sqlmock.NewResult(1, 1))
					m.ExpectCommit()
				}
				got, err := s.change(context.Background(), v.ID, v.IdentitySHA, 1, action, "", "reason", 7, func(model.ManagementTraitsFormalVersion) bool { return true })
				if auditFail {
					if !errors.Is(err, ErrManagementTraitsFormalInvalid) {
						t.Fatal(err)
					}
				} else if err != nil || got.Version.Epoch != 2 || got.Version.State != map[string]string{"activate": "active", "revoke": "revoked"}[action] {
					t.Fatalf("transition: %+v %v", got, err)
				}
				if err = m.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestManagementTraitsFormalRejectsStaleAndRevoked(t *testing.T) {
	for _, failure := range []string{"epoch", "identity", "revoked", "assets"} {
		t.Run(failure, func(t *testing.T) {
			v, _ := formalRegistryFixture(t)
			if failure == "revoked" {
				v.State = "revoked"
			}
			db, m := managementRuntimeDB(t)
			s := &ManagementTraitsFormalRegistry{db: db, environment: "local"}
			m.ExpectBegin()
			m.ExpectQuery("SELECT .*el_mng_formal_version.*FOR UPDATE").WillReturnRows(managementRuntimeModelRows(t, v))
			epoch := int64(1)
			sha := v.IdentitySHA
			if failure == "epoch" {
				epoch = 2
			}
			if failure == "identity" {
				sha = strings.Repeat("f", 64)
			}
			if failure == "assets" {
				m.ExpectQuery("SELECT .*el_mng_formal_approval").WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}
			m.ExpectRollback()
			_, err := s.change(context.Background(), v.ID, sha, epoch, "approve", "content", "", 7, func(model.ManagementTraitsFormalVersion) bool { return failure != "assets" })
			if err == nil {
				t.Fatal("invalid transition accepted")
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

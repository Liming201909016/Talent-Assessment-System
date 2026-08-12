package service

import (
	"os"
	"strings"
	"testing"

	"github.com/talent-assessment/refactored/internal/model"
)

func TestNormalizeCompetencyVersionSet(t *testing.T) {
	tests := []struct {
		name    string
		input   CompetencyVersionSet
		want    CompetencyVersionSet
		wantErr bool
	}{
		{
			name:  "current defaults",
			input: CompetencyVersionSet{},
			want: CompetencyVersionSet{
				ProductVersion:        CompetencyProductV1,
				ScoringVersion:        CompetencyScoringV1,
				ContentVersion:        CompetencyTemporaryContentVersion,
				ReportTemplateVersion: CompetencyReportTemplateV1,
			},
		},
		{
			name: "future content version remains explicit",
			input: CompetencyVersionSet{
				ProductVersion:        CompetencyProductV1,
				ScoringVersion:        CompetencyScoringV1,
				ContentVersion:        "phase1-content-20260807",
				ReportTemplateVersion: CompetencyReportTemplateV1,
			},
			want: CompetencyVersionSet{
				ProductVersion:        CompetencyProductV1,
				ScoringVersion:        CompetencyScoringV1,
				ContentVersion:        "phase1-content-20260807",
				ReportTemplateVersion: CompetencyReportTemplateV1,
			},
		},
		{name: "unsupported product", input: CompetencyVersionSet{ProductVersion: "phase1-unknown"}, wantErr: true},
		{name: "unsupported scoring", input: CompetencyVersionSet{ScoringVersion: "competency-v2"}, wantErr: true},
		{name: "unsupported template", input: CompetencyVersionSet{ReportTemplateVersion: "competency-report-v2"}, wantErr: true},
		{name: "invalid content identifier", input: CompetencyVersionSet{ContentVersion: "阶段一 正式版"}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeCompetencyVersionSet(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Fatalf("versions=%+v want=%+v", got, test.want)
			}
		})
	}
}

func TestCompetencyRuntime_UsesFrozenVersionSet(t *testing.T) {
	data, err := os.ReadFile("competency_runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, required := range []string{
		"IsPhase1CompetencyVersionSet",
		"CompetencyVersionSetFromExam(exam)",
		"ProductVersion: versions.ProductVersion",
		"ScoringVersion: versions.ScoringVersion",
		"ContentVersion: versions.ContentVersion",
		"ReportTemplateVersion: versions.ReportTemplateVersion",
		`content_version = ? AND audience = ? AND status = 0", result.ContentVersion`,
		"BuildCompetencyReportTextSnapshot(result.ContentVersion",
		`data["reportTemplateVersion"] = result.ReportTemplateVersion`,
	} {
		if !strings.Contains(source, required) {
			t.Errorf("competency runtime version freeze missing %q", required)
		}
	}
}

func TestCompetencyVersionSetFromExam_UsesFrozenValues(t *testing.T) {
	exam := model.Exam{
		CompetencyProductVersion:        "product-frozen",
		CompetencyScoringVersion:        "score-frozen",
		CompetencyContentVersion:        "content-frozen",
		CompetencyReportTemplateVersion: "template-frozen",
	}
	got := CompetencyVersionSetFromExam(exam)
	want := CompetencyVersionSet{
		ProductVersion:        "product-frozen",
		ScoringVersion:        "score-frozen",
		ContentVersion:        "content-frozen",
		ReportTemplateVersion: "template-frozen",
	}
	if got != want {
		t.Fatalf("versions=%+v want=%+v", got, want)
	}
}

func TestValidateExecutableCompetencyVersions_RejectsPhase1UntilReportRendererExists(t *testing.T) {
	if err := ValidateExecutableCompetencyVersions(NormalizePhase1CompetencyConfiguration().Versions); err == nil {
		t.Fatal("phase-1 report versions accepted before the L1-L5/group/validity renderer exists")
	}
}

func TestNormalizePhase1CompetencyConfiguration(t *testing.T) {
	wantDimensions := []string{
		"competency-a1-01", "competency-a1-02", "competency-a1-03", "competency-a1-04", "competency-a1-05",
		"competency-b1-01", "competency-b1-02", "competency-b1-03", "competency-b1-04", "competency-b1-05",
	}
	wantVersions := CompetencyVersionSet{
		ProductVersion:        "competency-frontline-phase1-v1",
		ScoringVersion:        "competency-phase1-scoring-v1",
		ContentVersion:        "competency-phase1-content-v1",
		ReportTemplateVersion: "competency-phase1-report-v1",
	}

	profile := NormalizePhase1CompetencyConfiguration()
	if profile.ReportAudience != CompetencyReportAudienceFrontlineEmployee {
		t.Fatalf("audience=%q", profile.ReportAudience)
	}
	if strings.Join(profile.DimensionIDs, ",") != strings.Join(wantDimensions, ",") {
		t.Fatalf("dimensions=%v", profile.DimensionIDs)
	}
	if profile.Versions != wantVersions {
		t.Fatalf("versions=%+v", profile.Versions)
	}
	if profile.DefaultDurationMinutes != 20 || profile.DimensionQuestionCount != 80 || profile.ValidityQuestionCount != 10 {
		t.Fatalf("profile counts/duration=%+v", profile)
	}

	profile.DimensionIDs[0] = "mutated"
	if NormalizePhase1CompetencyConfiguration().DimensionIDs[0] != "competency-a1-01" {
		t.Fatal("phase-1 dimension IDs must be returned as an independent slice")
	}
}

func TestValidatePhase1CompetencyConfiguration(t *testing.T) {
	valid := NormalizePhase1CompetencyConfiguration()
	tests := []struct {
		name       string
		audience   string
		dimensions []string
		versions   CompetencyVersionSet
		wantErr    bool
	}{
		{"valid fixed profile", valid.ReportAudience, valid.DimensionIDs, valid.Versions, false},
		{"leader rejected", CompetencyReportAudienceLeader, valid.DimensionIDs, valid.Versions, true},
		{"missing dimension rejected", valid.ReportAudience, valid.DimensionIDs[:9], valid.Versions, true},
		{"reordered dimensions rejected", valid.ReportAudience, append([]string{valid.DimensionIDs[1]}, append([]string{valid.DimensionIDs[0]}, valid.DimensionIDs[2:]...)...), valid.Versions, true},
		{"other product rejected", valid.ReportAudience, valid.DimensionIDs, CompetencyVersionSet{ProductVersion: CompetencyProductV1, ScoringVersion: valid.Versions.ScoringVersion, ContentVersion: valid.Versions.ContentVersion, ReportTemplateVersion: valid.Versions.ReportTemplateVersion}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidatePhase1CompetencyConfiguration(test.audience, test.dimensions, test.versions)
			if (err != nil) != test.wantErr {
				t.Fatalf("error=%v wantErr=%v", err, test.wantErr)
			}
		})
	}
}

func TestValidatePhase1QuestionInventory(t *testing.T) {
	profile := NormalizePhase1CompetencyConfiguration()
	valid := make(map[string]Phase1QuestionTypeCounts, len(profile.DimensionIDs))
	for _, id := range profile.DimensionIDs {
		valid[id] = Phase1QuestionTypeCounts{Dimension: 8, Validity: 1}
	}
	if err := ValidatePhase1QuestionInventory(profile.DimensionIDs, valid); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(map[string]Phase1QuestionTypeCounts){
		"dimension count": func(counts map[string]Phase1QuestionTypeCounts) {
			counts[profile.DimensionIDs[0]] = Phase1QuestionTypeCounts{Dimension: 7, Validity: 1}
		},
		"validity count": func(counts map[string]Phase1QuestionTypeCounts) {
			counts[profile.DimensionIDs[0]] = Phase1QuestionTypeCounts{Dimension: 8, Validity: 0}
		},
		"unknown type": func(counts map[string]Phase1QuestionTypeCounts) {
			counts[profile.DimensionIDs[0]] = Phase1QuestionTypeCounts{Dimension: 8, Validity: 1, Other: 1}
		},
	} {
		t.Run(name, func(t *testing.T) {
			counts := make(map[string]Phase1QuestionTypeCounts, len(valid))
			for key, value := range valid {
				counts[key] = value
			}
			mutate(counts)
			if err := ValidatePhase1QuestionInventory(profile.DimensionIDs, counts); err == nil {
				t.Fatal("invalid phase-1 inventory must be rejected")
			}
		})
	}
}

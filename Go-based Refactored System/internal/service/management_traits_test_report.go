package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/talent-assessment/refactored/internal/model"
	"github.com/xuri/excelize/v2"
)

const managementTraitsTestWorkbookSHA = "b0498249ae057e3aa1b798922ed2d53a6ca304811943b3b41f4f64f9b024e84c"

type managementTraitsTestRule struct{ Summary, Diagnosis, Advice string }
type managementTraitsTestOverallRule struct {
	Level, Diagnosis string
	Advice           []string
}

// An opaque, SHA-locked candidate-only content set. It cannot be activated as
// formal content and offers no caller-supplied reviewed/approved flag.
type ManagementTraitsTestContent struct {
	dimensions map[string][5]managementTraitsTestRule
	overall    [5]managementTraitsTestOverallRule
	sourceSHA  string
}

func (c ManagementTraitsTestContent) SourceSHA() string { return c.sourceSHA }
func (c ManagementTraitsTestContent) RuleCount() int {
	if c.sourceSHA != managementTraitsTestWorkbookSHA || len(c.dimensions) != 13 {
		return 0
	}
	return 13*5*3 + 5*2
}

func LoadManagementTraitsTestContent(raw []byte) (ManagementTraitsTestContent, error) {
	fail := func() (ManagementTraitsTestContent, error) {
		return ManagementTraitsTestContent{}, errors.New("management traits test content rejected")
	}
	hash := sha256.Sum256(raw)
	if len(raw) > 20<<20 || hex.EncodeToString(hash[:]) != managementTraitsTestWorkbookSHA {
		return fail()
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return fail()
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) != 6 {
		return fail()
	}
	cell := func(sheet int, column string, row int) (string, error) {
		v, err := f.GetCellValue(sheets[sheet-1], fmt.Sprintf("%s%d", column, row))
		if err != nil || !utf8.ValidString(v) || strings.TrimSpace(v) == "" {
			return "", errors.New("missing exact content cell")
		}
		return v, nil
	}
	c := ManagementTraitsTestContent{dimensions: make(map[string][5]managementTraitsTestRule, 13), sourceSHA: managementTraitsTestWorkbookSHA}
	for i, d := range ManagementTraitsDimensions() {
		name, err := cell(2, "B", i+3)
		if err != nil || strings.TrimSpace(name) != d.Name {
			return fail()
		}
		var rules [5]managementTraitsTestRule
		for level := 0; level < 5; level++ {
			summary, e1 := cell(1, string("CDEFG"[level]), i+3)
			diagnosis, e2 := cell(2, string("DEFGH"[level]), i+3)
			advice, e3 := cell(3, string("CDEFG"[level]), i+3)
			if e1 != nil || e2 != nil || e3 != nil {
				return fail()
			}
			rules[level] = managementTraitsTestRule{summary, diagnosis, advice}
		}
		c.dimensions[d.Key] = rules
	}
	for level := 0; level < 5; level++ {
		label, e1 := cell(4, "B", level+2)
		diagnosis, e2 := cell(4, "C", level+2)
		advice, e3 := cell(4, "D", level+2)
		paragraphs := strings.Split(advice, "\n")
		if e1 != nil || e2 != nil || e3 != nil || len(paragraphs) != 3 {
			return fail()
		}
		for _, p := range paragraphs {
			if strings.TrimSpace(p) == "" {
				return fail()
			}
		}
		c.overall[level] = managementTraitsTestOverallRule{label, diagnosis, paragraphs}
	}
	return c, nil
}

type ManagementTraitsTestReportParticipant struct {
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Telephone   string  `json:"telephone"`
	Affiliation string  `json:"affiliation"`
	Post        string  `json:"post"`
	Age         *int    `json:"age,omitempty"`
	Degree      *string `json:"degree,omitempty"`
	Major       *string `json:"major,omitempty"`
	StuFlag     *int    `json:"stuFlag,omitempty"`
}
type ManagementTraitsTestReportDimension struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	Score      string `json:"score"`
	Norm       string `json:"norm"`
	Level      string `json:"level"`
	Diagnosis  string `json:"diagnosis"`
	Advice     string `json:"advice"`
	ChartScore string `json:"chartScore"`
	ChartNorm  string `json:"chartNorm"`
}
type ManagementTraitsTestReportModule struct{ Key, Score, ChartScore string }
type ManagementTraitsTestReportSelection struct{ Key, Name, Text string }
type ManagementTraitsTestReportOverall struct {
	Score      string   `json:"score"`
	Level      string   `json:"level"`
	Diagnosis  string   `json:"diagnosis"`
	Advice     []string `json:"advice"`
	ChartScore string   `json:"chartScore"`
}
type ManagementTraitsTestReportData struct {
	Schema           string                                `json:"schema"`
	TestOnly         bool                                  `json:"testOnly"`
	RunID            string                                `json:"runId"`
	PaperID          string                                `json:"paperId"`
	ExamID           string                                `json:"examId"`
	Questionnaire    string                                `json:"questionnaire"`
	Versions         ManagementTraitsScoringVersions       `json:"versions"`
	InputSHA         string                                `json:"inputSha"`
	ContentSourceSHA string                                `json:"contentSourceSha"`
	Participant      ManagementTraitsTestReportParticipant `json:"participant"`
	SubmittedAt      string                                `json:"submittedAt"`
	Overall          ManagementTraitsTestReportOverall     `json:"overall"`
	Dimensions       []ManagementTraitsTestReportDimension `json:"dimensions"`
	Modules          []ManagementTraitsTestReportModule    `json:"modules"`
	Highest          []ManagementTraitsTestReportSelection `json:"highest"`
	Lowest           []ManagementTraitsTestReportSelection `json:"lowest"`
}

// Pure adapter for trusted stored facts. Runtime callers still must load these
// models from a consistent DB transaction and validate the actual owner/profile.
// This is never a formal-report permission or a historical-recompute path.
func BuildManagementTraitsTestReport(bundle model.ManagementTraitsDefinitionBundle, snapshot model.ManagementTraitsPaperSnapshot, questions []model.ManagementTraitsPaperQuestionSnapshot, run model.ManagementTraitsResultRun, dimensions []model.ManagementTraitsResultDimension, modules []model.ManagementTraitsResultModule, receipt model.ManagementTraitsRuntimeReceipt, content ManagementTraitsTestContent, budget int) (ManagementTraitsTestReportData, error) {
	fail := func() (ManagementTraitsTestReportData, error) {
		return ManagementTraitsTestReportData{}, errors.New("management traits test report rejected")
	}
	if content.RuleCount() != 205 || snapshot.Source != "new_creation" || snapshot.IdentitySource != "submitted_snapshot" || run.Status != "completed" {
		return fail()
	}
	exact, err := validateManagementTraitsRuntimeResult(bundle, snapshot, questions, managementTraitsRuntimeRecords{run, dimensions, modules, receipt}, budget)
	if err != nil || !exact.IsComplete {
		return fail()
	}
	var identity managementTraitsRuntimeIdentity
	var fields managementTraitsRuntimeFieldWire
	if managementTraitsDecodeStrict([]byte(snapshot.ParticipantSnapshot), budget, &identity) != nil || managementTraitsDecodeStrict([]byte(snapshot.FieldContract), budget, &fields) != nil || !managementTraitsIdentityLegal(fields, identity) {
		return fail()
	}
	out := ManagementTraitsTestReportData{Schema: "mng-test-report-data-v1", TestOnly: true, RunID: run.ID, PaperID: run.PaperID, ExamID: run.ExamID, Questionnaire: run.Questionnaire, Versions: ManagementTraitsScoringVersions{bundle.ProductVersion, bundle.QuestionVersion, bundle.ScoringVersion, bundle.NormVersion}, InputSHA: run.InputSHA, ContentSourceSHA: content.sourceSHA, Participant: ManagementTraitsTestReportParticipant{Name: identity.Name, Gender: identity.Gender, Telephone: identity.Telephone, Affiliation: identity.Affiliation, Post: identity.Post, Age: identity.Age, Degree: identity.Degree, Major: identity.Major, StuFlag: identity.StuFlag}, SubmittedAt: run.SubmittedAt.In(snapshot.StartedAt.Location()).Format("2006-01-02 15:04")}
	return managementTraitsReportFromExact(exact, content, out)
}

// Both callers validate their own source policy before sharing content selection.
// This helper neither changes the new_creation guard nor authorizes reissuance.
func managementTraitsReportFromExact(exact ManagementTraitsResult, content ManagementTraitsTestContent, out ManagementTraitsTestReportData) (ManagementTraitsTestReportData, error) {
	fail := func() (ManagementTraitsTestReportData, error) {
		return ManagementTraitsTestReportData{}, errors.New("management traits test report rejected")
	}
	index := func(level string) int {
		for i, l := range []string{"excellent", "good", "qualified", "weak", "insufficient"} {
			if level == l {
				return i
			}
		}
		return -1
	}
	overallIndex := index(exact.OverallLevel)
	if overallIndex < 0 {
		return fail()
	}
	o := content.overall[overallIndex]
	out.Overall = ManagementTraitsTestReportOverall{exact.OverallScore.FloatString(2), o.Level, o.Diagnosis, append([]string{}, o.Advice...), exact.OverallScore.FloatString(12)}
	out.Dimensions = make([]ManagementTraitsTestReportDimension, 0, 13)
	out.Modules = make([]ManagementTraitsTestReportModule, 0, 4)
	out.Highest = make([]ManagementTraitsTestReportSelection, 0, 3)
	out.Lowest = make([]ManagementTraitsTestReportSelection, 0, 3)
	lookup := make(map[string]ManagementTraitsTestReportSelection, 13)
	labels := []string{"优秀", "良好", "合格", "欠佳", "不足"}
	for _, d := range exact.Dimensions {
		i := index(d.Level)
		rules, ok := content.dimensions[d.Key]
		if !ok || i < 0 {
			return fail()
		}
		rule := rules[i]
		out.Dimensions = append(out.Dimensions, ManagementTraitsTestReportDimension{d.Key, d.Name, d.Score.FloatString(2), d.Norm.FloatString(2), labels[i], rule.Diagnosis, rule.Advice, d.Score.FloatString(12), d.Norm.FloatString(12)})
		lookup[d.Key] = ManagementTraitsTestReportSelection{d.Key, d.Name, rule.Summary}
	}
	for _, m := range exact.Modules {
		out.Modules = append(out.Modules, ManagementTraitsTestReportModule{m.Key, m.Score.FloatString(2), m.Score.FloatString(12)})
	}
	for _, key := range exact.Highest {
		out.Highest = append(out.Highest, lookup[key])
	}
	for _, key := range exact.Lowest {
		out.Lowest = append(out.Lowest, lookup[key])
	}
	return out, nil
}

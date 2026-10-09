package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"unicode"
	"unicode/utf8"
)

type ManagementTraitsScoringVersions struct {
	Product  string `json:"product"`
	Question string `json:"question"`
	Scoring  string `json:"scoring"`
	Norm     string `json:"norm"`
}

type ManagementTraitsManifestOption struct {
	Raw          int    `json:"raw"`
	Content      string `json:"content"`
	DisplayOrder int    `json:"displayOrder"`
}

type ManagementTraitsManifestQuestion struct {
	Number       int                              `json:"number"`
	DimensionKey string                           `json:"dimensionKey"`
	Reverse      bool                             `json:"reverse"`
	Content      string                           `json:"content"`
	Options      []ManagementTraitsManifestOption `json:"options"`
}

type ManagementTraitsManifest struct {
	Questionnaire string                             `json:"questionnaire"`
	Versions      ManagementTraitsScoringVersions    `json:"versions"`
	Questions     []ManagementTraitsManifestQuestion `json:"questions"`
}

type ManagementTraitsCanonical struct {
	JSON   []byte
	SHA256 string
}

type ManagementTraitsMappedOption struct {
	SourceOptionID string `json:"sourceOptionId"`
	Raw            int    `json:"raw"`
	Content        string `json:"content"`
	DisplayOrder   int    `json:"displayOrder"`
}

type ManagementTraitsMappedQuestion struct {
	Number           int                            `json:"number"`
	SourceQuestionID string                         `json:"sourceQuestionId"`
	Content          string                         `json:"content"`
	Options          []ManagementTraitsMappedOption `json:"options"`
}

type ManagementTraitsMapping struct {
	Questionnaire string                           `json:"questionnaire"`
	ManifestSHA   string                           `json:"manifestSha"`
	Questions     []ManagementTraitsMappedQuestion `json:"questions"`
}

type ManagementTraitsSnapshotAnswer struct {
	Number           int    `json:"number"`
	PaperQuestionID  string `json:"paperQuestionId"`
	SourceQuestionID string `json:"sourceQuestionId"`
	DisplayOrder     int    `json:"displayOrder"`
	Answered         bool   `json:"answered"`
	SelectedOptionID string `json:"selectedOptionId"`
	Raw              int    `json:"raw"`
}

type ManagementTraitsSnapshotInput struct {
	PaperID         string                           `json:"paperId"`
	ExamID          string                           `json:"examId"`
	ParticipantType string                           `json:"participantType"`
	ParticipantID   string                           `json:"participantId"`
	Questionnaire   string                           `json:"questionnaire"`
	ManifestSHA     string                           `json:"manifestSha"`
	MappingSHA      string                           `json:"mappingSha"`
	Answers         []ManagementTraitsSnapshotAnswer `json:"answers"`
}

type ManagementTraitsValidatedInput struct {
	Canonical ManagementTraitsCanonical
	Answers   []ManagementTraitsAnswer
}

type managementTraitsCanonicalItem struct {
	Number  int  `json:"number"`
	Reverse bool `json:"reverse"`
}

type managementTraitsCanonicalDimension struct {
	Key    string                          `json:"key"`
	Module string                          `json:"module"`
	Order  int                             `json:"order"`
	Norm   string                          `json:"norm"`
	Items  []managementTraitsCanonicalItem `json:"items"`
}

type managementTraitsCanonicalPolicy struct {
	Dimension   string   `json:"dimension"`
	Modules     string   `json:"modules"`
	ModuleOrder []string `json:"moduleOrder"`
	Overall     string   `json:"overall"`
	NormSum     string   `json:"normSum"`
	OverallNorm string   `json:"overallNorm"`
	Reverse     string   `json:"reverse"`
	Incomplete  string   `json:"incomplete"`
	Precision   string   `json:"precision"`
	GradeBounds []int    `json:"gradeBounds"`
	Levels      []string `json:"levels"`
	Highest     int      `json:"highest"`
	Lowest      int      `json:"lowest"`
	Tie         string   `json:"tie"`
	Overlap     bool     `json:"overlap"`
}

type managementTraitsCanonicalManifest struct {
	Schema        string                               `json:"schema"`
	Questionnaire string                               `json:"questionnaire"`
	Versions      ManagementTraitsScoringVersions      `json:"versions"`
	Questions     []ManagementTraitsManifestQuestion   `json:"questions"`
	Dimensions    []managementTraitsCanonicalDimension `json:"dimensions"`
	Policy        managementTraitsCanonicalPolicy      `json:"policy"`
}

type managementTraitsCanonicalMapping struct {
	Schema        string                           `json:"schema"`
	Questionnaire string                           `json:"questionnaire"`
	ManifestSHA   string                           `json:"manifestSha"`
	Questions     []ManagementTraitsMappedQuestion `json:"questions"`
}

type managementTraitsFrozenQuestion struct {
	Schema          string                         `json:"schema"`
	ManifestSHA     string                         `json:"manifestSha"`
	MappingSHA      string                         `json:"mappingSha"`
	Question        ManagementTraitsMappedQuestion `json:"question"`
	DimensionKey    string                         `json:"dimensionKey"`
	Reverse         bool                           `json:"reverse"`
	PaperQuestionID string                         `json:"paperQuestionId"`
	DisplayOrder    int                            `json:"displayOrder"`
}

type managementTraitsCanonicalRow struct {
	ManagementTraitsSnapshotAnswer
	QuestionSHA string `json:"questionSha"`
}

type managementTraitsCanonicalInput struct {
	Schema          string                          `json:"schema"`
	PaperID         string                          `json:"paperId"`
	ExamID          string                          `json:"examId"`
	ParticipantType string                          `json:"participantType"`
	ParticipantID   string                          `json:"participantId"`
	Questionnaire   string                          `json:"questionnaire"`
	Versions        ManagementTraitsScoringVersions `json:"versions"`
	ManifestSHA     string                          `json:"manifestSha"`
	MappingSHA      string                          `json:"mappingSha"`
	Rows            []managementTraitsCanonicalRow  `json:"rows"`
}

func managementTraitsCanonicalBytes(payload interface{}) (ManagementTraitsCanonical, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return ManagementTraitsCanonical{}, errors.New("management traits canonical encoding failed")
	}
	h := sha256.Sum256(b)
	return ManagementTraitsCanonical{JSON: b, SHA256: hex.EncodeToString(h[:])}, nil
}

func managementTraitsVersionSyntax(v string) bool {
	if len(v) == 0 || len(v) > 64 {
		return false
	}
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if i > 0 && (c == '.' || c == '_' || c == '-') {
			continue
		}
		return false
	}
	return true
}

func managementTraitsOpaqueID(v string) bool {
	if len(v) == 0 || len(v) > 64 || !utf8.ValidString(v) || strings.TrimSpace(v) != v {
		return false
	}
	for _, r := range v {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// Syntax and claimed content only: this neither approves versions nor proves a source.
func CanonicalManagementTraitsManifest(manifest ManagementTraitsManifest) (ManagementTraitsCanonical, error) {
	if manifest.Questionnaire != ManagementTraitsQuestionnaireStaff && manifest.Questionnaire != ManagementTraitsQuestionnaireLeader {
		return ManagementTraitsCanonical{}, errors.New("invalid management traits questionnaire")
	}
	for _, v := range []string{manifest.Versions.Product, manifest.Versions.Question, manifest.Versions.Scoring, manifest.Versions.Norm} {
		if !managementTraitsVersionSyntax(v) {
			return ManagementTraitsCanonical{}, errors.New("invalid management traits version syntax")
		}
	}
	if len(manifest.Questions) != 140 {
		return ManagementTraitsCanonical{}, errors.New("management traits requires 140 manifest rows")
	}
	var keys [141]string
	var reverse [141]bool
	dimensions := make([]managementTraitsCanonicalDimension, 0, 13)
	normSum := new(big.Rat)
	for _, d := range ManagementTraitsDimensions() {
		items := make([]managementTraitsCanonicalItem, 0, len(d.Items))
		for _, item := range d.Items {
			keys[item.Number], reverse[item.Number] = d.Key, item.Reverse
			items = append(items, managementTraitsCanonicalItem{item.Number, item.Reverse})
		}
		dimensions = append(dimensions, managementTraitsCanonicalDimension{d.Key, d.Module, d.Order, d.Norm.RatString(), items})
		normSum.Add(normSum, d.Norm)
	}
	questions := make([]ManagementTraitsManifestQuestion, 140)
	var seen [141]bool
	texts := [...]string{"", "不符合", "不太符合", "一般", "比较符合", "很符合"}
	for _, q := range manifest.Questions {
		if q.Number < 1 || q.Number > 140 || seen[q.Number] || q.DimensionKey != keys[q.Number] || q.Reverse != reverse[q.Number] || !utf8.ValidString(q.Content) || strings.TrimSpace(q.Content) == "" || len(q.Options) != 5 {
			return ManagementTraitsCanonical{}, errors.New("invalid management traits manifest question")
		}
		seen[q.Number] = true
		options := make([]ManagementTraitsManifestOption, 5)
		var rawSeen, orderSeen [6]bool
		for _, o := range q.Options {
			if o.Raw < 1 || o.Raw > 5 || rawSeen[o.Raw] || o.DisplayOrder < 1 || o.DisplayOrder > 5 || orderSeen[o.DisplayOrder] || o.Content != texts[o.Raw] {
				return ManagementTraitsCanonical{}, errors.New("invalid management traits manifest option")
			}
			rawSeen[o.Raw], orderSeen[o.DisplayOrder] = true, true
			options[o.Raw-1] = o
		}
		q.Options = options
		questions[q.Number-1] = q
	}
	policy := managementTraitsCanonicalPolicy{
		Dimension: "25*(sum-count)/count", Modules: "equal-member-dimensions", ModuleOrder: []string{"self", "interpersonal", "task", "development"}, Overall: "equal-13-dimensions",
		NormSum: normSum.RatString(), OverallNorm: new(big.Rat).Quo(normSum, big.NewRat(13, 1)).RatString(), Reverse: "6-raw-once", Incomplete: "no-formal-scores", Precision: "exact-rational;display-half-up-2",
		GradeBounds: []int{90, 70, 30, 10}, Levels: []string{"excellent", "good", "qualified", "weak", "insufficient"}, Highest: 3, Lowest: 3, Tie: "dimension-order-ascending;unrounded;no-grade-filter", Overlap: true,
	}
	return managementTraitsCanonicalBytes(managementTraitsCanonicalManifest{"mng-scoring-manifest-v1", manifest.Questionnaire, manifest.Versions, questions, dimensions, policy})
}

func managementTraitsOrderedManifest(manifest ManagementTraitsManifest) [141]ManagementTraitsManifestQuestion {
	var rows [141]ManagementTraitsManifestQuestion
	for _, q := range manifest.Questions {
		rows[q.Number] = q
	}
	return rows
}

// Source IDs are opaque, globally unique within their entity kind, and unverified externally.
func CanonicalManagementTraitsMapping(manifest ManagementTraitsManifest, mapping ManagementTraitsMapping) (ManagementTraitsCanonical, error) {
	mc, err := CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		return ManagementTraitsCanonical{}, err
	}
	if mapping.Questionnaire != manifest.Questionnaire || mapping.ManifestSHA != mc.SHA256 || len(mapping.Questions) != 140 {
		return ManagementTraitsCanonical{}, errors.New("invalid management traits mapping binding")
	}
	manifestRows := managementTraitsOrderedManifest(manifest)
	rows := make([]ManagementTraitsMappedQuestion, 140)
	seenQuestions, seenOptions := make(map[string]bool, 140), make(map[string]bool, 700)
	var seen [141]bool
	for _, q := range mapping.Questions {
		if q.Number < 1 || q.Number > 140 || seen[q.Number] || !managementTraitsOpaqueID(q.SourceQuestionID) || seenQuestions[q.SourceQuestionID] || q.Content != manifestRows[q.Number].Content || len(q.Options) != 5 {
			return ManagementTraitsCanonical{}, errors.New("invalid management traits mapped question")
		}
		seen[q.Number], seenQuestions[q.SourceQuestionID] = true, true
		var expected [6]ManagementTraitsManifestOption
		for _, o := range manifestRows[q.Number].Options {
			expected[o.Raw] = o
		}
		options := make([]ManagementTraitsMappedOption, 5)
		var rawSeen [6]bool
		for _, o := range q.Options {
			if o.Raw < 1 || o.Raw > 5 || rawSeen[o.Raw] || !managementTraitsOpaqueID(o.SourceOptionID) || seenOptions[o.SourceOptionID] || o.Content != expected[o.Raw].Content || o.DisplayOrder != expected[o.Raw].DisplayOrder {
				return ManagementTraitsCanonical{}, errors.New("invalid management traits mapped option")
			}
			rawSeen[o.Raw], seenOptions[o.SourceOptionID] = true, true
			options[o.Raw-1] = o
		}
		q.Options = options
		rows[q.Number-1] = q
	}
	return managementTraitsCanonicalBytes(managementTraitsCanonicalMapping{"mng-source-mapping-v1", mapping.Questionnaire, mapping.ManifestSHA, rows})
}

// Validates claimed links, not actual DB ownership, history, submission state or authorization.
// Incomplete inputs are valid; consumers may pass Answers to S1 without double reversal.
func CanonicalManagementTraitsInput(manifest ManagementTraitsManifest, mapping ManagementTraitsMapping, input ManagementTraitsSnapshotInput) (ManagementTraitsValidatedInput, error) {
	pc, err := CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil {
		return ManagementTraitsValidatedInput{}, err
	}
	if input.Questionnaire != manifest.Questionnaire || input.ManifestSHA != mapping.ManifestSHA || input.MappingSHA != pc.SHA256 || !managementTraitsOpaqueID(input.PaperID) || !managementTraitsOpaqueID(input.ExamID) || !managementTraitsOpaqueID(input.ParticipantID) || (input.ParticipantType != "candidate" && input.ParticipantType != "tester") || len(input.Answers) != 140 {
		return ManagementTraitsValidatedInput{}, errors.New("invalid management traits input binding")
	}
	manifestRows := managementTraitsOrderedManifest(manifest)
	var mappingRows [141]ManagementTraitsMappedQuestion
	for _, q := range mapping.Questions {
		options := make([]ManagementTraitsMappedOption, 5)
		for _, o := range q.Options {
			options[o.Raw-1] = o
		}
		q.Options = options
		mappingRows[q.Number] = q
	}
	rows := make([]managementTraitsCanonicalRow, 140)
	answers := make([]ManagementTraitsAnswer, 140)
	var seen, orders [141]bool
	paperQuestions := make(map[string]bool, 140)
	for _, a := range input.Answers {
		if a.Number < 1 || a.Number > 140 || seen[a.Number] || a.DisplayOrder < 1 || a.DisplayOrder > 140 || orders[a.DisplayOrder] || !managementTraitsOpaqueID(a.PaperQuestionID) || paperQuestions[a.PaperQuestionID] || a.SourceQuestionID != mappingRows[a.Number].SourceQuestionID {
			return ManagementTraitsValidatedInput{}, errors.New("invalid management traits snapshot question")
		}
		q := mappingRows[a.Number]
		if a.Answered {
			if a.Raw < 1 || a.Raw > 5 || a.SelectedOptionID != q.Options[a.Raw-1].SourceOptionID {
				return ManagementTraitsValidatedInput{}, errors.New("invalid management traits selected option")
			}
		} else if a.Raw != 0 || a.SelectedOptionID != "" {
			return ManagementTraitsValidatedInput{}, errors.New("invalid management traits unanswered row")
		}
		seen[a.Number], orders[a.DisplayOrder], paperQuestions[a.PaperQuestionID] = true, true, true
		d := manifestRows[a.Number]
		frozen, err := managementTraitsCanonicalBytes(managementTraitsFrozenQuestion{"mng-frozen-question-v1", input.ManifestSHA, input.MappingSHA, q, d.DimensionKey, d.Reverse, a.PaperQuestionID, a.DisplayOrder})
		if err != nil {
			return ManagementTraitsValidatedInput{}, err
		}
		rows[a.Number-1] = managementTraitsCanonicalRow{a, frozen.SHA256}
		answers[a.Number-1] = ManagementTraitsAnswer{Number: a.Number, Answered: a.Answered, Raw: a.Raw}
	}
	c, err := managementTraitsCanonicalBytes(managementTraitsCanonicalInput{"mng-score-input-v1", input.PaperID, input.ExamID, input.ParticipantType, input.ParticipantID, input.Questionnaire, manifest.Versions, input.ManifestSHA, input.MappingSHA, rows})
	if err != nil {
		return ManagementTraitsValidatedInput{}, err
	}
	return ManagementTraitsValidatedInput{Canonical: c, Answers: answers}, nil
}

package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/talent-assessment/refactored/internal/model"
)

var errManagementTraitsStoredInput = errors.New("invalid management traits stored input")

// Stored canonical manifests only, not public manifest-input JSON. Reordering
// questions/options is permitted; normative dimension/item/policy order is fixed.
func DecodeManagementTraitsManifest(data []byte, maxBytes int) (ManagementTraitsManifest, error) {
	var stored managementTraitsCanonicalManifest
	if managementTraitsDecodeStrict(data, maxBytes, &stored) != nil {
		return ManagementTraitsManifest{}, errManagementTraitsStoredInput
	}
	manifest := ManagementTraitsManifest{Questionnaire: stored.Questionnaire, Versions: stored.Versions, Questions: stored.Questions}
	c, err := CanonicalManagementTraitsManifest(manifest)
	if err != nil {
		return ManagementTraitsManifest{}, errManagementTraitsStoredInput
	}
	var expected managementTraitsCanonicalManifest
	if json.Unmarshal(c.JSON, &expected) != nil {
		return ManagementTraitsManifest{}, errManagementTraitsStoredInput
	}
	// Normalize only S2B's unordered question/option collections, never metadata.
	sort.Slice(stored.Questions, func(i, j int) bool { return stored.Questions[i].Number < stored.Questions[j].Number })
	for i := range stored.Questions {
		opts := stored.Questions[i].Options
		sort.Slice(opts, func(i, j int) bool { return opts[i].Raw < opts[j].Raw })
	}
	if !reflect.DeepEqual(stored, expected) {
		return ManagementTraitsManifest{}, errManagementTraitsStoredInput
	}
	return ManagementTraitsManifest{Questionnaire: expected.Questionnaire, Versions: expected.Versions, Questions: expected.Questions}, nil
}

// Stored canonical mappings only. SHA relationships are revalidated by S2B;
// neither syntactic versions nor claimed source identities are approved here.
func DecodeManagementTraitsMapping(data []byte, manifest ManagementTraitsManifest, maxBytes int) (ManagementTraitsMapping, error) {
	var stored managementTraitsCanonicalMapping
	if managementTraitsDecodeStrict(data, maxBytes, &stored) != nil || stored.Schema != "mng-source-mapping-v1" {
		return ManagementTraitsMapping{}, errManagementTraitsStoredInput
	}
	mapping := ManagementTraitsMapping{Questionnaire: stored.Questionnaire, ManifestSHA: stored.ManifestSHA, Questions: stored.Questions}
	c, err := CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil {
		return ManagementTraitsMapping{}, errManagementTraitsStoredInput
	}
	var expected managementTraitsCanonicalMapping
	if json.Unmarshal(c.JSON, &expected) != nil {
		return ManagementTraitsMapping{}, errManagementTraitsStoredInput
	}
	return ManagementTraitsMapping{Questionnaire: expected.Questionnaire, ManifestSHA: expected.ManifestSHA, Questions: expected.Questions}, nil
}

// Private schema reader is restricted to the stored structs/option arrays above.
// Every exact JSON tag is required, including false values; maps, embedded fields,
// pointers, nulls, unknown fields and encoding/json case-fold aliases are rejected.
type managementTraitsJSONReader struct {
	decoder *json.Decoder
	fields  map[reflect.Type]map[string]reflect.Type
}

func (r *managementTraitsJSONReader) value(typ reflect.Type, depth int) error {
	if depth > 32 {
		return errManagementTraitsStoredInput
	}
	if typ.Kind() == reflect.Ptr {
		return r.value(typ.Elem(), depth+1)
	}
	token, err := r.decoder.Token()
	if err != nil || token == nil {
		return errManagementTraitsStoredInput
	}
	switch typ.Kind() {
	case reflect.Struct:
		if token != json.Delim('{') {
			return errManagementTraitsStoredInput
		}
		fields, ok := r.fields[typ]
		if !ok {
			fields = make(map[string]reflect.Type, typ.NumField())
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				key := strings.Split(f.Tag.Get("json"), ",")[0]
				if f.Anonymous || f.PkgPath != "" || key == "" || key == "-" {
					return errManagementTraitsStoredInput
				}
				if f.Type.Kind() == reflect.Ptr && typ != reflect.TypeOf(managementTraitsRuntimeIdentity{}) {
					return errManagementTraitsStoredInput
				}
				fields[key] = f.Type
			}
			r.fields[typ] = fields
		}
		seen := make(map[string]bool, len(fields))
		for r.decoder.More() {
			keyToken, err := r.decoder.Token()
			key, isString := keyToken.(string)
			field, exists := fields[key]
			if err != nil || !isString || !exists || seen[key] {
				return errManagementTraitsStoredInput
			}
			seen[key] = true
			if r.value(field, depth+1) != nil {
				return errManagementTraitsStoredInput
			}
		}
		end, err := r.decoder.Token()
		if err != nil || end != json.Delim('}') {
			return errManagementTraitsStoredInput
		}
		for key := range fields {
			optional := typ == reflect.TypeOf(managementTraitsRuntimeIdentity{}) && (key == "age" || key == "degree" || key == "major" || key == "stuFlag")
			if !seen[key] && !optional {
				return errManagementTraitsStoredInput
			}
		}
	case reflect.Slice:
		if token != json.Delim('[') {
			return errManagementTraitsStoredInput
		}
		for r.decoder.More() {
			if r.value(typ.Elem(), depth+1) != nil {
				return errManagementTraitsStoredInput
			}
		}
		end, err := r.decoder.Token()
		if err != nil || end != json.Delim(']') {
			return errManagementTraitsStoredInput
		}
	case reflect.String:
		if _, ok := token.(string); !ok {
			return errManagementTraitsStoredInput
		}
	case reflect.Bool:
		if _, ok := token.(bool); !ok {
			return errManagementTraitsStoredInput
		}
	case reflect.Int:
		n, ok := token.(json.Number)
		if !ok {
			return errManagementTraitsStoredInput
		}
		// ParseInt rejects decimal/exponent spelling, even if mathematically integral.
		if _, err := strconv.ParseInt(string(n), 10, typ.Bits()); err != nil {
			return errManagementTraitsStoredInput
		}
	default:
		return errManagementTraitsStoredInput
	}
	return nil
}

func managementTraitsDecodeStrict(data []byte, maxBytes int, target interface{}) error {
	if maxBytes <= 0 || len(data) > maxBytes || !utf8.Valid(data) || bytes.HasPrefix(data, []byte{239, 187, 191}) || !managementTraitsJSONSurrogates(data) {
		return errManagementTraitsStoredInput
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	r := managementTraitsJSONReader{decoder: d, fields: make(map[reflect.Type]map[string]reflect.Type)}
	if r.value(reflect.TypeOf(target).Elem(), 1) != nil {
		return errManagementTraitsStoredInput
	}
	if _, err := d.Token(); err != io.EOF {
		return errManagementTraitsStoredInput
	}
	if json.Unmarshal(data, target) != nil {
		return errManagementTraitsStoredInput
	}
	return nil
}

// encoding/json replaces lone UTF-16 surrogates with U+FFFD. Inspect escapes
// before that lossy step, without rejecting a legitimate literal/escaped U+FFFD.
// Ordinary JSON syntax and escape validity are subsequently checked by Token.
func managementTraitsJSONSurrogates(data []byte) bool {
	inString := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' {
			inString = !inString
			continue
		}
		if !inString || data[i] != '\\' {
			continue
		}
		i++
		if i >= len(data) {
			return false
		}
		if data[i] != 'u' {
			continue
		}
		if i+4 >= len(data) {
			return false
		}
		code, err := strconv.ParseUint(string(data[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if code >= 0xDC00 && code <= 0xDFFF {
			return false
		}
		if code >= 0xD800 && code <= 0xDBFF {
			if i+6 >= len(data) || data[i+1] != '\\' || data[i+2] != 'u' {
				return false
			}
			low, err := strconv.ParseUint(string(data[i+3:i+7]), 16, 16)
			if err != nil || low < 0xDC00 || low > 0xDFFF {
				return false
			}
			i += 6
		}
	}
	return true
}

// Read-only S2A -> S2B adapter: maxJSONBytes is BOTH a per-field limit and a
// cumulative budget for manifest + mapping + all 140 option JSON fields. The
// subtraction preflight bounds decoding allocations without integer overflow.
// Source/Status/evidence/identity-source/participant-snapshot/field-contract and
// all time/submission metadata are deliberately NOT validated. ProfileExamID is
// only a claimed matching exam link: no profile, owner or approval is loaded.
// This returns raw answers and validated input, not S1 scores or authorization.
func ValidateManagementTraitsStoredInput(bundle model.ManagementTraitsDefinitionBundle, paper model.ManagementTraitsPaperSnapshot, questions []model.ManagementTraitsPaperQuestionSnapshot, maxJSONBytes int) (ManagementTraitsValidatedInput, error) {
	if maxJSONBytes <= 0 || len(questions) != 140 || !managementTraitsOpaqueID(bundle.ID) || paper.BundleID != bundle.ID {
		return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
	}
	remaining := maxJSONBytes
	for _, size := range []int{len(bundle.ScoringManifest), len(paper.MappingSnapshot)} {
		if size > remaining {
			return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
		}
		remaining -= size
	}
	for _, q := range questions {
		if len(q.OptionsSnapshot) > remaining {
			return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
		}
		remaining -= len(q.OptionsSnapshot)
	}
	manifest, err := DecodeManagementTraitsManifest([]byte(bundle.ScoringManifest), maxJSONBytes)
	if err != nil {
		return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
	}
	mc, err := CanonicalManagementTraitsManifest(manifest)
	versions := ManagementTraitsScoringVersions{Product: bundle.ProductVersion, Question: bundle.QuestionVersion, Scoring: bundle.ScoringVersion, Norm: bundle.NormVersion}
	if err != nil || bundle.ScoringManifestSHA != mc.SHA256 || paper.ScoringManifestSHA != mc.SHA256 || bundle.Questionnaire != manifest.Questionnaire || versions != manifest.Versions {
		return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
	}
	if paper.ProfileExamID != nil && (!managementTraitsOpaqueID(*paper.ProfileExamID) || *paper.ProfileExamID != paper.ExamID) {
		return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
	}
	mapping, err := DecodeManagementTraitsMapping([]byte(paper.MappingSnapshot), manifest, maxJSONBytes)
	if err != nil {
		return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
	}
	pc, err := CanonicalManagementTraitsMapping(manifest, mapping)
	if err != nil || paper.MappingSHA != pc.SHA256 {
		return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
	}
	input := ManagementTraitsSnapshotInput{PaperID: paper.PaperID, ExamID: paper.ExamID, ParticipantType: paper.ParticipantType, ParticipantID: paper.ParticipantID, Questionnaire: manifest.Questionnaire, ManifestSHA: mc.SHA256, MappingSHA: pc.SHA256, Answers: make([]ManagementTraitsSnapshotAnswer, 0, 140)}
	ids := make(map[string]bool, 140)
	var seen [141]bool
	for _, row := range questions {
		if !managementTraitsOpaqueID(row.ID) || ids[row.ID] || row.PaperID != paper.PaperID || row.Number < 1 || row.Number > 140 || seen[row.Number] {
			return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
		}
		ids[row.ID], seen[row.Number] = true, true
		q, mapped := manifest.Questions[row.Number-1], mapping.Questions[row.Number-1]
		if row.DimensionKey != q.DimensionKey || row.Reverse != q.Reverse || row.Content != q.Content {
			return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
		}
		var options []ManagementTraitsMappedOption
		if managementTraitsDecodeStrict([]byte(row.OptionsSnapshot), maxJSONBytes, &options) != nil || len(options) != 5 {
			return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
		}
		var rawSeen [6]bool
		ordered := make([]ManagementTraitsMappedOption, 5)
		for _, option := range options {
			if option.Raw < 1 || option.Raw > 5 || rawSeen[option.Raw] {
				return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
			}
			rawSeen[option.Raw] = true
			ordered[option.Raw-1] = option
		}
		if !reflect.DeepEqual(ordered, mapped.Options) {
			return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
		}
		a := ManagementTraitsSnapshotAnswer{Number: row.Number, PaperQuestionID: row.PaperQuestionID, SourceQuestionID: row.SourceQuestionID, DisplayOrder: row.DisplayOrder}
		if row.SelectedOptionID != nil || row.RawAnswer != nil || row.FinalScore != nil {
			if row.SelectedOptionID == nil || row.RawAnswer == nil || row.FinalScore == nil {
				return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
			}
			a.Answered, a.SelectedOptionID, a.Raw = true, *row.SelectedOptionID, *row.RawAnswer
			if a.Raw < 1 || a.Raw > 5 {
				return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
			}
			final := a.Raw
			if q.Reverse {
				final = 6 - a.Raw
			}
			if *row.FinalScore != final {
				return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
			}
		}
		frozen, err := managementTraitsCanonicalBytes(managementTraitsFrozenQuestion{"mng-frozen-question-v1", mc.SHA256, pc.SHA256, mapped, q.DimensionKey, q.Reverse, a.PaperQuestionID, a.DisplayOrder})
		if err != nil || row.ScoringSnapshotSHA != frozen.SHA256 {
			return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
		}
		input.Answers = append(input.Answers, a)
	}
	validated, err := CanonicalManagementTraitsInput(manifest, mapping, input)
	if err != nil {
		return ManagementTraitsValidatedInput{}, errManagementTraitsStoredInput
	}
	return validated, nil
}

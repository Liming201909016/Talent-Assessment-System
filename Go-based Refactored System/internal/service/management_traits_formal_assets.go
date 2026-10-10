package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type managementTraitsFormalAssets struct{ contentSHA, workbookSHA, templateSHA, bindingSHA string }

type managementTraitsFormalAssetError struct{ reason string }

func (e *managementTraitsFormalAssetError) Error() string { return e.reason }
func (e *managementTraitsFormalAssetError) Unwrap() error { return ErrManagementTraitsFormalInvalid }
func formalAssetError(reason string) error                { return &managementTraitsFormalAssetError{reason: reason} }
func formalAssetReason(err error, fallback string) string {
	var target *managementTraitsFormalAssetError
	if errors.As(err, &target) {
		return target.reason
	}
	return fallback
}

func formalReadAsset(root, key, name string) ([]byte, error) {
	return formalReadAssetChecked(root, key, name, nil)
}

func formalReadAssetChecked(root, key, name string, afterOpen func(string) error) ([]byte, error) {
	prefix := strings.TrimSuffix(name, filepath.Ext(name))
	if !filepath.IsAbs(root) || !formalKey.MatchString(key) || (name != "content.xlsx" && name != "template.docx") {
		return nil, formalAssetError(prefix + "_location_invalid")
	}
	root = filepath.Clean(root)
	for _, p := range []string{root, filepath.Join(root, key)} {
		st, err := os.Lstat(p)
		if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return nil, formalAssetError(prefix + "_location_invalid")
		}
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil || resolved != root {
		return nil, formalAssetError(prefix + "_location_invalid")
	}
	file := filepath.Join(root, key, name)
	st, err := os.Lstat(file)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, formalAssetError(prefix + "_file_invalid")
	}
	if st.Size() <= 0 {
		return nil, formalAssetError(prefix + "_file_empty")
	}
	if st.Size() > 20<<20 {
		return nil, formalAssetError(prefix + "_file_too_large")
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, formalAssetError(prefix + "_read_failed")
	}
	defer f.Close()
	if afterOpen != nil {
		if err = afterOpen(file); err != nil {
			return nil, formalAssetError(prefix + "_identity_changed")
		}
	}
	actual, err := f.Stat()
	if err != nil || !os.SameFile(st, actual) || actual.Size() != st.Size() {
		return nil, formalAssetError(prefix + "_identity_changed")
	}
	b, err := io.ReadAll(io.LimitReader(f, (20<<20)+1))
	if err != nil || int64(len(b)) != actual.Size() || len(b) > 20<<20 {
		return nil, formalAssetError(prefix + "_identity_changed")
	}
	current, err := os.Lstat(file)
	if err != nil || !os.SameFile(st, current) || current.Size() != st.Size() {
		return nil, formalAssetError(prefix + "_identity_changed")
	}
	return b, nil
}

func loadManagementTraitsFormalAssets(root, key string) (managementTraitsFormalAssets, error) {
	var a managementTraitsFormalAssets
	raw, err := formalReadAsset(root, key, "content.xlsx")
	if err != nil {
		return a, ErrManagementTraitsFormalInvalid
	}
	a.contentSHA, a.workbookSHA, _, err = managementTraitsFormalContentIdentity(raw)
	if err != nil {
		return a, err
	}
	template, err := formalReadAsset(root, key, "template.docx")
	if os.IsNotExist(err) {
		return a, nil
	} // Usable draft; no approval/activation.
	if err != nil {
		return a, ErrManagementTraitsFormalInvalid
	}
	binding, err := validateManagementTraitsFormalTemplate(template)
	if err != nil {
		return a, err
	}
	a.templateSHA = formalSHA(template)
	a.bindingSHA = binding
	return a, nil
}

func managementTraitsFormalContentIdentity(raw []byte) (contentSHA, workbookSHA string, ruleCount int, err error) {
	c, err := LoadManagementTraitsTestContent(raw)
	if err != nil || c.RuleCount() != 205 {
		return "", "", 0, formalAssetError("content_contract_invalid")
	}
	rules := make([]any, 0, 13)
	for _, d := range ManagementTraitsDimensions() {
		rules = append(rules, struct {
			Key   string
			Rules [5]managementTraitsTestRule
		}{d.Key, c.dimensions[d.Key]})
	}
	workbookSHA = formalSHA(raw)
	contentSHA = formalSHA([]byte(formalJSON(struct {
		Domain     string
		Dimensions []any
		Overall    [5]managementTraitsTestOverallRule
	}{"mng-formal-content-v1", rules, c.overall})))
	return contentSHA, workbookSHA, c.RuleCount(), nil
}

type managementTraitsFormalTemplateBinding struct {
	SHA           string
	Tags          []string
	Charts        []string
	NumericLabels []string
}

type formalXMLNode struct {
	name     xml.Name
	attrs    []xml.Attr
	children []*formalXMLNode
	text     string
}

func (n *formalXMLNode) attr(key string) string {
	for _, a := range n.attrs {
		if a.Name.Local == key {
			return a.Value
		}
	}
	return ""
}
func (n *formalXMLNode) all(space, key string) []*formalXMLNode {
	out := make([]*formalXMLNode, 0)
	var visit func(*formalXMLNode)
	visit = func(p *formalXMLNode) {
		if p.name.Space == space && p.name.Local == key {
			out = append(out, p)
		}
		for _, c := range p.children {
			visit(c)
		}
	}
	visit(n)
	return out
}
func formalXML(raw []byte) (*formalXMLNode, error) {
	d := xml.NewDecoder(bytes.NewReader(raw))
	root := &formalXMLNode{}
	stack := []*formalXMLNode{root}
	nodes := 0
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, ErrManagementTraitsFormalInvalid
		}
		switch x := token.(type) {
		case xml.StartElement:
			nodes++
			if len(stack) > 128 || nodes > 100000 {
				return nil, ErrManagementTraitsFormalInvalid
			}
			n := &formalXMLNode{name: x.Name, attrs: x.Attr}
			p := stack[len(stack)-1]
			p.children = append(p.children, n)
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) < 2 {
				return nil, ErrManagementTraitsFormalInvalid
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			stack[len(stack)-1].text += string(x)
		case xml.Directive:
			return nil, ErrManagementTraitsFormalInvalid
		}
	}
	if len(stack) != 1 || len(root.children) != 1 {
		return nil, ErrManagementTraitsFormalInvalid
	}
	return root, nil
}

func formalRequiredTags() map[string]bool {
	keys := map[string]bool{"participant.name": true, "participant.gender": true, "participant.telephone": true, "participant.affiliation": true, "participant.post": true, "result.submittedAt": true, "overall.level": true, "overall.diagnosis": true}
	for i := 1; i <= 3; i++ {
		keys["overall.advice."+strconv.Itoa(i)] = true
		for _, kind := range []string{"high", "low"} {
			for _, field := range []string{"name", "text"} {
				keys["overview."+kind+"."+strconv.Itoa(i)+"."+field] = true
			}
		}
	}
	for _, d := range ManagementTraitsDimensions() {
		for _, f := range []string{"score", "norm", "level", "diagnosis", "advice"} {
			keys["dimension."+d.Key+"."+f] = true
		}
	}
	return keys
}

func validateManagementTraitsFormalTemplate(raw []byte) (string, error) {
	binding, err := validateManagementTraitsFormalTemplateBinding(raw)
	return binding.SHA, err
}

func validateManagementTraitsFormalTemplateBinding(raw []byte) (managementTraitsFormalTemplateBinding, error) {
	const w = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	const chartNS = "http://schemas.openxmlformats.org/drawingml/2006/chart"
	const drawing = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	const relNS = "http://schemas.openxmlformats.org/package/2006/relationships"
	var result managementTraitsFormalTemplateBinding
	result.Tags = make([]string, 0)
	result.Charts = make([]string, 0)
	result.NumericLabels = make([]string, 0)
	if len(raw) == 0 || len(raw) > 20<<20 {
		return result, formalAssetError("template_file_invalid")
	}
	z, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return result, formalAssetError("template_zip_invalid")
	}
	if len(z.File) > 256 {
		return result, formalAssetError("template_parts_limit")
	}
	parts := make(map[string][]byte)
	trees := make(map[string]*formalXMLNode)
	total := int64(0)
	for _, f := range z.File {
		name := f.Name
		lower := strings.ToLower(name)
		if _, exists := parts[name]; exists || path.Clean(name) != strings.TrimSuffix(name, "/") || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") || strings.Contains(lower, "vba") || strings.Contains(lower, "embeddings/") || strings.Contains(lower, "activex/") {
			return result, formalAssetError("template_security_invalid")
		}
		r, e := f.Open()
		if e != nil {
			return result, formalAssetError("template_zip_invalid")
		}
		b, e := io.ReadAll(io.LimitReader(r, (64<<20)-total+1))
		ce := r.Close()
		total += int64(len(b))
		if e != nil || ce != nil || total > 64<<20 {
			return result, formalAssetError("template_uncompressed_limit")
		}
		parts[name] = b
		if strings.HasSuffix(name, ".xml") || strings.HasSuffix(name, ".rels") {
			tree, e := formalXML(b)
			if e != nil {
				return result, formalAssetError("template_xml_invalid")
			}
			trees[name] = tree
		}
	}
	for _, required := range []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml", "word/_rels/document.xml.rels"} {
		if trees[required] == nil {
			return result, formalAssetError("template_contract_invalid")
		}
	}
	targets := make(map[string]string)
	for name, tree := range trees {
		if strings.HasSuffix(name, ".rels") {
			seen := make(map[string]bool)
			for _, r := range tree.all(relNS, "Relationship") {
				id, target := r.attr("Id"), r.attr("Target")
				if id == "" || seen[id] || target == "" || r.attr("TargetMode") == "External" || strings.ContainsAny(target, "\\:%?#") {
					return result, formalAssetError("template_relationship_invalid")
				}
				seen[id] = true
				resolved := path.Clean(path.Join(path.Dir(path.Dir(name)), target))
				if strings.HasPrefix(target, "/") {
					resolved = path.Clean(strings.TrimPrefix(target, "/"))
				}
				if strings.HasPrefix(resolved, "../") || parts[resolved] == nil {
					return result, formalAssetError("template_relationship_invalid")
				}
				if name == "word/_rels/document.xml.rels" {
					targets[id] = resolved
				}
			}
		}
		if strings.HasPrefix(name, "word/") {
			for _, n := range tree.all(w, "t") {
				if strings.Contains(strings.ToUpper(n.text), "TEST") || strings.Contains(n.text, "仅供系统测试") || strings.Contains(n.text, "不可作为人才决策依据") {
					return result, formalAssetError("template_test_marking_present")
				}
			}
			if len(tree.all(w, "instrText")) > 0 || len(tree.all(w, "altChunk")) > 0 {
				return result, formalAssetError("template_field_code_invalid")
			}
		}
		if len(tree.all(chartNS, "f")) > 0 || len(tree.all(chartNS, "externalData")) > 0 {
			return result, formalAssetError("template_chart_external_data")
		}
	}
	doc := trees["word/document.xml"]
	expected := formalRequiredTags()
	seen := make(map[string]bool)
	for name, tree := range trees {
		if !strings.HasPrefix(name, "word/") {
			continue
		}
		for _, sdt := range tree.all(w, "sdt") {
			tags := sdt.all(w, "tag")
			if len(tags) != 1 || len(sdt.all(w, "t")) == 0 || !expected[tags[0].attr("val")] || seen[tags[0].attr("val")] {
				return result, formalAssetError("template_contract_invalid")
			}
			seen[tags[0].attr("val")] = true
		}
	}
	if len(seen) != len(expected) {
		return result, formalAssetError("template_contract_invalid")
	}
	chartCounts := map[string][]int{"chart.overall": {2}, "chart.module.self": {2}, "chart.module.interpersonal": {2}, "chart.module.task": {2}, "chart.module.development": {2}, "chart.dimension.comparison": {13, 13}}
	resolved, numeric := make(map[string]bool), make(map[string]bool)
	for _, anchor := range doc.all(drawing, "anchor") {
		props := anchor.all(drawing, "docPr")
		if len(props) != 1 {
			continue
		}
		key := props[0].attr("title")
		if strings.HasSuffix(key, ".numeric-label") {
			base := strings.TrimSuffix(key, ".numeric-label")
			text := anchor.all(w, "t")
			if numeric[base] || base == "chart.dimension.comparison" || chartCounts[base] == nil || (len(text) != 3 && len(text) != 4) {
				return result, formalAssetError("template_contract_invalid")
			}
			numeric[base] = true
			continue
		}
		refs := anchor.all(chartNS, "chart")
		if len(refs) == 0 {
			continue
		}
		if len(refs) != 1 || resolved[key] || chartCounts[key] == nil {
			return result, formalAssetError("template_contract_invalid")
		}
		part := targets[refs[0].attr("id")]
		tree := trees[part]
		if tree == nil || !strings.HasPrefix(part, "word/charts/") {
			return result, formalAssetError("template_contract_invalid")
		}
		series := tree.all(chartNS, "ser")
		if len(series) != len(chartCounts[key]) {
			return result, formalAssetError("template_contract_invalid")
		}
		for i, s := range series {
			vals := s.all(chartNS, "val")
			if len(vals) != 1 {
				return result, formalAssetError("template_contract_invalid")
			}
			lit := vals[0].all(chartNS, "numLit")
			if len(lit) != 1 {
				return result, formalAssetError("template_contract_invalid")
			}
			points := lit[0].all(chartNS, "pt")
			if len(points) != chartCounts[key][i] {
				return result, formalAssetError("template_contract_invalid")
			}
			for j, p := range points {
				if p.attr("idx") != strconv.Itoa(j) || len(p.all(chartNS, "v")) != 1 {
					return result, formalAssetError("template_contract_invalid")
				}
			}
		}
		resolved[key] = true
	}
	if len(resolved) != 6 || len(numeric) != 5 {
		return result, formalAssetError("template_contract_invalid")
	}
	keys := make([]string, 0, len(expected))
	for k := range expected {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	chartKeys := make([]string, 0, len(resolved))
	for key := range resolved {
		chartKeys = append(chartKeys, key)
	}
	sort.Strings(chartKeys)
	numericKeys := make([]string, 0, len(numeric))
	for key := range numeric {
		numericKeys = append(numericKeys, key)
	}
	sort.Strings(numericKeys)
	result.Tags, result.Charts, result.NumericLabels = keys, chartKeys, numericKeys
	result.SHA = formalSHA([]byte(formalJSON(struct {
		Schema       string
		Tags         []string
		Charts       map[string][]int
		NumericSlots int
	}{"mng-formal-word-binding-v1", keys, chartCounts, 5})))
	return result, nil
}

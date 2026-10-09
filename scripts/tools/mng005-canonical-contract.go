package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

const (
	catalogName  = "managementTraitsDimensionSpecs"
	functionName = "ManagementTraitsDimensions"
	scoringName  = "CalculateManagementTraits"
)

type sourceContract struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type nodeContract struct {
	Source      string `json:"source"`
	StartLine   int    `json:"startLine"`
	StartColumn int    `json:"startColumn"`
	EndLine     int    `json:"endLine"`
	EndColumn   int    `json:"endColumn"`
	SHA256      string `json:"sha256"`
}

type itemContract struct {
	Number  int  `json:"number"`
	Reverse bool `json:"reverse"`
}

type dimensionContract struct {
	Key    string         `json:"key"`
	Name   string         `json:"name"`
	Module string         `json:"module"`
	Order  int            `json:"order"`
	Norm   string         `json:"norm"`
	Items  []itemContract `json:"items"`
}

type moduleContract struct {
	Key        string   `json:"key"`
	Dimensions []string `json:"dimensions"`
}

type canonicalContract struct {
	Schema            string                  `json:"schema"`
	Source            string                  `json:"source"`
	SourceSHA256      string                  `json:"sourceSHA256"`
	FunctionSHA256    string                  `json:"functionSHA256"`
	SpecLiteralSHA256 string                  `json:"specLiteralSHA256"`
	Sources           []sourceContract        `json:"sources"`
	ASTNodes          map[string]nodeContract `json:"astNodes"`
	Dimensions        []dimensionContract     `json:"dimensions"`
	Modules           []moduleContract        `json:"modules"`
	QuestionCount     int                     `json:"questionCount"`
	ForwardCount      int                     `json:"forwardCount"`
	ReverseCount      int                     `json:"reverseCount"`
}

type parsedSource struct {
	path string
	raw  []byte
	fset *token.FileSet
	file *ast.File
}

func hash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func nodeBytes(raw []byte, fset *token.FileSet, node ast.Node) ([]byte, error) {
	file := fset.File(node.Pos())
	if file == nil {
		return nil, errors.New("node has no token file")
	}
	start := file.Offset(node.Pos())
	end := file.Offset(node.End())
	if start < 0 || end < start || end > len(raw) {
		return nil, errors.New("node offsets are outside source")
	}
	return raw[start:end], nil
}

func relativePath(value string) string {
	relative := filepath.ToSlash(value)
	if absolute, err := filepath.Abs(value); err == nil {
		if workingDirectory, workingErr := os.Getwd(); workingErr == nil {
			if result, relativeErr := filepath.Rel(workingDirectory, absolute); relativeErr == nil {
				relative = filepath.ToSlash(result)
			}
		}
	}
	return relative
}

func parseSource(path string) (parsedSource, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return parsedSource{}, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, raw, 0)
	if err != nil {
		return parsedSource{}, err
	}
	return parsedSource{path: relativePath(path), raw: raw, fset: fset, file: file}, nil
}

func nodeEvidence(source parsedSource, node ast.Node) (nodeContract, error) {
	raw, err := nodeBytes(source.raw, source.fset, node)
	if err != nil {
		return nodeContract{}, err
	}
	start, end := source.fset.Position(node.Pos()), source.fset.Position(node.End())
	return nodeContract{Source: source.path, StartLine: start.Line, StartColumn: start.Column, EndLine: end.Line, EndColumn: end.Column, SHA256: hash(raw)}, nil
}

func stringLiteral(expr ast.Expr) (string, error) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", errors.New("expected string literal")
	}
	return strconv.Unquote(literal.Value)
}

func integerLiteral(expr ast.Expr) (int64, error) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.INT {
		return 0, errors.New("expected integer literal")
	}
	return strconv.ParseInt(literal.Value, 10, 64)
}

func signedInteger(expr ast.Expr) (int64, error) {
	if unary, ok := expr.(*ast.UnaryExpr); ok {
		if unary.Op != token.SUB {
			return 0, errors.New("only unary minus is supported")
		}
		value, err := integerLiteral(unary.X)
		return -value, err
	}
	return integerLiteral(expr)
}

func gcd(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func parseDimension(element ast.Expr, order int) (dimensionContract, error) {
	literal, ok := element.(*ast.CompositeLit)
	if !ok || len(literal.Elts) != 6 {
		return dimensionContract{}, errors.New("dimension spec must contain six literal fields")
	}
	key, err := stringLiteral(literal.Elts[0])
	if err != nil {
		return dimensionContract{}, err
	}
	name, err := stringLiteral(literal.Elts[1])
	if err != nil {
		return dimensionContract{}, err
	}
	module, err := stringLiteral(literal.Elts[2])
	if err != nil {
		return dimensionContract{}, err
	}
	numerator, err := integerLiteral(literal.Elts[3])
	if err != nil {
		return dimensionContract{}, err
	}
	denominator, err := integerLiteral(literal.Elts[4])
	if err != nil || denominator <= 0 {
		return dimensionContract{}, errors.New("dimension norm denominator must be positive")
	}
	terms, ok := literal.Elts[5].(*ast.CompositeLit)
	if !ok || len(terms.Elts) == 0 {
		return dimensionContract{}, errors.New("dimension terms must be a non-empty literal")
	}
	items := make([]itemContract, 0, len(terms.Elts))
	for _, expression := range terms.Elts {
		term, termErr := signedInteger(expression)
		if termErr != nil || term == 0 {
			return dimensionContract{}, errors.New("dimension term must be a non-zero signed integer literal")
		}
		item := itemContract{Number: int(term), Reverse: term < 0}
		if item.Reverse {
			item.Number = -item.Number
		}
		items = append(items, item)
	}
	divisor := gcd(numerator, denominator)
	return dimensionContract{Key: key, Name: name, Module: module, Order: order, Norm: fmt.Sprintf("%d/%d", numerator/divisor, denominator/divisor), Items: items}, nil
}

func containsIdentifier(node ast.Node, name string) bool {
	found := false
	ast.Inspect(node, func(current ast.Node) bool {
		identifier, ok := current.(*ast.Ident)
		if ok && identifier.Name == name {
			found = true
		}
		return !found
	})
	return found
}

func findFunction(file *ast.File, name string) *ast.FuncDecl {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == name {
			return function
		}
	}
	return nil
}

func stringConstants(files ...*ast.File) (map[string]string, error) {
	values := make(map[string]string)
	for _, file := range files {
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.CONST {
				continue
			}
			for _, rawSpec := range general.Specs {
				spec := rawSpec.(*ast.ValueSpec)
				if len(spec.Names) != len(spec.Values) {
					continue
				}
				for index, name := range spec.Names {
					if value, err := stringLiteral(spec.Values[index]); err == nil {
						values[name.Name] = value
					}
				}
			}
		}
	}
	return values, nil
}

func evaluateString(expr ast.Expr, constants map[string]string) (string, error) {
	if value, err := stringLiteral(expr); err == nil {
		return value, nil
	}
	identifier, ok := expr.(*ast.Ident)
	if !ok {
		return "", errors.New("expected a string literal or string constant")
	}
	value, exists := constants[identifier.Name]
	if !exists {
		return "", fmt.Errorf("unsupported string identifier %s", identifier.Name)
	}
	return value, nil
}

func moduleAggregation(function *ast.FuncDecl, constants map[string]string) (*ast.RangeStmt, []string, error) {
	var matches []*ast.RangeStmt
	ast.Inspect(function.Body, func(node ast.Node) bool {
		rangeStatement, ok := node.(*ast.RangeStmt)
		if !ok {
			return true
		}
		literal, ok := rangeStatement.X.(*ast.CompositeLit)
		if !ok || len(literal.Elts) != 4 {
			return true
		}
		values := make([]string, 0, 4)
		for _, element := range literal.Elts {
			value, err := evaluateString(element, constants)
			if err != nil {
				return true
			}
			values = append(values, value)
		}
		if containsIdentifier(rangeStatement.Body, "DimensionCount") && containsIdentifier(rangeStatement.Body, "Modules") && containsIdentifier(rangeStatement.Body, "Dimensions") {
			matches = append(matches, rangeStatement)
		}
		return true
	})
	if len(matches) != 1 {
		return nil, nil, fmt.Errorf("production scorer must contain exactly one supported module aggregation loop, found %d", len(matches))
	}
	literal := matches[0].X.(*ast.CompositeLit)
	keys := make([]string, 0, 4)
	for _, element := range literal.Elts {
		value, err := evaluateString(element, constants)
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, value)
	}
	if !containsIdentifier(matches[0].Body, "Module") || !containsIdentifier(matches[0].Body, "Score") {
		return nil, nil, errors.New("module aggregation does not consume dimension module assignments and scores")
	}
	return matches[0], keys, nil
}

func verifyDimensionBuilder(function *ast.FuncDecl) (string, error) {
	if function.Body == nil || len(function.Body.List) != 3 {
		return "", errors.New("dimension builder must contain the exact three-statement build mechanism")
	}
	assignment, ok := function.Body.List[0].(*ast.AssignStmt)
	if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return "", errors.New("dimension builder must initialize one return variable")
	}
	returnName, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok || returnName.Name == "" {
		return "", errors.New("dimension builder return variable is invalid")
	}
	outer, ok := function.Body.List[1].(*ast.RangeStmt)
	if !ok || !containsIdentifier(outer.X, catalogName) || !containsIdentifier(outer.Body, "terms") || !containsIdentifier(outer.Body, "Items") || !containsIdentifier(outer.Body, "Norm") || !containsIdentifier(outer.Body, "Module") {
		return "", errors.New("dimension builder does not structurally derive definitions from the canonical catalog")
	}
	result, ok := function.Body.List[2].(*ast.ReturnStmt)
	if !ok || len(result.Results) != 1 {
		return "", errors.New("dimension builder must return exactly one value")
	}
	returned, ok := result.Results[0].(*ast.Ident)
	if !ok || returned.Name != returnName.Name {
		return "", errors.New("dimension builder must return the variable it constructs")
	}
	declarations := 0
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.DeclStmt); ok {
			declarations++
		}
		return true
	})
	if declarations != 0 {
		return "", errors.New("dimension builder contains unsupported AST declarations")
	}
	return returnName.Name, nil
}

func rejectAlternateMappings(source parsedSource, allowed []*ast.CompositeLit, dimensionKeys map[string]bool) error {
	allowedNodes := make(map[*ast.CompositeLit]bool, len(allowed))
	for _, node := range allowed {
		allowedNodes[node] = true
	}
	var found error
	ast.Inspect(source.file, func(node ast.Node) bool {
		if found != nil {
			return false
		}
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		if allowedNodes[literal] {
			return false
		}
		ast.Inspect(literal, func(child ast.Node) bool {
			if found != nil {
				return false
			}
			switch value := child.(type) {
			case *ast.BasicLit:
				if value.Kind == token.STRING {
					decoded, err := strconv.Unquote(value.Value)
					if err == nil && dimensionKeys[decoded] {
						found = fmt.Errorf("alternate dimension mapping-like composite at %s", source.fset.Position(literal.Pos()))
					}
				}
			case *ast.Ident:
				if dimensionKeys[value.Name] {
					found = fmt.Errorf("alternate symbolic dimension mapping-like composite at %s", source.fset.Position(literal.Pos()))
				}
			}
			return true
		})
		return found == nil
	})
	return found
}

func buildContract(sourcePath, scoringPath string) (canonicalContract, error) {
	identity, err := parseSource(sourcePath)
	if err != nil {
		return canonicalContract{}, err
	}
	scoring, err := parseSource(scoringPath)
	if err != nil {
		return canonicalContract{}, err
	}
	var catalog *ast.CompositeLit
	for _, declaration := range identity.file.Decls {
		switch current := declaration.(type) {
		case *ast.GenDecl:
			for _, spec := range current.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || len(value.Names) != 1 || value.Names[0].Name != catalogName || len(value.Values) != 1 {
					continue
				}
				catalog, ok = value.Values[0].(*ast.CompositeLit)
				if !ok {
					return canonicalContract{}, errors.New("canonical catalog is not a composite literal")
				}
			}
		}
	}
	function := findFunction(identity.file, functionName)
	scoringFunction := findFunction(scoring.file, scoringName)
	if catalog == nil || function == nil || function.Body == nil {
		return canonicalContract{}, errors.New("canonical catalog or production API was not found")
	}
	if scoringFunction == nil || scoringFunction.Body == nil {
		return canonicalContract{}, errors.New("production scoring function was not found")
	}
	if _, err = verifyDimensionBuilder(function); err != nil {
		return canonicalContract{}, err
	}
	dimensions := make([]dimensionContract, 0, len(catalog.Elts))
	modulesByKey := make(map[string][]string)
	moduleOrder := make([]string, 0, 4)
	seenKeys := make(map[string]bool)
	seenQuestions := make(map[int]bool)
	forward, reverse := 0, 0
	for index, expression := range catalog.Elts {
		dimension, parseErr := parseDimension(expression, index+1)
		if parseErr != nil {
			return canonicalContract{}, fmt.Errorf("dimension %d: %w", index+1, parseErr)
		}
		if dimension.Key == "" || dimension.Name == "" || dimension.Module == "" || seenKeys[dimension.Key] {
			return canonicalContract{}, errors.New("dimension identity is empty or duplicated")
		}
		seenKeys[dimension.Key] = true
		if _, exists := modulesByKey[dimension.Module]; !exists {
			moduleOrder = append(moduleOrder, dimension.Module)
		}
		modulesByKey[dimension.Module] = append(modulesByKey[dimension.Module], dimension.Key)
		for _, item := range dimension.Items {
			if item.Number < 1 || item.Number > 140 || seenQuestions[item.Number] {
				return canonicalContract{}, errors.New("question identity is duplicated or out of range")
			}
			seenQuestions[item.Number] = true
			if item.Reverse {
				reverse++
			} else {
				forward++
			}
		}
		dimensions = append(dimensions, dimension)
	}
	constants, err := stringConstants(identity.file, scoring.file)
	if err != nil {
		return canonicalContract{}, err
	}
	moduleLoop, scoredModuleOrder, err := moduleAggregation(scoringFunction, constants)
	if err != nil {
		return canonicalContract{}, err
	}
	if len(dimensions) != 13 || len(scoredModuleOrder) != 4 || len(seenQuestions) != 140 || forward != 100 || reverse != 40 {
		return canonicalContract{}, fmt.Errorf("unexpected canonical counts dimensions/modules/questions/forward/reverse=%d/%d/%d/%d/%d", len(dimensions), len(moduleOrder), len(seenQuestions), forward, reverse)
	}
	if fmt.Sprint(moduleOrder) != fmt.Sprint(scoredModuleOrder) {
		return canonicalContract{}, fmt.Errorf("dimension catalog module order %v differs from production aggregation %v", moduleOrder, scoredModuleOrder)
	}
	modules := make([]moduleContract, 0, len(scoredModuleOrder))
	for _, key := range scoredModuleOrder {
		modules = append(modules, moduleContract{Key: key, Dimensions: modulesByKey[key]})
	}
	dimensionKeys := make(map[string]bool, len(dimensions))
	for _, dimension := range dimensions {
		dimensionKeys[dimension.Key] = true
	}
	moduleLiteral := moduleLoop.X.(*ast.CompositeLit)
	if err = rejectAlternateMappings(identity, []*ast.CompositeLit{catalog}, dimensionKeys); err != nil {
		return canonicalContract{}, err
	}
	if err = rejectAlternateMappings(scoring, []*ast.CompositeLit{moduleLiteral}, dimensionKeys); err != nil {
		return canonicalContract{}, err
	}
	functionRaw, err := nodeBytes(identity.raw, identity.fset, function)
	if err != nil {
		return canonicalContract{}, err
	}
	catalogRaw, err := nodeBytes(identity.raw, identity.fset, catalog)
	if err != nil {
		return canonicalContract{}, err
	}
	nodes := make(map[string]nodeContract)
	for name, value := range map[string]struct {
		source parsedSource
		node   ast.Node
	}{
		"dimensionCatalog": {identity, catalog}, "dimensionBuilder": {identity, function},
		"scoringFunction": {scoring, scoringFunction}, "moduleAggregation": {scoring, moduleLoop},
	} {
		nodes[name], err = nodeEvidence(value.source, value.node)
		if err != nil {
			return canonicalContract{}, err
		}
	}
	sources := []sourceContract{{Path: identity.path, SHA256: hash(identity.raw)}, {Path: scoring.path, SHA256: hash(scoring.raw)}}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	return canonicalContract{Schema: "mng005-canonical-contract-v2", Source: identity.path, SourceSHA256: hash(identity.raw), FunctionSHA256: hash(functionRaw), SpecLiteralSHA256: hash(catalogRaw), Sources: sources, ASTNodes: nodes, Dimensions: dimensions, Modules: modules, QuestionCount: len(seenQuestions), ForwardCount: forward, ReverseCount: reverse}, nil
}

func main() {
	source := flag.String("source", "Go-based Refactored System/internal/service/management_traits_identity.go", "canonical Go source file")
	scoring := flag.String("scoring-source", "Go-based Refactored System/internal/service/management_traits_scoring.go", "production scoring source file")
	output := flag.String("output", "", "optional create-only JSON output path")
	flag.Parse()
	contract, err := buildContract(*source, *scoring)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	raw, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	raw = append(raw, '\n')
	if *output != "" {
		file, createErr := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if createErr != nil {
			fmt.Fprintln(os.Stderr, createErr)
			os.Exit(1)
		}
		if _, err = file.Write(raw); err == nil {
			err = file.Close()
		} else {
			_ = file.Close()
		}
		if err != nil {
			_ = os.Remove(*output)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	_, _ = os.Stdout.Write(raw)
}

package runner

import (
	"strconv"
	"strings"

	declparser "infomunge/internal/declarations"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/stringutils"
	"infomunge/pkg/values"
)

// DeclarationKind identifies a parsed script header or module declaration.
type DeclarationKind string

const (
	DeclarationVersion   DeclarationKind = "version"
	DeclarationOutput    DeclarationKind = "output"
	DeclarationInput     DeclarationKind = "input"
	DeclarationNamespace DeclarationKind = "namespace"
	DeclarationImport    DeclarationKind = "import"
	DeclarationVar       DeclarationKind = "var"
	DeclarationFun       DeclarationKind = "fun"
	DeclarationType      DeclarationKind = "type"
)

// SourceSpan records the byte range for a declaration in the raw source used by
// runner diagnostics.
type SourceSpan struct {
	Start int
	End   int
}

// DeclarationSource preserves the original declaration text alongside its
// source range.
type DeclarationSource struct {
	Line    string
	Lines   []string
	Trimmed string
	Offset  int
	Span    SourceSpan
}

// Declaration is the shared declaration IR for script headers and modules.
type Declaration struct {
	Kind      DeclarationKind
	Source    DeclarationSource
	Version   *VersionDeclaration
	Output    *OutputDeclaration
	Input     *InputDeclaration
	Namespace *NamespaceDeclaration
	Import    *ImportDeclaration
	Var       *VarDeclaration
	Function  *FunctionDeclaration
	Type      *TypeDeclaration
}

type VersionDeclaration struct {
	Text string
}

type OutputDeclaration struct {
	MimeType string
	Options  string
}

type InputDeclaration struct {
	Text string
}

type NamespaceDeclaration struct {
	Prefix string
	URI    string
}

type ImportDeclaration struct {
	ModuleSpec    string
	NamesPart     string
	Names         []string
	Star          bool
	NamespaceOnly bool
}

type VarDeclaration = declparser.VarDeclaration
type ParamDeclaration = declparser.ParamDeclaration
type FunctionDeclaration = declparser.FunctionDeclaration

type TypeDeclaration struct {
	Name       string
	BaseType   string
	Properties values.Object
}

func newDeclarationSource(lines []string, index, consumed, offset int) DeclarationSource {
	selected := append([]string(nil), lines[index:index+consumed]...)
	raw := strings.Join(selected, "\n")
	return DeclarationSource{
		Line:    lines[index],
		Lines:   selected,
		Trimmed: strings.TrimSpace(lines[index]),
		Offset:  offset,
		Span: SourceSpan{
			Start: offset,
			End:   offset + len(raw),
		},
	}
}

func parseOutputDeclaration(trimmedLine string) OutputDeclaration {
	rest := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "output "))
	mimeType, options := splitFirstToken(rest)
	return OutputDeclaration{MimeType: mimeType, Options: options}
}

func parseNamespaceDeclaration(trimmedLine string) (*NamespaceDeclaration, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "ns "))
	parts := strings.Fields(rest)
	switch len(parts) {
	case 1:
		return &NamespaceDeclaration{URI: parts[0]}, nil
	case 2:
		return &NamespaceDeclaration{Prefix: parts[0], URI: parts[1]}, nil
	default:
		return nil, unifiederrors.ParseError("invalid namespace declaration: expected 'ns [prefix] uri'")
	}
}

func parseImportDeclaration(trimmedLine string) ImportDeclaration {
	rest := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "import "))
	if !strings.Contains(rest, " from ") {
		return ImportDeclaration{ModuleSpec: rest, NamespaceOnly: true}
	}

	parts := strings.SplitN(rest, " from ", 2)
	namesPart := strings.TrimSpace(parts[0])
	moduleSpec := strings.TrimSpace(parts[1])
	decl := ImportDeclaration{
		ModuleSpec: moduleSpec,
		NamesPart:  namesPart,
		Star:       namesPart == "*",
	}
	if !decl.Star {
		for _, name := range strings.Split(namesPart, ",") {
			trimmed := strings.TrimSpace(name)
			if trimmed != "" {
				decl.Names = append(decl.Names, trimmed)
			}
		}
	}
	return decl
}

func parseVarDeclarationFromLines(lines []string, start int) (*VarDeclaration, int, error) {
	return declparser.ParseVarDeclarationFromLines(lines, start)
}

func parseFunctionDeclarationLine(trimmedLine string) (string, []ParamDeclaration, string, error) {
	return declparser.ParseFunctionDeclarationLine(trimmedLine)
}

func parseFunctionDeclarationFromLines(lines []string, start int) (*FunctionDeclaration, int, error) {
	return declparser.ParseFunctionDeclarationFromLines(lines, start)
}

func parseTypeDeclaration(trimmedLine string) (*TypeDeclaration, error) {
	rest := strings.TrimPrefix(trimmedLine, "type ")

	eqIdx := strings.Index(rest, "=")
	if eqIdx < 0 {
		return nil, unifiederrors.ParseError("invalid type declaration: missing '='")
	}

	typeName := strings.TrimSpace(rest[:eqIdx])
	if typeName == "" {
		return nil, unifiederrors.ParseError("invalid type declaration: missing type name")
	}

	rhs := strings.TrimSpace(rest[eqIdx+1:])
	if rhs == "" {
		return nil, unifiederrors.ParseError("invalid type declaration: missing base type")
	}

	var baseType string
	var properties values.Object

	braceIdx := strings.Index(rhs, "{")
	if braceIdx < 0 {
		baseType = rhs
	} else {
		baseType = strings.TrimSpace(rhs[:braceIdx])
		propsStr := rhs[braceIdx:]

		if !strings.HasSuffix(strings.TrimSpace(propsStr), "}") {
			return nil, unifiederrors.ParseError("invalid type declaration: unclosed properties block")
		}

		parsed, err := parseTypeDeclarationProperties(propsStr)
		if err != nil {
			return nil, unifiederrors.WrapParse(err, "invalid type declaration")
		}
		properties = parsed
	}

	if baseType == "" {
		return nil, unifiederrors.ParseError("invalid type declaration: missing base type")
	}

	return &TypeDeclaration{
		Name:       typeName,
		BaseType:   baseType,
		Properties: properties,
	}, nil
}

func parseTypeDeclarationProperties(propsStr string) (values.Object, error) {
	propsStr = strings.TrimSpace(propsStr)
	if !strings.HasPrefix(propsStr, "{") || !strings.HasSuffix(propsStr, "}") {
		return nil, unifiederrors.ParseError("properties must be enclosed in braces")
	}

	inner := strings.TrimSpace(propsStr[1 : len(propsStr)-1])
	if inner == "" {
		return make(values.Object), nil
	}

	properties := make(values.Object)
	for _, pair := range splitPropertyPairs(inner) {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		colonIdx := strings.Index(pair, ":")
		if colonIdx < 0 {
			return nil, unifiederrors.ParseError("invalid property: missing ':'")
		}

		key := strings.TrimSpace(pair[:colonIdx])
		valueStr := strings.TrimSpace(pair[colonIdx+1:])
		if key == "" {
			return nil, unifiederrors.ParseError("invalid property: empty key")
		}

		value, err := parseTypeDeclarationPropertyValue(valueStr)
		if err != nil {
			return nil, unifiederrors.WrapParsef(err, "invalid property value for %q", key)
		}
		properties[key] = value
	}

	return properties, nil
}

// splitPropertyPairs splits a comma-separated list of property pairs, respecting quoted strings.
func splitPropertyPairs(s string) []string {
	var pairs []string
	var current strings.Builder
	var sc stringutils.ScanState

	for _, ch := range s {
		sc.AdvanceRune(ch)
		if !sc.InString() && ch == ',' {
			pairs = append(pairs, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(ch)
	}

	if current.Len() > 0 {
		pairs = append(pairs, current.String())
	}

	return pairs
}

func parseTypeDeclarationPropertyValue(s string) (values.Value, error) {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) >= 2 {
		return s[1 : len(s)-1], nil
	}

	if b, ok := parseDeclarationBoolLiteral(s); ok {
		return b, nil
	}

	if s == "null" || s == "nil" {
		return nil, nil
	}

	if num, ok := parseDeclarationNumericLiteral(s); ok {
		return num, nil
	}

	return nil, unifiederrors.ParseErrorf("cannot parse value: %s", s)
}

func parseDeclarationBoolLiteral(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}

func parseDeclarationNumericLiteral(s string) (values.Value, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}

	if !strings.ContainsAny(s, ".eE") {
		if iv, err := strconv.Atoi(s); err == nil {
			return iv, true
		}
	}

	if fv, err := strconv.ParseFloat(s, 64); err == nil {
		return fv, true
	}

	return nil, false
}

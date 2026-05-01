package runner

import (
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/preprocessor"
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

// Keep existing internal test names readable while the implementation moves to
// the shared Declaration IR.
type headerDirectiveKind = DeclarationKind

const (
	headerDirectiveVersion   = DeclarationVersion
	headerDirectiveOutput    = DeclarationOutput
	headerDirectiveInput     = DeclarationInput
	headerDirectiveNamespace = DeclarationNamespace
	headerDirectiveImport    = DeclarationImport
	headerDirectiveVar       = DeclarationVar
	headerDirectiveFun       = DeclarationFun
	headerDirectiveType      = DeclarationType
)

type headerDirective = Declaration

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

type VarDeclaration struct {
	Name       string
	Expression string
}

type ParamDeclaration struct {
	Name string
	Type string
}

type FunctionDeclaration struct {
	Name   string
	Params []ParamDeclaration
	Body   string
}

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
	prefix, uri, err := parseNamespaceDecl(trimmedLine)
	if err != nil {
		return nil, err
	}
	return &NamespaceDeclaration{Prefix: prefix, URI: uri}, nil
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
	if start >= len(lines) {
		return nil, 0, nil
	}

	firstLine := lines[start]
	trimmedFirst := strings.TrimSpace(firstLine)
	if !strings.HasPrefix(trimmedFirst, "var ") {
		return nil, 0, nil
	}

	rest := strings.TrimSpace(strings.TrimPrefix(trimmedFirst, "var "))
	if rest == "" {
		return nil, 0, unifiederrors.ParseErrorf("invalid variable declaration: missing variable name in %q", trimmedFirst)
	}

	namePart := rest
	if eqIdx := strings.Index(rest, "="); eqIdx >= 0 {
		namePart = strings.TrimSpace(rest[:eqIdx])
	}
	fields := strings.Fields(namePart)
	if len(fields) != 1 {
		return nil, 0, unifiederrors.ParseErrorf("invalid variable declaration: expected a single variable name in %q", trimmedFirst)
	}
	varName := fields[0]

	declLines := []string{firstLine}
	declRaw := firstLine
	eqIdx := strings.Index(declRaw, "=")
	if eqIdx < 0 {
		return nil, 0, unifiederrors.ParseErrorf("invalid variable declaration: missing '=' in %q", trimmedFirst)
	}
	exprStr := strings.TrimSpace(declRaw[eqIdx+1:])
	if exprStr != "" {
		strippedExpr := preprocessor.StripSingleLineComment(exprStr)
		if strings.TrimSpace(strippedExpr) != "" && isDelimiterBalanced(strippedExpr) {
			return &VarDeclaration{
				Name:       varName,
				Expression: exprStr,
			}, 1, nil
		}
	}

	i := start + 1
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			declLines = append(declLines, line)
			i++
			continue
		}
		if isDirectiveLine(trimmed) || (trimmed == "---" && parsedVarLinesAreComplete(declLines)) {
			break
		}
		declLines = append(declLines, line)
		i++
	}

	declRaw = strings.Join(declLines, "\n")
	eqIdx = strings.Index(declRaw, "=")
	exprStr = strings.TrimSpace(declRaw[eqIdx+1:])
	if exprStr == "" {
		return nil, 0, unifiederrors.ParseErrorf("invalid variable declaration: missing expression for %q", varName)
	}
	if strings.ContainsAny(exprStr, "\n\r") {
		exprStr = collapseWhitespaceOutsideStrings(exprStr)
	}

	return &VarDeclaration{
		Name:       varName,
		Expression: exprStr,
	}, len(declLines), nil
}

func parseFunctionDeclarationLine(trimmedLine string) (string, []ParamDeclaration, string, error) {
	rest := strings.TrimPrefix(trimmedLine, "fun ")

	parenIdx := strings.Index(rest, "(")
	if parenIdx < 0 {
		return "", nil, "", unifiederrors.ParseError("invalid function declaration: missing parameter list")
	}
	fnName := strings.TrimSpace(rest[:parenIdx])
	if fnName == "" {
		return "", nil, "", unifiederrors.ParseError("invalid function declaration: missing function name")
	}

	closeParenIdx := strings.Index(rest, ")")
	if closeParenIdx < 0 || closeParenIdx < parenIdx {
		return "", nil, "", unifiederrors.ParseError("invalid function declaration: missing closing parenthesis")
	}

	paramStr := rest[parenIdx+1 : closeParenIdx]
	var params []ParamDeclaration
	if strings.TrimSpace(paramStr) != "" {
		for _, p := range strings.Split(paramStr, ",") {
			trimmed := strings.TrimSpace(p)
			paramType := ""
			if colonIdx := strings.Index(trimmed, ":"); colonIdx >= 0 {
				paramType = strings.TrimSpace(trimmed[colonIdx+1:])
				trimmed = strings.TrimSpace(trimmed[:colonIdx])
			}
			if trimmed == "" {
				continue
			}
			params = append(params, ParamDeclaration{Name: trimmed, Type: paramType})
		}
	}

	afterParams := strings.TrimSpace(rest[closeParenIdx+1:])
	eqIdx := strings.Index(afterParams, "=")
	if eqIdx < 0 {
		return "", nil, "", unifiederrors.ParseError("invalid function declaration: missing '=' after parameters")
	}

	bodyStr := strings.TrimSpace(afterParams[eqIdx+1:])
	return fnName, params, bodyStr, nil
}

func parseFunctionDeclarationFromLines(lines []string, start int) (*FunctionDeclaration, int, error) {
	trimmedLine := strings.TrimSpace(lines[start])
	fnName, params, bodyStr, err := parseFunctionDeclarationLine(trimmedLine)
	if err != nil {
		return nil, 0, err
	}

	if bodyStr != "" {
		strippedBody := preprocessor.StripSingleLineComment(bodyStr)
		if strings.TrimSpace(strippedBody) != "" && isDelimiterBalanced(strippedBody) {
			return &FunctionDeclaration{
				Name:   fnName,
				Params: params,
				Body:   strippedBody,
			}, 1, nil
		}
	}

	var bodyLines []string
	if bodyStr != "" {
		bodyLines = append(bodyLines, bodyStr)
	}
	i := start + 1
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			bodyLines = append(bodyLines, "")
			i++
			continue
		}
		if strings.HasPrefix(trimmed, "//") {
			i++
			continue
		}
		if isDirectiveLine(trimmed) {
			candidate := strings.TrimSpace(strings.Join(bodyLines, "\n"))
			if candidate != "" {
				candidate = preprocessor.StripLineComments(candidate)
				if isDelimiterBalanced(candidate) {
					break
				}
			}
		}
		bodyLines = append(bodyLines, line)
		i++
	}

	bodyStr = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	bodyStr = preprocessor.StripLineComments(bodyStr)
	return &FunctionDeclaration{
		Name:   fnName,
		Params: params,
		Body:   bodyStr,
	}, i - start, nil
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

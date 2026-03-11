package runner

import (
	"context"
	goparser "go/parser"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/preprocessor"
	"infomunge/internal/stringutils"
	"strings"
)

// parseVarDecl parses a variable declaration line and evaluates its value.
// baseOffset is the offset of the start of this line within fullRaw.
// Returns the evaluated value, variable name, and any error.
func parseVarDecl(line, trimmedLine string, baseOffset int, evalCtx map[string]interface{}, fullRaw string) (interface{}, string, error) {
	return parseVarDeclWithGoContext(line, trimmedLine, baseOffset, evalCtx, context.Background(), fullRaw)
}

func parseVarDeclWithGoContext(line, trimmedLine string, baseOffset int, evalCtx map[string]interface{}, goCtx context.Context, fullRaw string) (interface{}, string, error) {
	parts := strings.SplitN(trimmedLine, "=", 2)
	if len(parts) != 2 {
		return nil, "", unifiederrors.ParseErrorf("invalid variable declaration: missing '=' in %q", trimmedLine)
	}

	declParts := strings.Fields(parts[0])
	if len(declParts) != 2 || declParts[0] != "var" {
		return nil, "", unifiederrors.ParseErrorf("invalid variable declaration: expected 'var <name> = <expression>' in %q", trimmedLine)
	}

	varName := declParts[1]
	if strings.TrimSpace(varName) == "" {
		return nil, "", unifiederrors.ParseErrorf("invalid variable declaration: missing variable name in %q", trimmedLine)
	}
	exprStr := strings.TrimSpace(parts[1])
	if exprStr == "" {
		return nil, "", unifiederrors.ParseErrorf("invalid variable declaration: missing expression for %q", varName)
	}

	parseableVal, mapping, err := preprocessor.PrepareForParsing(exprStr, preprocessor.Options{})
	if err != nil {
		return nil, "", unifiederrors.WrapParsef(err, "preprocessing error in variable declaration: %v", err)
	}

	eqIdx := strings.Index(line, "=")
	if eqIdx < 0 {
		return nil, "", unifiederrors.InternalError("internal error: malformed variable declaration")
	}
	lineIdx := eqIdx + 1
	leadingSpace := len(parts[1]) - len(strings.TrimLeft(parts[1], " \t"))
	valOffset := baseOffset + lineIdx + leadingSpace

	val, err := evaluator.EvaluateWithGoContext(parseableVal, evalCtx, goCtx, mapping, valOffset, fullRaw)
	if err != nil {
		return nil, "", err
	}

	return val, varName, nil
}

// parseVarDeclFromLines parses a variable declaration that may span multiple lines.
// baseOffset is the offset of the start of lines[start] within fullRaw.
// Returns the evaluated value, variable name, number of lines consumed, and any error.
func parseVarDeclFromLines(lines []string, start int, baseOffset int, evalCtx map[string]interface{}, fullRaw string) (interface{}, string, int, error) {
	return parseVarDeclFromLinesWithGoContext(lines, start, baseOffset, evalCtx, context.Background(), fullRaw)
}

type parsedVarDecl struct {
	varName  string
	exprStr  string
	consumed int
}

func parseVarDeclSource(lines []string, start int) (*parsedVarDecl, error) {
	if start >= len(lines) {
		return nil, nil
	}

	firstLine := lines[start]
	trimmedFirst := strings.TrimSpace(firstLine)
	if !strings.HasPrefix(trimmedFirst, "var ") {
		return nil, nil
	}

	rest := strings.TrimSpace(strings.TrimPrefix(trimmedFirst, "var "))
	if rest == "" {
		return nil, unifiederrors.ParseErrorf("invalid variable declaration: missing variable name in %q", trimmedFirst)
	}

	namePart := rest
	if eqIdx := strings.Index(rest, "="); eqIdx >= 0 {
		namePart = strings.TrimSpace(rest[:eqIdx])
	}
	fields := strings.Fields(namePart)
	if len(fields) != 1 {
		return nil, unifiederrors.ParseErrorf("invalid variable declaration: expected a single variable name in %q", trimmedFirst)
	}
	varName := fields[0]

	var declLines []string
	declLines = append(declLines, firstLine)
	i := start + 1
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			declLines = append(declLines, line)
			i++
			continue
		}
		if isDirectiveLine(trimmed) {
			break
		}
		declLines = append(declLines, line)
		i++
	}

	declRaw := strings.Join(declLines, "\n")
	eqIdx := strings.Index(declRaw, "=")
	if eqIdx < 0 {
		return nil, unifiederrors.ParseErrorf("invalid variable declaration: missing '=' in %q", trimmedFirst)
	}
	exprStr := strings.TrimSpace(declRaw[eqIdx+1:])
	if exprStr == "" {
		return nil, unifiederrors.ParseErrorf("invalid variable declaration: missing expression for %q", varName)
	}
	if strings.ContainsAny(exprStr, "\n\r") {
		exprStr = collapseWhitespaceOutsideStrings(exprStr)
	}

	return &parsedVarDecl{
		varName:  varName,
		exprStr:  exprStr,
		consumed: len(declLines),
	}, nil
}

func parseVarDeclFromLinesWithGoContext(lines []string, start int, baseOffset int, evalCtx map[string]interface{}, goCtx context.Context, fullRaw string) (interface{}, string, int, error) {
	spec, err := parseVarDeclSource(lines, start)
	if err != nil {
		consumed := 1
		if spec != nil && spec.consumed > 0 {
			consumed = spec.consumed
		}
		return nil, "", consumed, err
	}
	if spec == nil {
		return nil, "", 0, nil
	}

	parseableVal, mapping, err := preprocessor.PrepareForParsing(spec.exprStr, preprocessor.Options{})
	if err != nil {
		return nil, "", spec.consumed, unifiederrors.WrapParsef(err, "preprocessing error in variable declaration: %v", err)
	}

	declRaw := strings.Join(lines[start:start+spec.consumed], "\n")
	eqIdx := strings.Index(declRaw, "=")
	exprStart := eqIdx + 1
	for exprStart < len(declRaw) {
		ch := declRaw[exprStart]
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
		exprStart++
	}
	valOffset := baseOffset + exprStart

	val, err := evaluator.EvaluateWithGoContext(parseableVal, evalCtx, goCtx, mapping, valOffset, fullRaw)
	if err != nil {
		return nil, "", spec.consumed, err
	}

	return val, spec.varName, spec.consumed, nil
}

func parseFunDeclLine(trimmedLine string) (string, []evaluator.ParamDef, string, error) {
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
	var params []evaluator.ParamDef
	if strings.TrimSpace(paramStr) != "" {
		for _, p := range strings.Split(paramStr, ",") {
			trimmed := strings.TrimSpace(p)
			if colonIdx := strings.Index(trimmed, ":"); colonIdx >= 0 {
				trimmed = strings.TrimSpace(trimmed[:colonIdx])
			}
			if trimmed == "" {
				continue
			}
			params = append(params, evaluator.ParamDef{
				Name:         trimmed,
				ExpectedKind: evaluator.KindUnknown,
			})
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

func buildFunLambda(params []evaluator.ParamDef, bodyStr string, env map[string]interface{}) (*evaluator.Lambda, error) {
	if bodyStr == "" {
		return nil, unifiederrors.ParseError("invalid function declaration: empty body")
	}

	prepOpts := preprocessor.Options{}
	if strings.ContainsAny(bodyStr, "\n\r") {
		prepOpts.AllowMultilineIfElse = true
	}
	parseableBody, _, err := preprocessor.PrepareForParsing(bodyStr, prepOpts)
	if err != nil {
		return nil, unifiederrors.WrapParse(err, "preprocessing error in function body")
	}

	bodyAST, err := goparser.ParseExpr(parseableBody)
	if err != nil {
		return nil, unifiederrors.WrapParse(err, "invalid function body")
	}

	return &evaluator.Lambda{
		Params:  params,
		Body:    bodyStr,
		BodyAST: bodyAST,
		Env:     env,
	}, nil
}

func isDirectiveLine(trimmedLine string) bool {
	for _, kw := range directiveKeywords {
		if strings.HasPrefix(trimmedLine, kw) {
			return true
		}
	}
	return false
}

func parseFunDeclFromLines(lines []string, start int, env map[string]interface{}) (*evaluator.Lambda, string, int, error) {
	spec, err := parseFunDeclSource(lines, start)
	if err != nil {
		return nil, "", 0, err
	}
	if spec == nil {
		return nil, "", 0, nil
	}

	fn, err := buildFunLambda(spec.params, spec.bodyStr, env)
	if err != nil {
		return nil, "", 0, err
	}
	return fn, spec.fnName, spec.consumed, nil
}

type parsedFunDecl struct {
	fnName   string
	params   []evaluator.ParamDef
	bodyStr  string
	consumed int
}

func parseFunDeclSource(lines []string, start int) (*parsedFunDecl, error) {
	trimmedLine := strings.TrimSpace(lines[start])
	fnName, params, bodyStr, err := parseFunDeclLine(trimmedLine)
	if err != nil {
		return nil, err
	}

	// If body is complete on one line (non-empty and balanced delimiters), build immediately
	if bodyStr != "" {
		strippedBody := stripSingleLineComment(bodyStr)
		if strings.TrimSpace(strippedBody) != "" && isDelimiterBalanced(strippedBody) {
			return &parsedFunDecl{
				fnName:   fnName,
				params:   params,
				bodyStr:  strippedBody,
				consumed: 1,
			}, nil
		}
	}

	// Collect continuation lines for multi-line body
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
		// Skip standalone comment lines
		if strings.HasPrefix(trimmed, "//") {
			i++
			continue
		}
		if isDirectiveLine(trimmed) {
			candidate := strings.TrimSpace(strings.Join(bodyLines, "\n"))
			if candidate != "" {
				candidate = stripLineComments(candidate)
				if isDelimiterBalanced(candidate) {
					break
				}
			}
		}
		bodyLines = append(bodyLines, line)
		i++
	}

	bodyStr = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	bodyStr = stripLineComments(bodyStr)
	return &parsedFunDecl{
		fnName:   fnName,
		params:   params,
		bodyStr:  bodyStr,
		consumed: i - start,
	}, nil
}

func collapseWhitespaceOutsideStrings(input string) string {
	var b strings.Builder
	var sc stringutils.ScanState
	lastWasSpace := false

	for i := 0; i < len(input); i++ {
		ch := input[i]
		sc.Advance(ch)

		if sc.InString() {
			b.WriteByte(ch)
			lastWasSpace = false
			continue
		}

		switch ch {
		case ' ', '\t', '\n', '\r':
			if !lastWasSpace {
				b.WriteByte(' ')
				lastWasSpace = true
			}
		default:
			b.WriteByte(ch)
			lastWasSpace = false
		}
	}

	return strings.TrimSpace(b.String())
}

// isDelimiterBalanced checks if brackets, braces, and parentheses are balanced,
// respecting string literals.
func isDelimiterBalanced(s string) bool {
	var sc stringutils.ScanState
	for i := 0; i < len(s); i++ {
		sc.Advance(s[i])
	}
	return sc.Depth() == 0
}

// stripLineComments removes // line comments from each line of the body,
// respecting string literals. Line count is preserved for diagnostics.
func stripLineComments(body string) string {
	lines := strings.Split(body, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		stripped := stripSingleLineComment(line)
		result = append(result, stripped)
	}
	return strings.Join(result, "\n")
}

// stripSingleLineComment removes a // comment from a single line, respecting string literals.
func stripSingleLineComment(line string) string {
	inString := false
	stringChar := byte(0)
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if !inString && (ch == '"' || ch == '\'') {
			inString = true
			stringChar = ch
			continue
		}
		if inString && ch == stringChar {
			inString = false
			continue
		}
		if !inString && ch == '/' {
			if i+1 < len(line) && line[i+1] == '/' {
				return line[:i] + strings.Repeat(" ", len(line)-i)
			}
			if preprocessor.IsRegexContext(line, i) {
				regexEnd, _, _ := preprocessor.ParseRegexLiteral(line, i)
				if regexEnd > i {
					i = regexEnd - 1
					continue
				}
			}
		}
	}
	return line
}

// parseFunDecl parses a function declaration line like "fun toUser(user) = {firstName: user.name}"
// and returns a Lambda. If env is non-nil, it's attached to the Lambda for module-local scope.
func parseFunDecl(trimmedLine string, env map[string]interface{}) (*evaluator.Lambda, string, error) {
	fnName, params, bodyStr, err := parseFunDeclLine(trimmedLine)
	if err != nil {
		return nil, "", err
	}

	bodyStr = stripSingleLineComment(bodyStr)
	fn, err := buildFunLambda(params, bodyStr, env)
	if err != nil {
		return nil, "", err
	}

	return fn, fnName, nil
}

// parseTypeDecl parses a type declaration line like "type Currency = String { format: "##.00" }"
// and returns a TypeDef with properties.
func parseTypeDecl(trimmedLine string) (*evaluator.TypeDef, string, error) {
	rest := strings.TrimPrefix(trimmedLine, "type ")

	eqIdx := strings.Index(rest, "=")
	if eqIdx < 0 {
		return nil, "", unifiederrors.ParseError("invalid type declaration: missing '='")
	}

	typeName := strings.TrimSpace(rest[:eqIdx])
	if typeName == "" {
		return nil, "", unifiederrors.ParseError("invalid type declaration: missing type name")
	}

	rhs := strings.TrimSpace(rest[eqIdx+1:])
	if rhs == "" {
		return nil, "", unifiederrors.ParseError("invalid type declaration: missing base type")
	}

	var baseType string
	var properties map[string]interface{}

	braceIdx := strings.Index(rhs, "{")
	if braceIdx < 0 {
		baseType = rhs
	} else {
		baseType = strings.TrimSpace(rhs[:braceIdx])
		propsStr := rhs[braceIdx:]

		if !strings.HasSuffix(strings.TrimSpace(propsStr), "}") {
			return nil, "", unifiederrors.ParseError("invalid type declaration: unclosed properties block")
		}

		var err error
		properties, err = parseTypeProperties(propsStr)
		if err != nil {
			return nil, "", unifiederrors.WrapParse(err, "invalid type declaration")
		}
	}

	if baseType == "" {
		return nil, "", unifiederrors.ParseError("invalid type declaration: missing base type")
	}

	return &evaluator.TypeDef{
		Name:       typeName,
		BaseType:   baseType,
		Properties: properties,
	}, typeName, nil
}

// parseTypeProperties parses a properties block like "{ format: \"##.00\", locale: \"en_US\" }".
func parseTypeProperties(propsStr string) (map[string]interface{}, error) {
	propsStr = strings.TrimSpace(propsStr)
	if !strings.HasPrefix(propsStr, "{") || !strings.HasSuffix(propsStr, "}") {
		return nil, unifiederrors.ParseError("properties must be enclosed in braces")
	}

	inner := strings.TrimSpace(propsStr[1 : len(propsStr)-1])
	if inner == "" {
		return make(map[string]interface{}), nil
	}

	properties := make(map[string]interface{})
	pairs := splitPropertyPairs(inner)

	for _, pair := range pairs {
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

		value, err := parsePropertyValue(valueStr)
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

// parsePropertyValue parses a property value (string, number, or boolean).
func parsePropertyValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) >= 2 {
		return s[1 : len(s)-1], nil
	}

	if b, ok := evaluator.ParseBoolLiteral(s); ok {
		return b, nil
	}

	if s == "null" || s == "nil" {
		return nil, nil
	}

	if num, ok := evaluator.ParseNumericLiteral(s); ok {
		return num, nil
	}

	return nil, unifiederrors.ParseErrorf("cannot parse value: %s", s)
}

// parseNamespaceDecl parses a namespace declaration line.
// Format: "ns prefix uri" for prefixed namespace or "ns uri" for default namespace
// Returns the prefix (empty string for default) and the URI.
func parseNamespaceDecl(trimmedLine string) (string, string, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "ns "))
	parts := strings.Fields(rest)

	if len(parts) == 1 {
		// Default namespace: ns http://www.abc.com
		return "", parts[0], nil
	} else if len(parts) == 2 {
		// Prefixed namespace: ns ns0 http://www.abc.com
		return parts[0], parts[1], nil
	}

	return "", "", unifiederrors.ParseError("invalid namespace declaration: expected 'ns [prefix] uri'")
}

// handleImport processes an import directive and adds symbols to the context.
func handleImport(line string, context map[string]interface{}, loader *ModuleLoader) error {
	if loader == nil {
		return unifiederrors.ParseError("imports are not available in this context")
	}

	rest := strings.TrimSpace(strings.TrimPrefix(line, "import "))

	if !strings.Contains(rest, " from ") {
		m, err := loader.Load(rest)
		if err != nil {
			return err
		}
		// Convert typed Namespace to raw map for evaluation context
		context[m.Name] = m.Namespace.ToContext()
		return nil
	}

	parts := strings.SplitN(rest, " from ", 2)
	namesPart := strings.TrimSpace(parts[0])
	moduleSpec := strings.TrimSpace(parts[1])

	m, err := loader.Load(moduleSpec)
	if err != nil {
		return err
	}

	// Convert typed Namespace to raw map for evaluation context
	context[m.Name] = m.Namespace.ToContext()

	if namesPart == "*" {
		for k, entry := range m.Namespace {
			if strings.HasPrefix(k, "_") {
				continue
			}
			if _, exists := context[k]; exists && k != m.Name {
				return unifiederrors.ParseErrorf("import * from %s: name %q already defined", moduleSpec, k)
			}
			context[k] = entry.Value
		}
		return nil
	}

	for _, name := range strings.Split(namesPart, ",") {
		n := strings.TrimSpace(name)
		if n == "" {
			continue
		}
		entry, ok := m.Namespace[n]
		if !ok {
			return unifiederrors.ParseErrorf("import %s from %s: symbol %q not found", namesPart, moduleSpec, n)
		}
		if _, exists := context[n]; exists {
			return unifiederrors.ParseErrorf("import %s from %s: name %q already defined", namesPart, moduleSpec, n)
		}
		context[n] = entry.Value
	}

	return nil
}

package runner

import (
	"context"
	goparser "go/parser"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/preprocessor"
	"infomunge/internal/sourcemap"
	"infomunge/internal/stringutils"
	"strings"
)

// parseVarDecl parses a variable declaration line and evaluates its value.
// baseOffset is the offset of the start of this line within fullRaw.
// Returns the evaluated value, variable name, and any error.
func parseVarDecl(line, trimmedLine string, baseOffset int, evalCtx evaluator.Context, fullRaw string) (evaluator.Value, string, error) {
	return parseVarDeclWithGoContext(line, trimmedLine, baseOffset, evalCtx, context.Background(), fullRaw)
}

func parseVarDeclWithGoContext(line, trimmedLine string, baseOffset int, evalCtx evaluator.Context, goCtx context.Context, fullRaw string) (evaluator.Value, string, error) {
	return parseVarDeclWithScope(line, trimmedLine, baseOffset, evaluator.NewScope(evalCtx).WithGoContext(goCtx), fullRaw)
}

func parseVarDeclWithScope(line, trimmedLine string, baseOffset int, scope *evaluator.Scope, fullRaw string) (evaluator.Value, string, error) {
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
	if evaluator.IsReservedBindingName(varName) {
		return nil, "", unifiederrors.ParseErrorf("%q is reserved for runtime metadata", varName)
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
	exprMap := sourceMapForDeclExpr(fullRaw, baseOffset, line, eqIdx+1, false)
	val, err := evaluator.EvaluateWithScopeAndContext(parseableVal, scope, &evaluator.ErrorContext{SourceMap: exprMap.Compose(parseableVal, mapping)})
	if err != nil {
		return nil, "", err
	}

	return val, varName, nil
}

// parseVarDeclFromLines parses a variable declaration that may span multiple lines.
// baseOffset is the offset of the start of lines[start] within fullRaw.
// Returns the evaluated value, variable name, number of lines consumed, and any error.
func parseVarDeclFromLines(lines []string, start int, baseOffset int, evalCtx evaluator.Context, fullRaw string) (evaluator.Value, string, int, error) {
	return parseVarDeclFromLinesWithGoContext(lines, start, baseOffset, evalCtx, context.Background(), fullRaw)
}

type parsedVarDecl struct {
	varName  string
	exprStr  string
	consumed int
}

func parseVarDeclSource(lines []string, start int) (*parsedVarDecl, error) {
	decl, consumed, err := parseVarDeclarationFromLines(lines, start)
	if err != nil || decl == nil {
		return nil, err
	}
	if evaluator.IsReservedBindingName(decl.Name) {
		return nil, unifiederrors.ParseErrorf("%q is reserved for runtime metadata", decl.Name)
	}
	return &parsedVarDecl{
		varName:  decl.Name,
		exprStr:  decl.Expression,
		consumed: consumed,
	}, nil
}

func parsedVarLinesAreComplete(lines []string) bool {
	declRaw := strings.Join(lines, "\n")
	eqIdx := strings.Index(declRaw, "=")
	if eqIdx < 0 {
		return false
	}
	exprStr := strings.TrimSpace(declRaw[eqIdx+1:])
	return exprStr != "" && isDelimiterBalanced(exprStr)
}

func parseVarDeclFromLinesWithGoContext(lines []string, start int, baseOffset int, evalCtx evaluator.Context, goCtx context.Context, fullRaw string) (evaluator.Value, string, int, error) {
	return parseVarDeclFromLinesWithScope(lines, start, baseOffset, evaluator.NewScope(evalCtx).WithGoContext(goCtx), fullRaw)
}

func parseVarDeclFromLinesWithScope(lines []string, start int, baseOffset int, scope *evaluator.Scope, fullRaw string) (evaluator.Value, string, int, error) {
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
	exprMap := sourceMapForDeclExpr(fullRaw, baseOffset, declRaw, eqIdx+1, true)
	if strings.ContainsAny(spec.exprStr, "\n\r") {
		exprMap = collapseGeneratedWhitespaceMap(exprMap)
	}
	val, err := evaluator.EvaluateWithScopeAndContext(parseableVal, scope, &evaluator.ErrorContext{SourceMap: exprMap.Compose(parseableVal, mapping)})
	if err != nil {
		return nil, "", spec.consumed, err
	}

	return val, spec.varName, spec.consumed, nil
}

func evaluateVarDeclaration(decl *VarDeclaration, source DeclarationSource, scope *evaluator.Scope, fullRaw string) (evaluator.Value, error) {
	if decl == nil {
		return nil, unifiederrors.InternalError("internal error: missing variable declaration")
	}
	if evaluator.IsReservedBindingName(decl.Name) {
		return nil, unifiederrors.ParseErrorf("%q is reserved for runtime metadata", decl.Name)
	}
	if strings.TrimSpace(decl.Expression) == "" {
		return nil, unifiederrors.ParseErrorf("invalid variable declaration: missing expression for %q", decl.Name)
	}

	parseableVal, mapping, err := preprocessor.PrepareForParsing(decl.Expression, preprocessor.Options{})
	if err != nil {
		return nil, unifiederrors.WrapParsef(err, "preprocessing error in variable declaration: %v", err)
	}

	declRaw := strings.Join(source.Lines, "\n")
	eqIdx := strings.Index(declRaw, "=")
	if eqIdx < 0 {
		return nil, unifiederrors.InternalError("internal error: malformed variable declaration")
	}
	exprMap := sourceMapForDeclExpr(fullRaw, source.Offset, declRaw, eqIdx+1, true)
	if strings.ContainsAny(decl.Expression, "\n\r") {
		exprMap = collapseGeneratedWhitespaceMap(exprMap)
	}
	return evaluator.EvaluateWithScopeAndContext(parseableVal, scope, &evaluator.ErrorContext{SourceMap: exprMap.Compose(parseableVal, mapping)})
}

func parseFunDeclLine(trimmedLine string) (string, []evaluator.ParamDef, string, error) {
	fnName, params, bodyStr, err := parseFunctionDeclarationLine(trimmedLine)
	if err != nil {
		return "", nil, "", err
	}
	if evaluator.IsReservedBindingName(fnName) {
		return "", nil, "", unifiederrors.ParseErrorf("%q is reserved for runtime metadata", fnName)
	}
	return fnName, paramDeclarationsToParamDefs(params), bodyStr, nil
}

func buildFunLambda(params []evaluator.ParamDef, bodyStr string, env evaluator.Context, bodyMap *sourcemap.Map) (*evaluator.Lambda, error) {
	if bodyStr == "" {
		return nil, unifiederrors.ParseError("invalid function declaration: empty body")
	}

	prepOpts := preprocessor.Options{}
	if strings.ContainsAny(bodyStr, "\n\r") {
		prepOpts.AllowMultilineIfElse = true
	}
	parseableBody, mapping, err := preprocessor.PrepareForParsing(bodyStr, prepOpts)
	if err != nil {
		return nil, unifiederrors.WrapParse(err, "preprocessing error in function body")
	}

	bodyAST, err := goparser.ParseExpr(parseableBody)
	if err != nil {
		if bodyMap != nil {
			return nil, bodyMap.Compose(parseableBody, mapping).FormatParseError(err)
		}
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

func parseFunDeclFromLines(lines []string, start int, env evaluator.Context) (*evaluator.Lambda, string, int, error) {
	return parseFunDeclFromLinesWithSource(lines, start, env, "", 0)
}

func parseFunDeclFromLinesWithSource(lines []string, start int, env evaluator.Context, fullRaw string, baseOffset int) (*evaluator.Lambda, string, int, error) {
	spec, err := parseFunDeclSource(lines, start)
	if err != nil {
		return nil, "", 0, err
	}
	if spec == nil {
		return nil, "", 0, nil
	}

	var bodyMap *sourcemap.Map
	if fullRaw != "" {
		declRaw := strings.Join(lines[start:start+spec.consumed], "\n")
		m := sourceMapForDeclExpr(fullRaw, baseOffset, declRaw, strings.Index(declRaw, "=")+1, false)
		bodyMap = &m
	}

	fn, err := buildFunLambda(spec.params, spec.bodyStr, env, bodyMap)
	if err != nil {
		return nil, "", 0, err
	}
	return fn, spec.fnName, spec.consumed, nil
}

func paramDeclarationsToParamDefs(params []ParamDeclaration) []evaluator.ParamDef {
	defs := make([]evaluator.ParamDef, 0, len(params))
	for _, param := range params {
		defs = append(defs, evaluator.ParamDef{
			Name:         param.Name,
			ExpectedKind: evaluator.KindUnknown,
		})
	}
	return defs
}

func buildFunctionDeclaration(decl *FunctionDeclaration, source DeclarationSource, env evaluator.Context, fullRaw string) (*evaluator.Lambda, error) {
	if decl == nil {
		return nil, unifiederrors.InternalError("internal error: missing function declaration")
	}
	if evaluator.IsReservedBindingName(decl.Name) {
		return nil, unifiederrors.ParseErrorf("%q is reserved for runtime metadata", decl.Name)
	}

	var bodyMap *sourcemap.Map
	if fullRaw != "" {
		declRaw := strings.Join(source.Lines, "\n")
		m := sourceMapForDeclExpr(fullRaw, source.Offset, declRaw, strings.Index(declRaw, "=")+1, false)
		bodyMap = &m
	}

	return buildFunLambda(paramDeclarationsToParamDefs(decl.Params), decl.Body, env, bodyMap)
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
		strippedBody := preprocessor.StripSingleLineComment(bodyStr)
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
	return &parsedFunDecl{
		fnName:   fnName,
		params:   params,
		bodyStr:  bodyStr,
		consumed: i - start,
	}, nil
}

func collapseWhitespaceOutsideStrings(input string) string {
	collapsed, _ := collapseWhitespaceOutsideStringsWithMapping(input)
	return collapsed
}

func collapseWhitespaceOutsideStringsWithMapping(input string) (string, []int) {
	var b strings.Builder
	var sc stringutils.ScanState
	lastWasSpace := false
	var mapping []int

	for i := 0; i < len(input); i++ {
		ch := input[i]
		sc.Advance(ch)

		if sc.InString() {
			b.WriteByte(ch)
			mapping = append(mapping, i)
			lastWasSpace = false
			continue
		}

		switch ch {
		case ' ', '\t', '\n', '\r':
			if !lastWasSpace {
				b.WriteByte(' ')
				mapping = append(mapping, i)
				lastWasSpace = true
			}
		default:
			b.WriteByte(ch)
			mapping = append(mapping, i)
			lastWasSpace = false
		}
	}

	collapsed := b.String()
	start := 0
	for start < len(collapsed) && collapsed[start] == ' ' {
		start++
	}
	end := len(collapsed)
	for end > start && collapsed[end-1] == ' ' {
		end--
	}
	return collapsed[start:end], mapping[start:end]
}

func sourceMapForDeclExpr(fullRaw string, baseOffset int, declRaw string, exprStart int, collapseWhitespace bool) sourcemap.Map {
	exprEnd := len(declRaw)
	for exprStart < exprEnd {
		ch := declRaw[exprStart]
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
		exprStart++
	}
	for exprEnd > exprStart {
		ch := declRaw[exprEnd-1]
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
		exprEnd--
	}
	exprMap := sourcemap.Identity(fullRaw).SliceSource(baseOffset+exprStart, baseOffset+exprEnd)
	if !collapseWhitespace {
		return exprMap
	}
	return collapseGeneratedWhitespaceMap(exprMap)
}

func collapseGeneratedWhitespaceMap(exprMap sourcemap.Map) sourcemap.Map {
	collapsed, mapping := collapseWhitespaceOutsideStringsWithMapping(exprMap.Generated())
	return exprMap.Compose(collapsed, mapping)
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

// parseFunDecl parses a function declaration line like "fun toUser(user) = {firstName: user.name}"
// and returns a Lambda. If env is non-nil, it's attached to the Lambda for module-local scope.
func parseFunDecl(trimmedLine string, env evaluator.Context) (*evaluator.Lambda, string, error) {
	fnName, params, bodyStr, err := parseFunDeclLine(trimmedLine)
	if err != nil {
		return nil, "", err
	}

	bodyStr = preprocessor.StripSingleLineComment(bodyStr)
	fn, err := buildFunLambda(params, bodyStr, env, nil)
	if err != nil {
		return nil, "", err
	}

	return fn, fnName, nil
}

// parseTypeDecl parses a type declaration line like "type Currency = String { format: "##.00" }"
// and returns a TypeDef with properties.
func parseTypeDecl(trimmedLine string) (*evaluator.TypeDef, string, error) {
	decl, err := parseTypeDeclaration(trimmedLine)
	if err != nil {
		return nil, "", err
	}
	typeDef, err := bindTypeDeclaration(decl)
	if err != nil {
		return nil, "", err
	}
	return typeDef, decl.Name, nil
}

func bindTypeDeclaration(decl *TypeDeclaration) (*evaluator.TypeDef, error) {
	if decl == nil {
		return nil, unifiederrors.InternalError("internal error: missing type declaration")
	}
	if evaluator.IsReservedBindingName(decl.Name) {
		return nil, unifiederrors.ParseErrorf("%q is reserved for runtime metadata", decl.Name)
	}
	return &evaluator.TypeDef{
		Name:       decl.Name,
		BaseType:   decl.BaseType,
		Properties: evaluator.Object(decl.Properties),
	}, nil
}

func applyTypeDeclaration(decl *TypeDeclaration, context evaluator.Context) error {
	typeDef, err := bindTypeDeclaration(decl)
	if err != nil {
		return err
	}
	if decl.Name != "" {
		context[decl.Name] = typeDef
	}
	return nil
}

// parseTypeProperties parses a properties block like "{ format: \"##.00\", locale: \"en_US\" }".
func parseTypeProperties(propsStr string) (evaluator.Context, error) {
	propsStr = strings.TrimSpace(propsStr)
	if !strings.HasPrefix(propsStr, "{") || !strings.HasSuffix(propsStr, "}") {
		return nil, unifiederrors.ParseError("properties must be enclosed in braces")
	}

	inner := strings.TrimSpace(propsStr[1 : len(propsStr)-1])
	if inner == "" {
		return make(evaluator.Context), nil
	}

	properties := make(evaluator.Context)
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
func parsePropertyValue(s string) (evaluator.Value, error) {
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
func handleImport(line string, context evaluator.Context, loader *ModuleLoader) error {
	decl := parseImportDeclaration(line)
	return applyImportDeclaration(&decl, context, loader)
}

func applyImportDeclaration(decl *ImportDeclaration, context evaluator.Context, loader *ModuleLoader) error {
	if loader == nil {
		return unifiederrors.ParseError("imports are not available in this context")
	}
	if decl == nil {
		return unifiederrors.InternalError("internal error: missing import declaration")
	}

	if decl.NamespaceOnly {
		m, err := loader.Load(decl.ModuleSpec)
		if err != nil {
			return err
		}
		// Convert typed Namespace to raw map for evaluation context
		context[m.Name] = m.Namespace.ToContext()
		return nil
	}

	m, err := loader.Load(decl.ModuleSpec)
	if err != nil {
		return err
	}

	// Convert typed Namespace to raw map for evaluation context
	context[m.Name] = m.Namespace.ToContext()

	if decl.Star {
		for k, entry := range m.Namespace {
			if strings.HasPrefix(k, "_") {
				continue
			}
			if _, exists := context[k]; exists && k != m.Name {
				return unifiederrors.ParseErrorf("import * from %s: name %q already defined", decl.ModuleSpec, k)
			}
			context[k] = entry.Value
		}
		return nil
	}

	for _, n := range decl.Names {
		entry, ok := m.Namespace[n]
		if !ok {
			return unifiederrors.ParseErrorf("import %s from %s: symbol %q not found", decl.NamesPart, decl.ModuleSpec, n)
		}
		if _, exists := context[n]; exists {
			return unifiederrors.ParseErrorf("import %s from %s: name %q already defined", decl.NamesPart, decl.ModuleSpec, n)
		}
		context[n] = entry.Value
	}

	return nil
}

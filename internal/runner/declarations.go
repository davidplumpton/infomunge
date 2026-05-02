package runner

import (
	goparser "go/parser"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/sourcemap"
	"infomunge/internal/stringutils"
	"strings"
)

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

	parseableVal, mapping, err := prepareExpressionForParsing(decl.Expression)
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

func buildFunLambda(params []evaluator.ParamDef, bodyStr string, env evaluator.Context, bodyMap *sourcemap.Map) (*evaluator.Lambda, error) {
	if bodyStr == "" {
		return nil, unifiederrors.ParseError("invalid function declaration: empty body")
	}

	parseableBody, mapping, err := prepareExpressionForParsing(bodyStr)
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

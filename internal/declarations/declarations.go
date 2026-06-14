package declarations

import (
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/preprocessor"
	"infomunge/internal/stringutils"
)

// DirectiveKeywords are declaration starters that delimit multiline var/fun declarations.
var DirectiveKeywords = []string{"output ", "input ", "%dw ", "%im ", "ns ", "import ", "var ", "fun ", "type "}

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

func IsDirectiveLine(trimmedLine string) bool {
	for _, kw := range DirectiveKeywords {
		if strings.HasPrefix(trimmedLine, kw) {
			return true
		}
	}
	return false
}

func ParseVarDeclarationFromLines(lines []string, start int) (*VarDeclaration, int, error) {
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
		if strings.TrimSpace(strippedExpr) != "" && IsDelimiterBalanced(strippedExpr) {
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
		if IsDirectiveLine(trimmed) || (trimmed == "---" && parsedVarLinesAreComplete(declLines)) {
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
		exprStr = CollapseWhitespaceOutsideStrings(exprStr)
	}

	return &VarDeclaration{
		Name:       varName,
		Expression: exprStr,
	}, len(declLines), nil
}

func parsedVarLinesAreComplete(lines []string) bool {
	declRaw := strings.Join(lines, "\n")
	eqIdx := strings.Index(declRaw, "=")
	if eqIdx < 0 {
		return false
	}
	exprStr := strings.TrimSpace(declRaw[eqIdx+1:])
	return exprStr != "" && IsDelimiterBalanced(exprStr)
}

func ParseFunctionDeclarationLine(trimmedLine string) (string, []ParamDeclaration, string, error) {
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

func ParseFunctionDeclarationFromLines(lines []string, start int) (*FunctionDeclaration, int, error) {
	trimmedLine := strings.TrimSpace(lines[start])
	fnName, params, bodyStr, err := ParseFunctionDeclarationLine(trimmedLine)
	if err != nil {
		return nil, 0, err
	}

	if bodyStr != "" {
		strippedBody := preprocessor.StripSingleLineComment(bodyStr)
		if strings.TrimSpace(strippedBody) != "" && IsDelimiterBalanced(strippedBody) {
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
		if IsDirectiveLine(trimmed) {
			candidate := strings.TrimSpace(strings.Join(bodyLines, "\n"))
			if candidate != "" {
				candidate = preprocessor.StripLineComments(candidate)
				if IsDelimiterBalanced(candidate) {
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

func CollapseWhitespaceOutsideStrings(input string) string {
	collapsed, _ := CollapseWhitespaceOutsideStringsWithMapping(input)
	return collapsed
}

func CollapseWhitespaceOutsideStringsWithMapping(input string) (string, []int) {
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

// IsDelimiterBalanced checks brackets, braces, and parentheses while respecting strings.
func IsDelimiterBalanced(s string) bool {
	var sc stringutils.ScanState
	for i := 0; i < len(s); i++ {
		sc.Advance(s[i])
	}
	return sc.Depth() == 0
}

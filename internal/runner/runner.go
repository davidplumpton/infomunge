package runner

import (
	"fmt"
	declparser "infomunge/internal/declarations"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/output"
	"infomunge/internal/preprocessor"
	"infomunge/internal/runtimeio"
	"infomunge/internal/stringutils"
	"strconv"
	"strings"
)

// RunnerOptions holds configuration for execution
type RunnerOptions struct {
	BaseDir               string
	Lazy                  bool
	FormatService         evaluator.FormatService
	URLReadService        evaluator.URLReadService
	DisableURLReadService bool
}

const lazyFlagUnsupportedMessage = "--lazy is currently unsupported; use lazy_eval/force_eval and __toStream/__lazyMap/__lazyFilter/__lazyReduce builtins directly"

func installEvaluationCapabilities(scope *evaluator.Scope, opts RunnerOptions) *evaluator.Scope {
	if scope == nil {
		scope = evaluator.NewScope(nil)
	}

	formatService, ok := scope.FormatService()
	if opts.FormatService != nil {
		formatService = opts.FormatService
	} else if !ok {
		formatService = runtimeio.FormatService{}
	}
	scope.SetFormatService(formatService)

	if opts.DisableURLReadService {
		scope.SetURLReadService(evaluator.DisabledURLReadService())
		return scope
	}
	if opts.URLReadService != nil {
		scope.SetURLReadService(opts.URLReadService)
		return scope
	}
	if _, ok := scope.URLReadService(); !ok {
		scope.SetURLReadService(runtimeio.NewURLReadService(formatService))
	}
	return scope
}

// RequireScriptHeader validates that an output-oriented script includes a header separator.
func RequireScriptHeader(raw string) error {
	_, _, bodyOffset := preprocessor.ExtractHeaderAndBody(raw)
	if bodyOffset == 0 {
		return unifiederrors.ParseError("script must have a header with '---' separator")
	}
	return nil
}

// FormatExecutionResult resolves and serializes an execution result without printing it.
func FormatExecutionResult(result ExecutionResult) (string, error) {
	resolved, err := result.Resolved()
	if err != nil {
		return "", err
	}
	if !resolved.HasHeader && strings.TrimSpace(resolved.OutputMimeType) == "" {
		return fmt.Sprint(resolved.Value), nil
	}
	return output.FormatResultWithMetadata(resolved.Value, resolved.OutputMimeType, resolved.Context, resolved.OutputMetadata)
}

func applyOutputDeclaration(decl *OutputDeclaration, outputMimeType *string, metadata *output.Metadata) error {
	if decl == nil {
		return unifiederrors.InternalError("internal error: missing output declaration")
	}
	if decl.MimeType == "" {
		*outputMimeType = ""
		return nil
	}

	*outputMimeType = decl.MimeType

	if decl.Options == "" {
		return nil
	}

	parsed, err := parseOutputOptions(decl.Options)
	if err != nil {
		return err
	}
	if len(parsed) > 0 {
		output.SetOptions(metadata, parsed)
	}
	return nil
}

func applyInputDeclaration(decl *InputDeclaration) {
	// Input directives are accepted as DataWeave-compatible metadata. Input
	// bytes are parsed by the adapter layer before the runner receives context.
	_ = decl
}

func splitFirstToken(s string) (string, string) {
	for i, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return s[:i], strings.TrimSpace(s[i:])
		}
	}
	return s, ""
}

func parseOutputOptions(input string) (map[string]string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}
	if strings.HasPrefix(trimmed, "{") {
		if !strings.HasSuffix(trimmed, "}") {
			return nil, unifiederrors.ParseError("output options missing closing brace")
		}
		trimmed = strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}
	if trimmed == "" {
		return nil, nil
	}

	parts := splitOptions(trimmed)
	options := make(map[string]string, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx := strings.IndexAny(part, "=:")
		if idx == -1 {
			return nil, unifiederrors.ParseErrorf("invalid output option %q", part)
		}
		key := strings.TrimSpace(part[:idx])
		value := strings.TrimSpace(part[idx+1:])
		if key == "" {
			return nil, unifiederrors.ParseErrorf("invalid output option %q", part)
		}
		unquoted, quoted, err := unquoteOptionValue(value)
		if err != nil {
			return nil, unifiederrors.ParseErrorf("invalid output option %q: %v", part, err)
		}
		if quoted {
			value = unquoted
		}
		options[key] = value
	}
	return options, nil
}

func splitOptions(input string) []string {
	var parts []string
	var sc stringutils.ScanState
	start := 0

	for i := 0; i < len(input); i++ {
		ch := input[i]
		sc.Advance(ch)
		if !sc.InString() && ch == ',' {
			parts = append(parts, input[start:i])
			start = i + 1
		}
	}
	if start <= len(input) {
		parts = append(parts, input[start:])
	}
	return parts
}

func unquoteOptionValue(value string) (string, bool, error) {
	if value == "" || (value[0] != '"' && value[0] != '\'') {
		return value, false, nil
	}

	quote := value[0]
	if len(value) < 2 || value[len(value)-1] != quote || stringutils.IsEscapedAt(value, len(value)-1) {
		return value, true, fmt.Errorf("unterminated quoted value")
	}

	if quote == '"' {
		unquoted, err := strconv.Unquote(value)
		if err != nil {
			return value, true, fmt.Errorf("invalid double-quoted value: %w", err)
		}
		return unquoted, true, nil
	}

	var unquoted strings.Builder
	unquoted.Grow(len(value) - 2)
	for i := 1; i < len(value)-1; i++ {
		ch := value[i]
		if ch == '\'' {
			return value, true, fmt.Errorf("unescaped quote in single-quoted value")
		}
		if ch != '\\' {
			unquoted.WriteByte(ch)
			continue
		}

		if i+1 >= len(value)-1 {
			return value, true, fmt.Errorf("unterminated escape in single-quoted value")
		}
		next := value[i+1]
		switch next {
		case '\'', '\\':
			unquoted.WriteByte(next)
			i++
		default:
			// InfoMunge single-quoted strings preserve unknown backslash
			// sequences rather than interpreting Go escape syntax.
			unquoted.WriteByte(ch)
		}
	}
	return unquoted.String(), true, nil
}

func applyNamespaceDeclaration(decl *NamespaceDeclaration, namespaces map[string]string) error {
	if decl == nil {
		return unifiederrors.InternalError("internal error: missing namespace declaration")
	}
	namespaces[decl.Prefix] = decl.URI
	return nil
}

var directiveKeywords = declparser.DirectiveKeywords

// normalizeHeader converts a single-line header into multi-line format by inserting
// newlines before directive keywords, being careful to skip keywords inside brackets or strings.
func normalizeHeader(header string) string {
	// If already multi-line, return as-is
	if strings.Contains(header, "\n") {
		return header
	}

	var result strings.Builder
	var sc stringutils.ScanState

	for i := 0; i < len(header); i++ {
		ch := header[i]
		sc.Advance(ch)

		// Check for directive keywords at top level
		if sc.AtTopLevel() {
			for _, kw := range directiveKeywords {
				if i+len(kw) <= len(header) && header[i:i+len(kw)] == kw {
					// Check word boundary before keyword
					if i > 0 && !isWordBoundaryChar(header[i-1]) {
						continue
					}
					// Insert newline before keyword (unless at start)
					if result.Len() > 0 {
						// Trim trailing space from result before adding newline
						s := result.String()
						result.Reset()
						result.WriteString(strings.TrimRight(s, " "))
						result.WriteByte('\n')
					}
					break
				}
			}
		}

		result.WriteByte(ch)
	}

	return result.String()
}

// isWordBoundaryChar returns true if ch is a word boundary character.
func isWordBoundaryChar(ch byte) bool {
	return !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_')
}

// isEscaped checks if the character at position i is escaped by counting
// consecutive backslashes before it. An odd number of backslashes means escaped.
// Example: "hello\"" -> quote at end is escaped (1 backslash)
// Example: "hello\\"  -> quote at end is NOT escaped (2 backslashes = escaped backslash)
func normalizeHeaderLines(header string) []string {
	return strings.Split(normalizeHeader(header), "\n")
}

func ResolveResult(result evaluator.Value) (evaluator.Value, error) {
	switch r := result.(type) {
	case *evaluator.LazyValue:
		resolved, err := r.GetValue()
		if err != nil {
			return nil, err
		}
		return ResolveResult(resolved)
	case *evaluator.StreamWithError:
		var values evaluator.Array
		for val := range r.Stream {
			values = append(values, val)
		}
		if err := r.WaitError(); err != nil {
			return nil, err
		}
		return values, nil
	case chan evaluator.Value:
		var values evaluator.Array
		for val := range r {
			values = append(values, val)
		}
		return values, nil
	default:
		return result, nil
	}
}

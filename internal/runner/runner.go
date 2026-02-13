package runner

import (
	"context"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/preprocessor"
	"infomunge/pkg/formats"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Context keys for eval context metadata shared between runner and handlers.
const (
	ContextKeyNamespaces    = "__namespaces__"
	ContextKeyOutputOptions = "__output_options__"
)

// RunnerOptions holds configuration for execution
type RunnerOptions struct {
	BaseDir string
	Lazy    bool
}

// Run executes the infomunge process on the given file.
func Run(filePath string) error {
	return RunWithConfig(filePath, RunnerOptions{})
}

// RunWithConfig executes the infomunge process on the given file with options.
func RunWithConfig(filePath string, opts RunnerOptions) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return unifiederrors.WrapIO(err, "error reading file")
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	if opts.BaseDir == "" {
		opts.BaseDir = filepath.Dir(absPath)
	}

	return runFromStringWithConfig(context.Background(), string(content), nil, opts)
}

// RunFromString executes an infomunge script from a string and prints formatted output.
func RunFromString(raw string) error {
	return RunFromStringWithContext(raw, nil)
}

// RunFromStringWithContext executes an infomunge script with additional context variables.
func RunFromStringWithContext(raw string, additionalContext map[string]interface{}) error {
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}
	return runFromStringWithConfig(context.Background(), raw, additionalContext, RunnerOptions{BaseDir: baseDir})
}

// RunFromStringWithContextAndOptions executes an infomunge script with additional context variables and options.
func RunFromStringWithContextAndOptions(raw string, additionalContext map[string]interface{}, opts RunnerOptions) error {
	return RunFromStringWithContextAndOptionsAndGoContext(context.Background(), raw, additionalContext, opts)
}

// RunFromStringWithContextAndOptionsAndGoContext executes an infomunge script with additional context variables, options, and Go context.
func RunFromStringWithContextAndOptionsAndGoContext(goCtx context.Context, raw string, additionalContext map[string]interface{}, opts RunnerOptions) error {
	if opts.BaseDir == "" {
		baseDir, err := os.Getwd()
		if err != nil {
			opts.BaseDir = "."
		} else {
			opts.BaseDir = baseDir
		}
	}
	return runFromStringWithConfig(goCtx, raw, additionalContext, opts)
}

// runFromStringWithConfig executes an infomunge script with configuration.
func runFromStringWithConfig(goCtx context.Context, raw string, additionalContext map[string]interface{}, opts RunnerOptions) error {
	_, _, bodyOffset := preprocessor.ExtractHeaderAndBody(raw)
	if bodyOffset == 0 {
		return unifiederrors.ParseError("script must have a header with '---' separator")
	}

	result, hasHeader, outputMimeType, context, err := evaluateWithContext(goCtx, raw, additionalContext, opts)
	if err != nil {
		return err
	}

	result, err = ResolveResult(result)
	if err != nil {
		return err
	}

	return formatOutputWithContext(result, hasHeader, outputMimeType, context)
}

// RunString executes an infomunge script from a string with optional additional context.
// The additionalContext map allows injecting variables (like "payload") into the evaluation context.
func RunString(script string, additionalContext map[string]interface{}) (interface{}, error) {
	return RunStringWithGoContext(context.Background(), script, additionalContext)
}

// RunStringWithGoContext executes an infomunge script from a string with optional additional context and Go context.
func RunStringWithGoContext(goCtx context.Context, script string, additionalContext map[string]interface{}) (interface{}, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}
	result, _, _, err := evaluate(goCtx, script, additionalContext, RunnerOptions{BaseDir: baseDir})
	if err != nil {
		return nil, err
	}
	return ResolveResult(result)
}

// RunStringWithContextAndOptionsWithOutput executes a script and returns result plus output metadata.
func RunStringWithContextAndOptionsWithOutput(script string, additionalContext map[string]interface{}, opts RunnerOptions) (interface{}, bool, string, map[string]interface{}, error) {
	return RunStringWithGoContextAndOptionsWithOutput(context.Background(), script, additionalContext, opts)
}

// RunStringWithGoContextAndOptionsWithOutput executes a script and returns result plus output metadata.
func RunStringWithGoContextAndOptionsWithOutput(goCtx context.Context, script string, additionalContext map[string]interface{}, opts RunnerOptions) (interface{}, bool, string, map[string]interface{}, error) {
	if opts.BaseDir == "" {
		baseDir, err := os.Getwd()
		if err != nil {
			opts.BaseDir = "."
		} else {
			opts.BaseDir = baseDir
		}
	}
	return evaluateWithContext(goCtx, script, additionalContext, opts)
}

// evaluate is the core evaluation logic shared by all runner functions.
func evaluate(goCtx context.Context, raw string, additionalContext map[string]interface{}, opts RunnerOptions) (interface{}, bool, string, error) {
	result, hasHeader, outputMimeType, _, err := evaluateWithContext(goCtx, raw, additionalContext, opts)
	return result, hasHeader, outputMimeType, err
}

// evaluateWithContext is the core evaluation logic that also returns the parsed context.
func evaluateWithContext(goCtx context.Context, raw string, additionalContext map[string]interface{}, opts RunnerOptions) (interface{}, bool, string, map[string]interface{}, error) {
	header, body, bodyOffset := preprocessor.ExtractHeaderAndBody(raw)
	hasHeader := bodyOffset != 0

	loader := NewModuleLoader(opts.BaseDir)
	context, outputMimeType, err := parseHeaderWithGoContext(header, hasHeader, goCtx, raw, loader)
	if err != nil {
		return nil, hasHeader, outputMimeType, nil, err
	}

	// Inject additional context (e.g., payload from -i flags)
	for k, v := range additionalContext {
		context[k] = v
	}

	prepOpts := preprocessor.Options{}
	if strings.ContainsAny(body, "\n\r") {
		prepOpts.AllowMultilineIfElse = true
	}
	parseableExpr, mapping, err := preprocessor.PrepareForParsing(body, prepOpts)
	if err != nil {
		return nil, hasHeader, outputMimeType, nil, err
	}
	result, err := evaluator.EvaluateWithGoContext(parseableExpr, context, goCtx, mapping, bodyOffset, raw)
	if err != nil {
		return nil, hasHeader, outputMimeType, nil, err
	}

	return result, hasHeader, outputMimeType, context, nil
}

// handleOutputDecl processes output directive and captures output options.
func handleOutputDecl(trimmedLine string, outputMimeType *string, context map[string]interface{}) error {
	rest := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "output "))
	if rest == "" {
		*outputMimeType = ""
		return nil
	}

	mimeType, options := splitFirstToken(rest)
	*outputMimeType = mimeType

	if options == "" {
		return nil
	}

	parsed, err := parseOutputOptions(options)
	if err != nil {
		return err
	}
	if len(parsed) > 0 {
		context[ContextKeyOutputOptions] = parsed
	}
	return nil
}

// handleInputDecl processes input directive.
func handleInputDecl(trimmedLine string, context map[string]interface{}) {
	inputMimeType := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "input "))
	if inputMimeType != "" {
		context["__input_mime__"] = inputMimeType
	}
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
		if unquoted, ok := unquoteOptionValue(value); ok {
			value = unquoted
		}
		options[key] = value
	}
	return options, nil
}

func splitOptions(input string) []string {
	var parts []string
	inString := false
	escape := false
	quote := byte(0)
	start := 0

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if escape {
			escape = false
			continue
		}
		if inString {
			if ch == '\\' {
				escape = true
				continue
			}
			if ch == quote {
				inString = false
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			inString = true
			quote = ch
			continue
		}
		if ch == ',' {
			parts = append(parts, input[start:i])
			start = i + 1
		}
	}
	if start <= len(input) {
		parts = append(parts, input[start:])
	}
	return parts
}

func unquoteOptionValue(value string) (string, bool) {
	if len(value) < 2 {
		return value, false
	}
	if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
		unquoted, err := strconv.Unquote(value)
		if err != nil {
			return value, false
		}
		return unquoted, true
	}
	return value, false
}

// handleNamespaceDecl processes namespace declaration
func handleNamespaceDecl(trimmedLine string, namespaces map[string]string) error {
	prefix, uri, err := parseNamespaceDecl(trimmedLine)
	if err != nil {
		return err
	}
	namespaces[prefix] = uri
	return nil
}

// handleVariableDecl processes variable declaration
func handleVariableDecl(line, trimmedLine string, headerOffset int, context map[string]interface{}, goCtx context.Context, fullRaw string) error {
	val, varName, err := parseVarDeclWithGoContext(line, trimmedLine, headerOffset, context, goCtx, fullRaw)
	if err != nil {
		return err
	}
	if varName != "" {
		context[varName] = val
	}
	return nil
}

// handleFunctionDecl processes function declaration
func handleFunctionDecl(trimmedLine string, context map[string]interface{}) error {
	fn, fnName, err := parseFunDecl(trimmedLine, nil)
	if err != nil {
		return err
	}
	if fnName != "" {
		context[fnName] = fn
	}
	return nil
}

// handleTypeDecl processes type declaration
func handleTypeDecl(trimmedLine string, context map[string]interface{}) error {
	typeDef, typeName, err := parseTypeDecl(trimmedLine)
	if err != nil {
		return err
	}
	if typeName != "" {
		context[typeName] = typeDef
	}
	return nil
}

// parseState holds the state during header parsing
type parseState struct {
	context        map[string]interface{}
	namespaces     map[string]string
	outputMimeType *string
	headerOffset   int
	fullRaw        string
	loader         *ModuleLoader
	goCtx          context.Context
}

// directiveRegistration pairs a keyword with its handler function.
// Some directives (fun/var) are parsed by dedicated multiline handlers.
type directiveRegistration struct {
	keyword string
	handler func(line, trimmedLine string, state *parseState) error
}

// directiveRegistrations defines all known header directives in one place.
var directiveRegistrations = []directiveRegistration{
	{"output ", func(_, trimmed string, state *parseState) error {
		return handleOutputDecl(trimmed, state.outputMimeType, state.context)
	}},
	{"input ", func(_, trimmed string, state *parseState) error {
		handleInputDecl(trimmed, state.context)
		return nil
	}},
	{"%dw ", func(_, _ string, _ *parseState) error { return nil }},
	{"%im ", func(_, _ string, _ *parseState) error { return nil }},
	{"ns ", func(_, trimmed string, state *parseState) error {
		return handleNamespaceDecl(trimmed, state.namespaces)
	}},
	{"import ", func(_, trimmed string, state *parseState) error {
		return handleImport(trimmed, state.context, state.loader)
	}},
	{"var ", nil},
	{"fun ", nil},
	{"type ", func(_, trimmed string, state *parseState) error {
		return handleTypeDecl(trimmed, state.context)
	}},
}

// directiveKeywords are the keywords that start header directives.
var directiveKeywords = []string{"%im ", "%dw ", "output ", "input ", "var ", "fun ", "ns ", "import ", "type "}

// normalizeHeader converts a single-line header into multi-line format by inserting
// newlines before directive keywords, being careful to skip keywords inside brackets or strings.
func normalizeHeader(header string) string {
	// If already multi-line, return as-is
	if strings.Contains(header, "\n") {
		return header
	}

	var result strings.Builder
	i := 0
	depth := 0 // bracket depth
	inString := false
	stringChar := byte(0)

	for i < len(header) {
		ch := header[i]

		// Track string state
		if !inString && (ch == '"' || ch == '\'') {
			inString = true
			stringChar = ch
		} else if inString && ch == stringChar && !isEscaped(header, i) {
			inString = false
		}

		// Track bracket depth
		if !inString {
			if ch == '(' || ch == '[' || ch == '{' {
				depth++
			} else if ch == ')' || ch == ']' || ch == '}' {
				depth--
			}
		}

		// Check for directive keywords at depth 0, outside strings
		if depth == 0 && !inString {
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
		i++
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
func isEscaped(s string, i int) bool {
	backslashCount := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		backslashCount++
	}
	return backslashCount%2 == 1
}

func normalizeHeaderLines(header string) []string {
	return strings.Split(normalizeHeader(header), "\n")
}

func parseDirectiveLine(lines []string, index int, headerOffset int, state *parseState) (int, error) {
	line := lines[index]
	trimmedLine := strings.TrimSpace(line)
	if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
		return 1, nil
	}

	state.headerOffset = headerOffset
	if strings.HasPrefix(trimmedLine, "fun ") {
		fn, fnName, consumed, err := parseFunDeclFromLines(lines, index, state.context)
		if err != nil {
			return 0, withHeaderLineContext(err, state, headerOffset, line)
		}
		if fnName != "" {
			state.context[fnName] = fn
		}
		return consumed, nil
	}
	if strings.HasPrefix(trimmedLine, "var ") {
		val, varName, consumed, err := parseVarDeclFromLinesWithGoContext(lines, index, state.headerOffset, state.context, state.goCtx, state.fullRaw)
		if err != nil {
			return 0, withHeaderLineContext(err, state, headerOffset, line)
		}
		if varName != "" {
			state.context[varName] = val
		}
		return consumed, nil
	}

	for _, directive := range directiveRegistrations {
		if strings.HasPrefix(trimmedLine, directive.keyword) {
			if directive.handler == nil {
				return 0, withHeaderLineContext(unifiederrors.ParseErrorf("unrecognized header directive: %s", trimmedLine), state, headerOffset, line)
			}
			if err := directive.handler(line, trimmedLine, state); err != nil {
				return 0, withHeaderLineContext(err, state, headerOffset, line)
			}
			return 1, nil
		}
	}

	return 0, withHeaderLineContext(unifiederrors.ParseErrorf("unrecognized header directive: %s", trimmedLine), state, headerOffset, line)
}

func parseHeaderLines(lines []string, state *parseState) error {
	headerOffset := 0
	for i := 0; i < len(lines); {
		consumed, err := parseDirectiveLine(lines, i, headerOffset, state)
		if err != nil {
			return err
		}
		for j := 0; j < consumed; j++ {
			headerOffset += len(lines[i+j]) + 1
		}
		i += consumed
	}
	return nil
}

func withHeaderLineContext(err error, state *parseState, headerOffset int, line string) error {
	lineOffset := headerOffset + leadingWhitespaceOffset(line)
	return attachLineContext(err, state.fullRaw, lineOffset, line)
}

func parseHeader(header string, hasHeader bool, fullRaw string, loader *ModuleLoader) (map[string]interface{}, string, error) {
	return parseHeaderWithGoContext(header, hasHeader, context.Background(), fullRaw, loader)
}

func parseHeaderWithGoContext(header string, hasHeader bool, goCtx context.Context, fullRaw string, loader *ModuleLoader) (map[string]interface{}, string, error) {
	context := make(map[string]interface{})
	outputMimeType := "application/json"

	if !hasHeader {
		return context, outputMimeType, nil
	}

	state := &parseState{
		context:        context,
		namespaces:     make(map[string]string),
		outputMimeType: &outputMimeType,
		fullRaw:        fullRaw,
		loader:         loader,
		goCtx:          goCtx,
	}

	lines := normalizeHeaderLines(header)
	if err := parseHeaderLines(lines, state); err != nil {
		return nil, "", err
	}

	// Store namespaces in context if any were declared
	if len(state.namespaces) > 0 {
		state.context[ContextKeyNamespaces] = state.namespaces
	}

	return state.context, *state.outputMimeType, nil
}

func formatOutput(result interface{}, hasHeader bool, mimeType string) error {
	return formatOutputWithContext(result, hasHeader, mimeType, nil)
}

func ResolveResult(result interface{}) (interface{}, error) {
	switch r := result.(type) {
	case *evaluator.LazyValue:
		resolved, err := r.GetValue()
		if err != nil {
			return nil, err
		}
		return ResolveResult(resolved)
	case *evaluator.StreamWithError:
		var values []interface{}
		for val := range r.Stream {
			values = append(values, val)
		}
		if err := r.WaitError(); err != nil {
			return nil, err
		}
		return values, nil
	case chan evaluator.Value:
		var values []interface{}
		for val := range r {
			values = append(values, val)
		}
		return values, nil
	default:
		return result, nil
	}
}

// formatOutputWithContext formats output with optional context (for namespaces, etc).
func formatOutputWithContext(result interface{}, hasHeader bool, mimeType string, context map[string]interface{}) error {
	// Check if result is a stream (lazy evaluation result)
	switch r := result.(type) {
	case *evaluator.StreamWithError:
		var values []interface{}
		for val := range r.Stream {
			values = append(values, val)
		}
		if err := r.WaitError(); err != nil {
			return err
		}
		result = values
	case chan evaluator.Value:
		var values []interface{}
		for val := range r {
			values = append(values, val)
		}
		result = values
	}

	if !hasHeader {
		fmt.Print(result)
		return nil
	}

	var output string
	var err error

	// If XML output and namespaces are declared in context, use them
	if mimeType == "application/xml" {
		var nsMap map[string]string
		if context != nil {
			if declared, ok := context[ContextKeyNamespaces].(map[string]string); ok {
				nsMap = declared
			}
		}
		xmlOpts := formats.XMLOutputOptions{
			DeclaredNamespaces: nsMap,
			NamespaceVars:      extractNamespaceVars(context),
			WriteDeclaration:   true,
		}
		if context != nil {
			if rawOpts, ok := context[ContextKeyOutputOptions].(map[string]string); ok {
				if err := applyXMLOutputOptions(&xmlOpts, rawOpts); err != nil {
					return err
				}
			}
		}
		output, err = formats.FormatXMLWithOptions(result, xmlOpts)
	} else {
		output, err = formats.Format(result, mimeType)
	}

	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}

func extractNamespaceVars(context map[string]interface{}) map[string]formats.Namespace {
	if context == nil {
		return nil
	}
	vars := make(map[string]formats.Namespace)
	for k, v := range context {
		if ns, ok := v.(formats.Namespace); ok {
			vars[k] = ns
		}
	}
	if len(vars) == 0 {
		return nil
	}
	return vars
}

func applyXMLOutputOptions(opts *formats.XMLOutputOptions, raw map[string]string) error {
	for key, value := range raw {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "writedeclaration":
			parsed, err := parseBoolOption(value)
			if err != nil {
				return unifiederrors.ParseErrorf("output option writeDeclaration: %v", err)
			}
			opts.WriteDeclaration = parsed
		case "writedeclarednamespaces":
			opts.WriteDeclaredNamespaces = value
		case "writenilonnull":
			parsed, err := parseBoolOption(value)
			if err != nil {
				return unifiederrors.ParseErrorf("output option writeNilOnNull: %v", err)
			}
			opts.WriteNilOnNull = parsed
		case "skipnullon":
			opts.SkipNullOn = value
		}
	}
	return nil
}

func parseBoolOption(value string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false, fmt.Errorf("empty value")
	}
	return strconv.ParseBool(trimmed)
}

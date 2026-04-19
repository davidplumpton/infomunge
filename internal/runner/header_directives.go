package runner

import (
	"context"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
)

type headerDirectiveKind string

const (
	headerDirectiveVersion   headerDirectiveKind = "version"
	headerDirectiveOutput    headerDirectiveKind = "output"
	headerDirectiveInput     headerDirectiveKind = "input"
	headerDirectiveNamespace headerDirectiveKind = "namespace"
	headerDirectiveImport    headerDirectiveKind = "import"
	headerDirectiveVar       headerDirectiveKind = "var"
	headerDirectiveFun       headerDirectiveKind = "fun"
	headerDirectiveType      headerDirectiveKind = "type"
)

type headerDirective struct {
	kind    headerDirectiveKind
	line    string
	lines   []string
	trimmed string
	offset  int
}

func parseHeader(header string, hasHeader bool, fullRaw string, loader *ModuleLoader) (evaluator.Context, string, error) {
	return parseHeaderWithGoContext(header, hasHeader, context.Background(), fullRaw, loader)
}

func parseHeaderWithGoContext(header string, hasHeader bool, goCtx context.Context, fullRaw string, loader *ModuleLoader) (evaluator.Context, string, error) {
	directives, err := parseHeaderDirectives(header, hasHeader, fullRaw)
	if err != nil {
		return nil, "", err
	}
	return applyHeaderDirectives(directives, goCtx, fullRaw, loader)
}

func parseHeaderDirectives(header string, hasHeader bool, fullRaw string) ([]headerDirective, error) {
	if !hasHeader {
		return nil, nil
	}

	lines := normalizeHeaderLines(header)
	directives := make([]headerDirective, 0, len(lines))
	headerOffset := 0

	for i := 0; i < len(lines); {
		directive, consumed, err := parseHeaderDirective(lines, i, headerOffset, fullRaw)
		if err != nil {
			return nil, err
		}
		if directive != nil {
			directives = append(directives, *directive)
		}
		for j := 0; j < consumed; j++ {
			headerOffset += len(lines[i+j]) + 1
		}
		i += consumed
	}

	return directives, nil
}

func parseHeaderDirective(lines []string, index int, headerOffset int, fullRaw string) (*headerDirective, int, error) {
	line := lines[index]
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "//") {
		return nil, 1, nil
	}

	newDirective := func(kind headerDirectiveKind, consumed int) *headerDirective {
		return &headerDirective{
			kind:    kind,
			line:    line,
			lines:   append([]string(nil), lines[index:index+consumed]...),
			trimmed: trimmed,
			offset:  headerOffset,
		}
	}

	switch {
	case strings.HasPrefix(trimmed, "output "):
		if err := validateOutputDirective(trimmed); err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDirective(headerDirectiveOutput, 1), 1, nil
	case strings.HasPrefix(trimmed, "input "):
		return newDirective(headerDirectiveInput, 1), 1, nil
	case strings.HasPrefix(trimmed, "%dw "), strings.HasPrefix(trimmed, "%im "):
		return newDirective(headerDirectiveVersion, 1), 1, nil
	case strings.HasPrefix(trimmed, "ns "):
		if _, _, err := parseNamespaceDecl(trimmed); err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDirective(headerDirectiveNamespace, 1), 1, nil
	case strings.HasPrefix(trimmed, "import "):
		return newDirective(headerDirectiveImport, 1), 1, nil
	case strings.HasPrefix(trimmed, "var "):
		spec, err := parseVarDeclSource(lines, index)
		if err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDirective(headerDirectiveVar, spec.consumed), spec.consumed, nil
	case strings.HasPrefix(trimmed, "fun "):
		spec, err := parseFunDeclSource(lines, index)
		if err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDirective(headerDirectiveFun, spec.consumed), spec.consumed, nil
	case strings.HasPrefix(trimmed, "type "):
		if _, _, err := parseTypeDecl(trimmed); err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDirective(headerDirectiveType, 1), 1, nil
	default:
		return nil, 0, withHeaderLineContext(unifiederrors.ParseErrorf("unrecognized header directive: %s", trimmed), fullRaw, headerOffset, line)
	}
}

func validateOutputDirective(trimmedLine string) error {
	rest := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "output "))
	if rest == "" {
		return nil
	}

	_, options := splitFirstToken(rest)
	if options == "" {
		return nil
	}

	_, err := parseOutputOptions(options)
	return err
}

func applyHeaderDirectives(directives []headerDirective, goCtx context.Context, fullRaw string, loader *ModuleLoader) (evaluator.Context, string, error) {
	context := make(evaluator.Context)
	namespaces := make(map[string]string)
	outputMimeType := "application/json"

	for _, directive := range directives {
		if err := applyHeaderDirective(directive, context, namespaces, &outputMimeType, goCtx, fullRaw, loader); err != nil {
			return nil, "", err
		}
	}

	if len(namespaces) > 0 {
		context[ContextKeyNamespaces] = namespaces
	}

	return context, outputMimeType, nil
}

func applyHeaderDirective(directive headerDirective, context evaluator.Context, namespaces map[string]string, outputMimeType *string, goCtx context.Context, fullRaw string, loader *ModuleLoader) error {
	var err error

	switch directive.kind {
	case headerDirectiveVersion:
		return nil
	case headerDirectiveOutput:
		err = handleOutputDecl(directive.trimmed, outputMimeType, context)
	case headerDirectiveInput:
		handleInputDecl(directive.trimmed, context)
	case headerDirectiveNamespace:
		err = handleNamespaceDecl(directive.trimmed, namespaces)
	case headerDirectiveImport:
		err = handleImport(directive.trimmed, context, loader)
	case headerDirectiveVar:
		var val evaluator.Value
		var varName string
		var consumed int
		val, varName, consumed, err = parseVarDeclFromLinesWithGoContext(directive.lines, 0, directive.offset, context, goCtx, fullRaw)
		if err == nil && consumed > 0 && varName != "" {
			context[varName] = val
		}
	case headerDirectiveFun:
		var fnName string
		var fn evaluator.Value
		var consumed int
		fn, fnName, consumed, err = parseFunDeclFromLinesWithSource(directive.lines, 0, context, fullRaw, directive.offset)
		if err == nil && consumed > 0 && fnName != "" {
			context[fnName] = fn
		}
	case headerDirectiveType:
		err = handleTypeDecl(directive.trimmed, context)
	}

	if err != nil {
		return withHeaderLineContext(err, fullRaw, directive.offset, directive.line)
	}
	return nil
}

func withHeaderLineContext(err error, fullRaw string, headerOffset int, line string) error {
	lineOffset := headerOffset + leadingWhitespaceOffset(line)
	return attachLineContext(err, fullRaw, lineOffset, line)
}

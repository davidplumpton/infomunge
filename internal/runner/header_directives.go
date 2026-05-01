package runner

import (
	"context"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/output"
)

type directiveParsePolicy struct {
	allowedKinds       map[headerDirectiveKind]bool
	bodySeparatorError string
	disallowContext    string
}

var scriptHeaderDirectivePolicy = directiveParsePolicy{
	allowedKinds: map[headerDirectiveKind]bool{
		headerDirectiveVersion:   true,
		headerDirectiveOutput:    true,
		headerDirectiveInput:     true,
		headerDirectiveNamespace: true,
		headerDirectiveImport:    true,
		headerDirectiveVar:       true,
		headerDirectiveFun:       true,
		headerDirectiveType:      true,
	},
	disallowContext: "script headers",
}

var moduleDirectivePolicy = directiveParsePolicy{
	allowedKinds: map[headerDirectiveKind]bool{
		headerDirectiveVersion: true,
		headerDirectiveImport:  true,
		headerDirectiveVar:     true,
		headerDirectiveFun:     true,
		headerDirectiveType:    true,
	},
	bodySeparatorError: "modules cannot contain a body section (---)",
	disallowContext:    "modules",
}

func parseHeader(header string, hasHeader bool, fullRaw string, loader *ModuleLoader) (*evaluator.Scope, string, output.Metadata, error) {
	return parseHeaderWithGoContext(header, hasHeader, context.Background(), fullRaw, loader)
}

func parseHeaderWithGoContext(header string, hasHeader bool, goCtx context.Context, fullRaw string, loader *ModuleLoader) (*evaluator.Scope, string, output.Metadata, error) {
	return parseHeaderWithGoContextAndOptions(header, hasHeader, goCtx, fullRaw, loader, RunnerOptions{})
}

func parseHeaderWithGoContextAndOptions(header string, hasHeader bool, goCtx context.Context, fullRaw string, loader *ModuleLoader, opts RunnerOptions) (*evaluator.Scope, string, output.Metadata, error) {
	declarations, err := parseHeaderDirectives(header, hasHeader, fullRaw)
	if err != nil {
		return nil, "", output.Metadata{}, err
	}
	return applyHeaderDirectivesWithOptions(declarations, goCtx, fullRaw, loader, opts)
}

func parseHeaderDirectives(header string, hasHeader bool, fullRaw string) ([]Declaration, error) {
	if !hasHeader {
		return nil, nil
	}
	return parseDirectives(header, fullRaw, scriptHeaderDirectivePolicy)
}

func parseModuleDirectives(content string) ([]Declaration, error) {
	return parseDirectives(content, content, moduleDirectivePolicy)
}

func parseDirectives(source string, fullRaw string, policy directiveParsePolicy) ([]Declaration, error) {
	lines := normalizeHeaderLines(source)
	declarations := make([]Declaration, 0, len(lines))
	headerOffset := 0

	for i := 0; i < len(lines); {
		declaration, consumed, err := parseDirective(lines, i, headerOffset, fullRaw, policy)
		if err != nil {
			return nil, err
		}
		if declaration != nil {
			declarations = append(declarations, *declaration)
		}
		for j := 0; j < consumed; j++ {
			headerOffset += len(lines[i+j]) + 1
		}
		i += consumed
	}

	return declarations, nil
}

func parseHeaderDirective(lines []string, index int, headerOffset int, fullRaw string) (*Declaration, int, error) {
	return parseDirective(lines, index, headerOffset, fullRaw, scriptHeaderDirectivePolicy)
}

func parseDirective(lines []string, index int, headerOffset int, fullRaw string, policy directiveParsePolicy) (*Declaration, int, error) {
	line := lines[index]
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "//") {
		return nil, 1, nil
	}

	if trimmed == "---" && policy.bodySeparatorError != "" {
		return nil, 0, withHeaderLineContext(unifiederrors.ParseError(policy.bodySeparatorError), fullRaw, headerOffset, line)
	}

	newDeclaration := func(kind headerDirectiveKind, consumed int, fill func(*Declaration)) (*Declaration, int, error) {
		if !policy.allowedKinds[kind] {
			err := unifiederrors.ParseErrorf("%s directive is not allowed in %s", kind, policy.disallowContext)
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		declaration := &Declaration{
			Kind:   kind,
			Source: newDeclarationSource(lines, index, consumed, headerOffset),
		}
		if fill != nil {
			fill(declaration)
		}
		return declaration, consumed, nil
	}

	switch {
	case strings.HasPrefix(trimmed, "output "):
		if err := validateOutputDirective(trimmed); err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		outputDecl := parseOutputDeclaration(trimmed)
		return newDeclaration(headerDirectiveOutput, 1, func(declaration *Declaration) {
			declaration.Output = &outputDecl
		})
	case strings.HasPrefix(trimmed, "input "):
		inputDecl := InputDeclaration{Text: strings.TrimSpace(strings.TrimPrefix(trimmed, "input "))}
		return newDeclaration(headerDirectiveInput, 1, func(declaration *Declaration) {
			declaration.Input = &inputDecl
		})
	case strings.HasPrefix(trimmed, "%dw "), strings.HasPrefix(trimmed, "%im "):
		versionDecl := VersionDeclaration{Text: trimmed}
		return newDeclaration(headerDirectiveVersion, 1, func(declaration *Declaration) {
			declaration.Version = &versionDecl
		})
	case strings.HasPrefix(trimmed, "ns "):
		namespaceDecl, err := parseNamespaceDeclaration(trimmed)
		if err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDeclaration(headerDirectiveNamespace, 1, func(declaration *Declaration) {
			declaration.Namespace = namespaceDecl
		})
	case strings.HasPrefix(trimmed, "import "):
		importDecl := parseImportDeclaration(trimmed)
		return newDeclaration(headerDirectiveImport, 1, func(declaration *Declaration) {
			declaration.Import = &importDecl
		})
	case strings.HasPrefix(trimmed, "var "):
		varDecl, consumed, err := parseVarDeclarationFromLines(lines, index)
		if err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDeclaration(headerDirectiveVar, consumed, func(declaration *Declaration) {
			declaration.Var = varDecl
		})
	case strings.HasPrefix(trimmed, "fun "):
		functionDecl, consumed, err := parseFunctionDeclarationFromLines(lines, index)
		if err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDeclaration(headerDirectiveFun, consumed, func(declaration *Declaration) {
			declaration.Function = functionDecl
		})
	case strings.HasPrefix(trimmed, "type "):
		typeDecl, err := parseTypeDeclaration(trimmed)
		if err != nil {
			return nil, 0, withHeaderLineContext(err, fullRaw, headerOffset, line)
		}
		return newDeclaration(headerDirectiveType, 1, func(declaration *Declaration) {
			declaration.Type = typeDecl
		})
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

func applyHeaderDirectives(declarations []Declaration, goCtx context.Context, fullRaw string, loader *ModuleLoader) (*evaluator.Scope, string, output.Metadata, error) {
	return applyHeaderDirectivesWithOptions(declarations, goCtx, fullRaw, loader, RunnerOptions{})
}

func applyHeaderDirectivesWithOptions(declarations []Declaration, goCtx context.Context, fullRaw string, loader *ModuleLoader, opts RunnerOptions) (*evaluator.Scope, string, output.Metadata, error) {
	scope := installEvaluationCapabilities(evaluator.NewScope(nil).WithGoContext(goCtx), opts)
	namespaces := make(map[string]string)
	outputMetadata := output.Metadata{}
	outputMimeType := "application/json"

	for _, declaration := range declarations {
		if err := applyHeaderDirective(declaration, scope, namespaces, &outputMimeType, &outputMetadata, fullRaw, loader); err != nil {
			return nil, "", output.Metadata{}, err
		}
	}

	if len(namespaces) > 0 {
		output.SetDeclaredNamespaces(&outputMetadata, namespaces)
	}

	return scope, outputMimeType, outputMetadata, nil
}

func applyHeaderDirective(declaration Declaration, scope *evaluator.Scope, namespaces map[string]string, outputMimeType *string, outputMetadata *output.Metadata, fullRaw string, loader *ModuleLoader) error {
	var err error

	switch declaration.Kind {
	case headerDirectiveVersion:
		return nil
	case headerDirectiveOutput:
		err = applyOutputDeclaration(declaration.Output, outputMimeType, outputMetadata)
	case headerDirectiveInput:
		applyInputDeclaration(declaration.Input)
	case headerDirectiveNamespace:
		err = applyNamespaceDeclaration(declaration.Namespace, namespaces)
	case headerDirectiveImport:
		err = applyImportDeclaration(declaration.Import, scope.Vars, loader)
	case headerDirectiveVar:
		var val evaluator.Value
		val, err = evaluateVarDeclaration(declaration.Var, declaration.Source, scope, fullRaw)
		if err == nil && declaration.Var != nil && declaration.Var.Name != "" {
			scope.Vars[declaration.Var.Name] = val
		}
	case headerDirectiveFun:
		var fn evaluator.Value
		fn, err = buildFunctionDeclaration(declaration.Function, declaration.Source, scope.Vars, fullRaw)
		if err == nil && declaration.Function != nil && declaration.Function.Name != "" {
			scope.Vars[declaration.Function.Name] = fn
		}
	case headerDirectiveType:
		err = applyTypeDeclaration(declaration.Type, scope.Vars)
	}

	if err != nil {
		return withHeaderLineContext(err, fullRaw, declaration.Source.Offset, declaration.Source.Line)
	}
	return nil
}

func withHeaderLineContext(err error, fullRaw string, headerOffset int, line string) error {
	lineOffset := headerOffset + leadingWhitespaceOffset(line)
	return attachLineContext(err, fullRaw, lineOffset, line)
}

package runner

import (
	"strings"

	"infomunge/internal/evaluator"
)

// parseModuleContent parses a module file (header-only, no ---) and returns its namespace.
func parseModuleContent(content string, loader *ModuleLoader) (Namespace, error) {
	opts := RunnerOptions{}
	if loader != nil {
		opts = loader.Options
	}
	return parseModuleContentWithOptions(content, loader, opts)
}

func parseModuleContentWithOptions(content string, loader *ModuleLoader, opts RunnerOptions) (Namespace, error) {
	content = strings.TrimSpace(content)

	ns := make(Namespace)
	// For evaluation within the module, we need a raw map
	rawNs := installEvaluationCapabilities(make(evaluator.Context), opts)
	directives, err := parseModuleDirectives(content)
	if err != nil {
		return nil, err
	}

	for _, directive := range directives {
		switch directive.kind {
		case headerDirectiveVersion:
			// Version declarations are accepted for DataWeave compatibility but do not affect modules.

		case headerDirectiveImport:
			if err := handleImport(directive.trimmed, rawNs, loader); err != nil {
				return nil, withHeaderLineContext(err, content, directive.offset, directive.line)
			}
			// Copy imported entries to typed namespace
			for k, v := range rawNs {
				if _, exists := ns[k]; !exists {
					ns[k] = inferNamespaceEntry(v)
				}
			}

		case headerDirectiveVar:
			val, varName, consumed, err := parseVarDeclFromLines(directive.lines, 0, directive.offset, rawNs, content)
			if err != nil {
				return nil, withHeaderLineContext(err, content, directive.offset, directive.line)
			}
			if consumed > 0 && varName != "" {
				ns[varName] = NewVarEntry(val)
				rawNs[varName] = val
			}

		case headerDirectiveFun:
			fn, fnName, consumed, err := parseFunDeclFromLinesWithSource(directive.lines, 0, rawNs, content, directive.offset)
			if err != nil {
				return nil, withHeaderLineContext(err, content, directive.offset, directive.line)
			}
			if consumed > 0 && fnName != "" {
				ns[fnName] = NewFuncEntry(fn)
				rawNs[fnName] = fn
			}

		case headerDirectiveType:
			if td, typeName, err := parseTypeDecl(directive.trimmed); err != nil {
				return nil, withHeaderLineContext(err, content, directive.offset, directive.line)
			} else if typeName != "" {
				ns[typeName] = NewTypeDefEntry(td)
				rawNs[typeName] = td
			}
		}
	}

	return ns, nil
}

// inferNamespaceEntry creates a NamespaceEntry by inspecting the value's type.
func inferNamespaceEntry(v evaluator.Value) NamespaceEntry {
	switch val := v.(type) {
	case *evaluator.Lambda:
		return NewFuncEntry(val)
	case *evaluator.TypeDef:
		return NewTypeDefEntry(val)
	default:
		return NewVarEntry(v)
	}
}

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
	// For evaluation within the module, we need a raw map.
	scope := installEvaluationCapabilities(evaluator.NewScope(nil), opts)
	rawNs := scope.Vars
	declarations, err := parseModuleDirectives(content)
	if err != nil {
		return nil, err
	}

	for _, declaration := range declarations {
		switch declaration.Kind {
		case headerDirectiveVersion:
			// Version declarations are accepted for DataWeave compatibility but do not affect modules.

		case headerDirectiveImport:
			if err := applyImportDeclaration(declaration.Import, rawNs, loader); err != nil {
				return nil, withHeaderLineContext(err, content, declaration.Source.Offset, declaration.Source.Line)
			}
			// Copy imported entries to typed namespace
			for k, v := range rawNs {
				if _, exists := ns[k]; !exists {
					ns[k] = inferNamespaceEntry(v)
				}
			}

		case headerDirectiveVar:
			val, err := evaluateVarDeclaration(declaration.Var, declaration.Source, scope, content)
			if err != nil {
				return nil, withHeaderLineContext(err, content, declaration.Source.Offset, declaration.Source.Line)
			}
			if declaration.Var != nil && declaration.Var.Name != "" {
				ns[declaration.Var.Name] = NewVarEntry(val)
				rawNs[declaration.Var.Name] = val
			}

		case headerDirectiveFun:
			fn, err := buildFunctionDeclaration(declaration.Function, declaration.Source, rawNs, content)
			if err != nil {
				return nil, withHeaderLineContext(err, content, declaration.Source.Offset, declaration.Source.Line)
			}
			if declaration.Function != nil && declaration.Function.Name != "" {
				ns[declaration.Function.Name] = NewFuncEntry(fn)
				rawNs[declaration.Function.Name] = fn
			}

		case headerDirectiveType:
			if td, err := bindTypeDeclaration(declaration.Type); err != nil {
				return nil, withHeaderLineContext(err, content, declaration.Source.Offset, declaration.Source.Line)
			} else if declaration.Type != nil && declaration.Type.Name != "" {
				ns[declaration.Type.Name] = NewTypeDefEntry(td)
				rawNs[declaration.Type.Name] = td
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

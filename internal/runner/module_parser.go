package runner

import (
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
)

// parseModuleContent parses a module file (header-only, no ---) and returns its namespace.
func parseModuleContent(content string, loader *ModuleLoader) (Namespace, error) {
	content = strings.TrimSpace(content)

	ns := make(Namespace)
	// For evaluation within the module, we need a raw map
	rawNs := make(evaluator.Context)
	lines := strings.Split(content, "\n")
	offset := 0

	for i := 0; i < len(lines); {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
			offset += len(line) + 1
			i++
			continue
		}

		if trimmedLine == "---" {
			lineOffset := offset + leadingWhitespaceOffset(line)
			return nil, attachLineContext(unifiederrors.ParseError("modules cannot contain a body section (---)"), content, lineOffset, line)
		}

		switch {
		case strings.HasPrefix(trimmedLine, "%dw "), strings.HasPrefix(trimmedLine, "%im "):
			// Version declarations - ignore in modules
			offset += len(line) + 1
			i++
			continue

		case strings.HasPrefix(trimmedLine, "import "):
			if err := handleImport(trimmedLine, rawNs, loader); err != nil {
				lineOffset := offset + leadingWhitespaceOffset(line)
				return nil, attachLineContext(err, content, lineOffset, line)
			}
			// Copy imported entries to typed namespace
			for k, v := range rawNs {
				if _, exists := ns[k]; !exists {
					ns[k] = inferNamespaceEntry(v)
				}
			}

		case strings.HasPrefix(trimmedLine, "var "):
			val, varName, consumed, err := parseVarDeclFromLines(lines, i, offset, rawNs, content)
			if err != nil {
				lineOffset := offset + leadingWhitespaceOffset(line)
				return nil, attachLineContext(err, content, lineOffset, line)
			}
			if varName != "" {
				ns[varName] = NewVarEntry(val)
				rawNs[varName] = val
			}
			for j := 0; j < consumed; j++ {
				offset += len(lines[i+j]) + 1
			}
			i += consumed
			continue

		case strings.HasPrefix(trimmedLine, "fun "):
			fn, fnName, consumed, err := parseFunDeclFromLinesWithSource(lines, i, rawNs, content, offset)
			if err != nil {
				lineOffset := offset + leadingWhitespaceOffset(line)
				return nil, attachLineContext(err, content, lineOffset, line)
			}
			if fnName != "" {
				ns[fnName] = NewFuncEntry(fn)
				rawNs[fnName] = fn
			}
			for j := 0; j < consumed; j++ {
				offset += len(lines[i+j]) + 1
			}
			i += consumed
			continue

		case strings.HasPrefix(trimmedLine, "type "):
			if td, typeName, err := parseTypeDecl(trimmedLine); err != nil {
				lineOffset := offset + leadingWhitespaceOffset(line)
				return nil, attachLineContext(err, content, lineOffset, line)
			} else if typeName != "" {
				ns[typeName] = NewTypeDefEntry(td)
				rawNs[typeName] = td
			}
		}

		offset += len(line) + 1
		i++
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

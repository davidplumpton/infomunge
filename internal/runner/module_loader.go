//go:build !js
// +build !js

package runner

import (
	"os"
	"path/filepath"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
)

// NamespaceEntryKind represents the type of a namespace entry.
type NamespaceEntryKind int

const (
	EntryVar     NamespaceEntryKind = iota // Variable (any value)
	EntryFunc                              // Function (*evaluator.Lambda)
	EntryTypeDef                           // Type definition (*evaluator.TypeDef)
)

// String returns the string representation of a NamespaceEntryKind.
func (k NamespaceEntryKind) String() string {
	switch k {
	case EntryVar:
		return "Var"
	case EntryFunc:
		return "Func"
	case EntryTypeDef:
		return "TypeDef"
	default:
		return "Unknown"
	}
}

// NamespaceEntry represents a typed entry in a module namespace.
// It provides type-safe access to exported vars, functions, and type definitions.
type NamespaceEntry struct {
	Kind  NamespaceEntryKind // Type of entry
	Value interface{}        // The actual value (typed based on Kind)
}

// NewVarEntry creates a NamespaceEntry for a variable.
func NewVarEntry(value interface{}) NamespaceEntry {
	return NamespaceEntry{Kind: EntryVar, Value: value}
}

// NewFuncEntry creates a NamespaceEntry for a function.
func NewFuncEntry(fn *evaluator.Lambda) NamespaceEntry {
	return NamespaceEntry{Kind: EntryFunc, Value: fn}
}

// NewTypeDefEntry creates a NamespaceEntry for a type definition.
func NewTypeDefEntry(td *evaluator.TypeDef) NamespaceEntry {
	return NamespaceEntry{Kind: EntryTypeDef, Value: td}
}

// AsFunc returns the entry as a Lambda if it's a function entry.
func (e NamespaceEntry) AsFunc() (*evaluator.Lambda, bool) {
	if e.Kind != EntryFunc {
		return nil, false
	}
	fn, ok := e.Value.(*evaluator.Lambda)
	return fn, ok
}

// AsTypeDef returns the entry as a TypeDef if it's a type definition entry.
func (e NamespaceEntry) AsTypeDef() (*evaluator.TypeDef, bool) {
	if e.Kind != EntryTypeDef {
		return nil, false
	}
	td, ok := e.Value.(*evaluator.TypeDef)
	return td, ok
}

// Namespace is a typed map of namespace entries.
type Namespace map[string]NamespaceEntry

// ToContext converts the typed namespace to a raw map for use in evaluation context.
func (ns Namespace) ToContext() map[string]interface{} {
	result := make(map[string]interface{}, len(ns))
	for k, entry := range ns {
		result[k] = entry.Value
	}
	return result
}

// Module represents an imported module with its namespace.
type Module struct {
	Name      string    // Simple module name (e.g., "MyModule")
	Namespace Namespace // Exported vars, functions, and types
}

// ModuleLoader handles loading and caching of modules.
type ModuleLoader struct {
	BaseDir     string             // Base directory for module resolution
	SearchPaths []string           // Directories to search for modules
	cache       map[string]*Module // path -> loaded module
	loading     map[string]bool    // Currently loading (for cycle detection)
}

// NewModuleLoader creates a new ModuleLoader with the given base directory.
func NewModuleLoader(baseDir string) *ModuleLoader {
	searchPaths := []string{baseDir}
	modulesDir := filepath.Join(baseDir, "modules")
	if _, err := os.Stat(modulesDir); err == nil {
		searchPaths = append(searchPaths, modulesDir)
	}

	return &ModuleLoader{
		BaseDir:     baseDir,
		SearchPaths: searchPaths,
		cache:       make(map[string]*Module),
		loading:     make(map[string]bool),
	}
}

// Resolve converts a module spec (e.g., "modules::MyModule") into a module name and file path.
func (l *ModuleLoader) Resolve(moduleSpec string) (moduleName string, path string, err error) {
	parts := strings.Split(moduleSpec, "::")
	moduleName = parts[len(parts)-1]

	relPath := filepath.Join(parts...) + ".im"

	for _, searchPath := range l.SearchPaths {
		candidate := filepath.Join(searchPath, relPath)
		if _, err := os.Stat(candidate); err == nil {
			return moduleName, candidate, nil
		}
	}

	return "", "", unifiederrors.IOErrorf("module %q not found (searched in %v)", moduleSpec, l.SearchPaths)
}

// Load loads a module by its spec and returns the Module struct.
func (l *ModuleLoader) Load(moduleSpec string) (*Module, error) {
	moduleName, path, err := l.Resolve(moduleSpec)
	if err != nil {
		return nil, err
	}

	if m, ok := l.cache[path]; ok {
		return m, nil
	}

	if l.loading[path] {
		return nil, unifiederrors.ParseErrorf("cyclic import detected: %s", moduleSpec)
	}
	l.loading[path] = true
	defer delete(l.loading, path)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, unifiederrors.WrapIOf(err, "failed to read module %s", moduleSpec)
	}

	ns, err := parseModuleContent(string(content), l)
	if err != nil {
		return nil, unifiederrors.WrapParsef(err, "failed to parse module %s", moduleSpec)
	}

	m := &Module{
		Name:      moduleName,
		Namespace: ns,
	}
	l.cache[path] = m

	return m, nil
}

// parseModuleContent parses a module file (header-only, no ---) and returns its namespace.
func parseModuleContent(content string, loader *ModuleLoader) (Namespace, error) {
	content = strings.TrimSpace(content)

	ns := make(Namespace)
	// For evaluation within the module, we need a raw map
	rawNs := make(map[string]interface{})
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
			fn, fnName, consumed, err := parseFunDeclFromLines(lines, i, rawNs)
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
func inferNamespaceEntry(v interface{}) NamespaceEntry {
	switch val := v.(type) {
	case *evaluator.Lambda:
		return NewFuncEntry(val)
	case *evaluator.TypeDef:
		return NewTypeDefEntry(val)
	default:
		return NewVarEntry(v)
	}
}

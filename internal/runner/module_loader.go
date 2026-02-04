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

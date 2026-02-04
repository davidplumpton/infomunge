//go:build js
// +build js

package runner

import (
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

// ModuleLoader handles loading and caching of modules in JS builds.
type ModuleLoader struct {
	BaseDir string             // Base directory for module resolution
	cache   map[string]*Module // module spec -> loaded module
	loading map[string]bool    // currently loading modules
}

// NewModuleLoader creates a module loader for JS builds with standard module support.
func NewModuleLoader(baseDir string) *ModuleLoader {
	return &ModuleLoader{
		BaseDir: baseDir,
		cache:   make(map[string]*Module),
		loading: make(map[string]bool),
	}
}

// Resolve returns metadata for standard modules in JS builds.
func (l *ModuleLoader) Resolve(moduleSpec string) (moduleName string, path string, err error) {
	if isStandardModule(moduleSpec) {
		return moduleNameFromSpec(moduleSpec), moduleSpec, nil
	}
	return "", "", unifiederrors.ParseError("imports are not supported in JS builds")
}

// Load supports standard modules and rejects file-based imports in JS builds.
func (l *ModuleLoader) Load(moduleSpec string) (*Module, error) {
	if !isStandardModule(moduleSpec) {
		return nil, unifiederrors.ParseError("imports are not supported in JS builds")
	}

	if m, ok := l.cache[moduleSpec]; ok {
		return m, nil
	}
	if l.loading[moduleSpec] {
		return nil, unifiederrors.ParseErrorf("cyclic import detected: %s", moduleSpec)
	}
	l.loading[moduleSpec] = true
	defer delete(l.loading, moduleSpec)

	content, err := standardModuleSource(moduleSpec)
	if err != nil {
		return nil, unifiederrors.WrapParsef(err, "failed to load standard module %s", moduleSpec)
	}
	ns, err := parseModuleContent(content, l)
	if err != nil {
		return nil, unifiederrors.WrapParsef(err, "failed to parse standard module %s", moduleSpec)
	}

	m := &Module{
		Name:      moduleNameFromSpec(moduleSpec),
		Namespace: ns,
	}
	l.cache[moduleSpec] = m
	return m, nil
}

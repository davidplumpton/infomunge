package runner

import (
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

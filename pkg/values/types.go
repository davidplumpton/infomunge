package values

// Value is the shared runtime value shape used by evaluation and format adapters.
type Value = interface{}

// Object represents a structured data object with string keys.
type Object = map[string]Value

// Array represents an ordered sequence of values.
type Array = []Value

// XMLMultiValue represents duplicate XML element values while preserving that
// they came from repeated object keys rather than a source array.
type XMLMultiValue []Value

// Namespace represents an XML namespace with an optional prefix.
// Prefix "" indicates the default namespace.
type Namespace struct {
	Prefix string
	URI    string
}

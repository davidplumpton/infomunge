package formats

import (
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
	"reflect"
	"strings"
	"testing"
)

func TestFormat_EmptyMimeType(t *testing.T) {
	_, err := Format("content", "")
	if err == nil {
		t.Error("expected error for empty mimeType")
	}
}

func TestStructuredWritersPreserveObjectOrder(t *testing.T) {
	object := values.NewObject(2)
	values.SetObjectValue(object, "b", 2)
	values.SetObjectValue(object, "a", 1)

	tests := []struct {
		mimeType string
		value    interface{}
		want     string
	}{
		{mimeType: "application/json", value: object, want: `{"b":2,"a":1}`},
		{mimeType: "application/yaml", value: object, want: "b: 2\na: 1\n"},
		{mimeType: "application/csv", value: Array{object}, want: "b,a\n2,1"},
	}
	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got, err := Format(tt.value, tt.mimeType)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormat_NilUsesSelectedCodec(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     string
		wantErr  string
	}{
		{name: "structured JSON", mimeType: "application/json", want: "null"},
		{name: "text", mimeType: "text/plain", want: "null"},
		{name: "structured CSV", mimeType: "application/csv", wantErr: "CSV output expects an array of objects"},
		{name: "raw", mimeType: "application/octet-stream", wantErr: "binary output expects string or []byte"},
		{name: "unknown", mimeType: "application/x-unknown", wantErr: "unsupported output mimeType"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(nil, tt.mimeType)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Format() error = nil, want it to contain %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Format() error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormat_JSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", `"hello"`},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"object", Object{"key": "value"}, `{"key":"value"}`},
		{"array", Array{1, 2, 3}, `[1,2,3]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Format(tt.input, "application/json")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormat_TextPlain(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"bool", true, "true"},
		{"object", Object{"key": "value"}, `{"key":"value"}`},
		{"array", Array{"a", "b"}, `["a","b"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Format(tt.input, "text/plain")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormat_XML(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			"simple object",
			Object{"root": "value"},
			"<?xml version='1.0' encoding='UTF-8'?>\n<root>value</root>",
		},
		{
			"nested object",
			Object{"root": Object{"child": "value"}},
			"<?xml version='1.0' encoding='UTF-8'?>\n<root><child>value</child></root>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Format(tt.input, "application/xml")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormat_XMLRejectsUnrepresentableDocuments(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		options *XMLOutputOptions
		wantErr string
	}{
		{
			name:    "multiple roots",
			input:   Object{"a": 1, "b": 2},
			wantErr: "XML output expects exactly one root element, got 2",
		},
		{
			name:    "non-object document",
			input:   "text",
			wantErr: "XML output expects an object containing exactly one root element",
		},
		{
			name:    "invalid element name",
			input:   Object{"bad key": 1},
			wantErr: `invalid XML element name "bad key"`,
		},
		{
			name:    "invalid attribute name",
			input:   Object{"root": Object{"@bad key": 1}},
			wantErr: `invalid XML attribute name "bad key"`,
		},
		{
			name:    "unbound namespace prefix",
			input:   Object{"missing:root": 1},
			wantErr: `namespace prefix "missing" is not declared`,
		},
		{
			name: "invalid resolved namespace prefix",
			input: Object{
				"source#root": 1,
			},
			options: &XMLOutputOptions{
				NamespaceVars: map[string]Namespace{
					"source": {Prefix: "bad prefix", URI: "urn:example"},
				},
				WriteDeclaration: true,
			},
			wantErr: `invalid XML element name "bad prefix:root"`,
		},
		{
			name: "reserved xml prefix with incorrect URI",
			input: Object{
				"source#root": 1,
			},
			options: &XMLOutputOptions{
				NamespaceVars: map[string]Namespace{
					"source": {Prefix: "xml", URI: "urn:not-the-xml-namespace"},
				},
				WriteDeclaration: true,
			},
			wantErr: `XML namespace prefix "xml" must use URI`,
		},
		{
			name: "invalid explicit namespace declaration",
			input: Object{
				"root": Object{
					"@xmlns": Object{"bad prefix": "urn:example"},
				},
			},
			wantErr: `invalid XML namespace prefix "bad prefix"`,
		},
		{
			name: "duplicate resolved attribute names",
			input: Object{
				"root": Object{
					"@first#id":  1,
					"@second#id": 2,
				},
			},
			options: &XMLOutputOptions{
				NamespaceVars: map[string]Namespace{
					"first":  {Prefix: "p", URI: "urn:first"},
					"second": {Prefix: "p", URI: "urn:second"},
				},
				WriteDeclaration: true,
			},
			wantErr: `XML element "root" contains duplicate resolved attribute name "p:id"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.options == nil {
				_, err = Format(tt.input, "application/xml")
			} else {
				_, err = FormatXMLWithOptions(tt.input, *tt.options)
			}
			if err == nil {
				t.Fatal("Format() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Format() error = %q, want it to contain %q", err, tt.wantErr)
			}
			typedErr, ok := err.(*unifiederrors.Error)
			if !ok {
				t.Fatalf("Format() error type = %T, want *errors.Error", err)
			}
			if typedErr.Type != unifiederrors.TypeValidate {
				t.Fatalf("Format() error category = %q, want %q", typedErr.Type, unifiederrors.TypeValidate)
			}
		})
	}
}

func TestFormat_XMLPreservesOrderedChildrenAndValidNamespaces(t *testing.T) {
	children := values.NewObject(4)
	values.SetObjectValue(children, "@xmlns", Object{"p": "urn:example"})
	values.SetObjectValue(children, "@p:id", "7")
	values.SetObjectValue(children, "p:b", 2)
	values.SetObjectValue(children, "p:a", 1)

	document := values.NewObject(1)
	values.SetObjectValue(document, "p:root", children)

	got, err := Format(document, "application/xml")
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	want := "<?xml version='1.0' encoding='UTF-8'?>\n" +
		`<p:root xmlns:p="urn:example" p:id="7"><p:b>2</p:b><p:a>1</p:a></p:root>`
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormat_CSV(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			"simple array",
			Array{
				Object{"name": "Alice", "age": "30"},
				Object{"name": "Bob", "age": "25"},
			},
			"age,name\n30,Alice\n25,Bob",
		},
		{
			"empty array",
			Array{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Format(tt.input, "application/csv")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormat_CSV_NonArray(t *testing.T) {
	// Non-array input returns validation error
	_, err := Format("not an array", "application/csv")
	if err == nil {
		t.Fatal("expected validation error for non-array CSV input")
	}
	if !strings.Contains(err.Error(), "CSV output expects an array of objects") {
		t.Errorf("expected validation error message, got: %v", err)
	}
}

func TestFormat_CSV_MissingKeys(t *testing.T) {
	// Test heterogeneous data where items have different keys
	input := Array{
		Object{"a": "1", "b": "2"},
		Object{"a": "3", "c": "4"},
	}
	result, err := Format(input, "application/csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should handle missing keys gracefully
	if result == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormat_Binary(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		result, err := Format("hello\x00world", "application/octet-stream")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "hello\x00world" {
			t.Fatalf("expected passthrough binary string, got %q", result)
		}
	})

	t.Run("byte slice input", func(t *testing.T) {
		result, err := Format([]byte{0x41, 0x42, 0x00, 0x43}, "application/octet-stream")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "AB\x00C" {
			t.Fatalf("unexpected formatted binary output: %q", result)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := Format(Object{"a": 1}, "application/octet-stream")
		if err == nil {
			t.Fatal("expected error for non-binary input")
		}
		if !strings.Contains(err.Error(), "binary output expects string or []byte") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestFormat_Avro(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		result, err := Format("avro\x00payload", "application/avro")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "avro\x00payload" {
			t.Fatalf("expected passthrough avro string, got %q", result)
		}
	})

	t.Run("byte slice input", func(t *testing.T) {
		result, err := Format([]byte{0x4f, 0x62, 0x6a, 0x01}, "application/avro")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Obj\x01" {
			t.Fatalf("unexpected formatted avro output: %q", result)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := Format(Object{"a": 1}, "application/avro")
		if err == nil {
			t.Fatal("expected error for non-binary input")
		}
		if !strings.Contains(err.Error(), "binary output expects string or []byte") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestFormat_DW(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		result, err := Format("%dw 2.0\n---\n42", "application/dw")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "%dw 2.0\n---\n42" {
			t.Fatalf("expected passthrough dw string, got %q", result)
		}
	})

	t.Run("byte slice input", func(t *testing.T) {
		result, err := Format([]byte("%dw 2.0\n---\ntrue"), "application/dw")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "%dw 2.0\n---\ntrue" {
			t.Fatalf("unexpected formatted dw output: %q", result)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := Format(Object{"a": 1}, "application/dw")
		if err == nil {
			t.Fatal("expected error for non-dw input")
		}
		if !strings.Contains(err.Error(), "binary output expects string or []byte") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestFormat_Flatfile(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		result, err := Format("HDR0001ALICE   000030NY ", "application/flatfile")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "HDR0001ALICE   000030NY " {
			t.Fatalf("expected passthrough flatfile string, got %q", result)
		}
	})

	t.Run("byte slice input", func(t *testing.T) {
		result, err := Format([]byte("DTL0002BOB     000025CA "), "application/flatfile")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "DTL0002BOB     000025CA " {
			t.Fatalf("unexpected formatted flatfile output: %q", result)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := Format(Object{"a": 1}, "application/flatfile")
		if err == nil {
			t.Fatal("expected error for non-flatfile input")
		}
		if !strings.Contains(err.Error(), "binary output expects string or []byte") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestFormatWithOptions_FlatfileStructured(t *testing.T) {
	input := Array{
		Object{"id": 1.0, "name": "ALICE", "age": 30.0, "state": "NY"},
		Object{"id": 2.0, "name": "BOB", "age": 25.0, "state": "CA"},
	}
	options := Object{
		"schema": Object{
			"fields": Array{
				Object{"name": "id", "length": 4, "type": "int", "align": "right", "pad": "0"},
				Object{"name": "name", "length": 8, "align": "left", "pad": " "},
				Object{"name": "age", "length": 6, "type": "int", "align": "right", "pad": "0"},
				Object{"name": "state", "length": 2},
			},
		},
	}

	result, err := FormatWithOptions(input, "application/flatfile", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "0001ALICE   000030NY\n0002BOB     000025CA"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFormatWithOptions_FlatfileStructuredLengthError(t *testing.T) {
	input := Object{"id": 1.0, "name": "TOO-LONG-NAME", "age": 30.0, "state": "NY"}
	options := Object{
		"schema": Object{
			"singleRecord": true,
			"fields": Array{
				Object{"name": "id", "length": 4, "type": "int", "align": "right", "pad": "0"},
				Object{"name": "name", "length": 8, "align": "left", "pad": " "},
				Object{"name": "age", "length": 6, "type": "int", "align": "right", "pad": "0"},
				Object{"name": "state", "length": 2},
			},
		},
	}

	_, err := FormatWithOptions(input, "application/flatfile", options)
	if err == nil {
		t.Fatal("expected error for value overflow")
	}
	if !strings.Contains(err.Error(), "value length") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormat_Java(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		result, err := Format("\xac\xed\x00\x05java-object", "application/java")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "\xac\xed\x00\x05java-object" {
			t.Fatalf("expected passthrough java string, got %q", result)
		}
	})

	t.Run("byte slice input", func(t *testing.T) {
		result, err := Format([]byte{0xac, 0xed, 0x00, 0x05, 0x01, 0x02}, "application/java")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "\xac\xed\x00\x05\x01\x02" {
			t.Fatalf("unexpected formatted java output: %q", result)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := Format(Object{"a": 1}, "application/java")
		if err == nil {
			t.Fatal("expected error for non-java input")
		}
		if !strings.Contains(err.Error(), "binary output expects string or []byte") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestFormatWithOptions_JavaStructured(t *testing.T) {
	result, err := FormatWithOptions(
		Object{"name": "Alice", "age": 30},
		"application/java",
		Object{"structured": true},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := Read(result, "application/json")
	if err != nil {
		t.Fatalf("expected structured java output to be valid json: %v", err)
	}

	expected := Object{
		"@class": "java.util.LinkedHashMap",
		"value":  Object{"name": "Alice", "age": 30.0},
	}
	if parsedObj, ok := parsed.(Object); !ok || !reflect.DeepEqual(parsedObj, expected) {
		t.Fatalf("expected %#v, got %#v", expected, parsed)
	}
}

func TestFormatWithOptions_JavaStructuredWithClassOverride(t *testing.T) {
	result, err := FormatWithOptions(
		Array{"a", "b"},
		"application/java",
		Object{
			"structured": true,
			"class":      "java.util.LinkedList",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := Read(result, "application/json")
	if err != nil {
		t.Fatalf("expected structured java output to be valid json: %v", err)
	}

	expected := Object{
		"@class": "java.util.LinkedList",
		"value":  Array{"a", "b"},
	}
	if parsedObj, ok := parsed.(Object); !ok || !reflect.DeepEqual(parsedObj, expected) {
		t.Fatalf("expected %#v, got %#v", expected, parsed)
	}
}

func TestFormatWithOptions_XLSXAliasUsesSameOptionHandler(t *testing.T) {
	options := Object{"unsupported": true}

	_, canonicalErr := FormatWithOptions(Object{"Sheet1": Array{}}, "application/xlsx", options)
	_, aliasErr := FormatWithOptions(Object{"Sheet1": Array{}}, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", options)

	if canonicalErr == nil || aliasErr == nil {
		t.Fatalf("expected both calls to fail, canonicalErr=%v aliasErr=%v", canonicalErr, aliasErr)
	}

	if canonicalErr.Error() != aliasErr.Error() {
		t.Fatalf("expected identical errors for canonical/alias, got %q vs %q", canonicalErr.Error(), aliasErr.Error())
	}

	if !strings.Contains(canonicalErr.Error(), "unsupported xlsx option") {
		t.Fatalf("expected xlsx option handler error, got %v", canonicalErr)
	}
}

func TestFormatWithOptions_JavaStructuredClassMismatch(t *testing.T) {
	_, err := FormatWithOptions(
		Object{"name": "Alice"},
		"application/java",
		Object{
			"structured": true,
			"class":      "java.util.List",
		},
	)
	if err == nil {
		t.Fatal("expected validation error for class/value mismatch")
	}
	if !strings.Contains(err.Error(), "is incompatible with value type") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatWithOptions_JavaStructuredUnknownClass(t *testing.T) {
	_, err := FormatWithOptions(
		Object{"name": "Alice"},
		"application/java",
		Object{
			"structured": true,
			"class":      "com.example.CustomType",
		},
	)
	if err == nil {
		t.Fatal("expected validation error for unknown class")
	}
	if !strings.Contains(err.Error(), "unsupported java class") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormat_Protobuf(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
	}{
		{name: "application/protobuf", mimeType: "application/protobuf"},
		{name: "application/x-protobuf", mimeType: "application/x-protobuf"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/string input", func(t *testing.T) {
			result, err := Format("\x0a\x05Alice\x10\x1e", tt.mimeType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != "\x0a\x05Alice\x10\x1e" {
				t.Fatalf("expected passthrough protobuf string, got %q", result)
			}
		})

		t.Run(tt.name+"/byte slice input", func(t *testing.T) {
			result, err := Format([]byte{0x0a, 0x03, 0x42, 0x6f, 0x62}, tt.mimeType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != "\x0a\x03Bob" {
				t.Fatalf("unexpected formatted protobuf output: %q", result)
			}
		})

		t.Run(tt.name+"/unsupported type", func(t *testing.T) {
			_, err := Format(Object{"a": 1}, tt.mimeType)
			if err == nil {
				t.Fatal("expected error for non-protobuf input")
			}
			if !strings.Contains(err.Error(), "binary output expects string or []byte") {
				t.Fatalf("unexpected error message: %v", err)
			}
		})
	}
}

func TestFormatWithOptions_ProtobufStructured(t *testing.T) {
	options := Object{
		"structured": true,
		"schema": Object{
			"message": "Person",
			"fields": Array{
				Object{"number": 1, "name": "name", "type": "string"},
				Object{"number": 2, "name": "age", "type": "int32"},
				Object{"number": 3, "name": "active", "type": "bool"},
			},
		},
	}

	input := Object{
		"name":   "Alice",
		"age":    30.0,
		"active": true,
	}

	result, err := FormatWithOptions(input, "application/protobuf", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "\x0a\x05Alice\x10\x1e\x18\x01" {
		t.Fatalf("unexpected protobuf output: %q", result)
	}
}

func TestFormatWithOptions_ProtobufStructuredRepeated(t *testing.T) {
	options := Object{
		"structured": true,
		"schema": Object{
			"fields": Array{
				Object{"number": 1, "name": "tags", "type": "string", "repeated": true},
			},
		},
	}

	input := Object{
		"tags": Array{"a", "bb"},
	}

	result, err := FormatWithOptions(input, "application/x-protobuf", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "\x0a\x01a\x0a\x02bb" {
		t.Fatalf("unexpected protobuf output: %q", result)
	}
}

func TestFormatWithOptions_ProtobufStructuredUnknownField(t *testing.T) {
	options := Object{
		"structured": true,
		"schema": Object{
			"fields": Array{
				Object{"number": 1, "name": "name", "type": "string"},
			},
		},
	}

	_, err := FormatWithOptions(
		Object{"name": "Alice", "extra": "ignored"},
		"application/protobuf",
		options,
	)
	if err == nil {
		t.Fatal("expected error for unknown protobuf field")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatWithOptions_ProtobufStructuredTypeMismatch(t *testing.T) {
	options := Object{
		"structured": true,
		"schema": Object{
			"fields": Array{
				Object{"number": 1, "name": "active", "type": "bool"},
			},
		},
	}

	_, err := FormatWithOptions(
		Object{"active": "yes"},
		"application/protobuf",
		options,
	)
	if err == nil {
		t.Fatal("expected error for protobuf bool type mismatch")
	}
	if !strings.Contains(err.Error(), "expects bool") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatWithOptions_ProtobufStructuredDescriptorSet(t *testing.T) {
	options := Object{
		"structured": true,
		"descriptor": Object{
			"set":     testPersonDescriptorSetBytes(t),
			"message": "test.Person",
		},
	}

	input := Object{
		"name":          "Bob",
		"lucky_numbers": Array{1.0, 150.0},
	}

	result, err := FormatWithOptions(input, "application/protobuf", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "\x0a\x03Bob\x12\x03\x01\x96\x01" {
		t.Fatalf("unexpected protobuf output: %q", result)
	}
}

func TestFormatWithOptions_ProtobufStructuredDescriptorSetMapField(t *testing.T) {
	options := Object{
		"structured": true,
		"descriptor": Object{
			"set":     testMapDescriptorSetBytes(t),
			"message": "t.M",
		},
	}

	input := Object{
		"kv": Object{
			"1": 10.0,
			"2": 20.0,
		},
	}

	result, err := FormatWithOptions(input, "application/protobuf", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "\x0a\x04\x08\x01\x10\x0a\x0a\x04\x08\x02\x10\x14" {
		t.Fatalf("unexpected protobuf output: %q", result)
	}
}

func TestFormatWithOptions_ProtobufStructuredDescriptorSetMapFieldInvalidKey(t *testing.T) {
	options := Object{
		"structured": true,
		"descriptor": Object{
			"set":     testMapDescriptorSetBytes(t),
			"message": "t.M",
		},
	}

	input := Object{
		"kv": Object{
			"abc": 10.0,
		},
	}

	_, err := FormatWithOptions(input, "application/protobuf", options)
	if err == nil {
		t.Fatal("expected error for invalid map key")
	}
	if !strings.Contains(err.Error(), `map key "abc" is invalid`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatWithOptions_ProtobufStructuredPackedRepeated(t *testing.T) {
	options := Object{
		"structured": true,
		"schema": Object{
			"fields": Array{
				Object{"number": 1, "name": "ids", "type": "int32", "repeated": true, "packed": true},
			},
		},
	}

	input := Object{
		"ids": Array{1.0, 2.0, 300.0},
	}

	result, err := FormatWithOptions(input, "application/protobuf", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "\x0a\x04\x01\x02\xac\x02" {
		t.Fatalf("unexpected protobuf output: %q", result)
	}
}

func TestFormat_Excel(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
	}{
		{name: "application/xlsx", mimeType: "application/xlsx"},
		{name: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", mimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/string input", func(t *testing.T) {
			result, err := Format("PK\x03\x04xlsx-content", tt.mimeType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != "PK\x03\x04xlsx-content" {
				t.Fatalf("expected passthrough excel string, got %q", result)
			}
		})

		t.Run(tt.name+"/byte slice input", func(t *testing.T) {
			result, err := Format([]byte("PK\x03\x04sheet1.xml"), tt.mimeType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != "PK\x03\x04sheet1.xml" {
				t.Fatalf("unexpected formatted excel output: %q", result)
			}
		})

		t.Run(tt.name+"/unsupported type", func(t *testing.T) {
			_, err := Format(Object{"a": 1}, tt.mimeType)
			if err == nil {
				t.Fatal("expected error for non-excel input")
			}
			if !strings.Contains(err.Error(), "binary output expects string or []byte") {
				t.Fatalf("unexpected error message: %v", err)
			}
		})
	}
}

func TestFormat_Multipart(t *testing.T) {
	input := Object{
		"name": "Alice",
		"tags": Array{"alpha", "beta"},
		"upload": Object{
			"filename":    "hello.txt",
			"contentType": "text/plain",
			"content":     "Hello file",
		},
	}

	formatted, err := Format(input, "multipart/form-data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(formatted, "--infomunge-boundary") {
		t.Fatalf("expected deterministic boundary, got: %s", formatted)
	}
	if !strings.Contains(formatted, `name="upload"; filename="hello.txt"`) {
		t.Fatalf("expected file part headers, got: %s", formatted)
	}

	parsed, err := Read(formatted, "multipart/form-data")
	if err != nil {
		t.Fatalf("expected formatted multipart to be readable, got error: %v", err)
	}

	if parsedObj, ok := parsed.(Object); !ok {
		t.Fatalf("expected parsed object, got %T", parsed)
	} else {
		if parsedObj["name"] != "Alice" {
			t.Fatalf("expected name field to round-trip, got %#v", parsedObj["name"])
		}
		if !strings.Contains(marshalToJSON(parsedObj["tags"]), `["alpha","beta"]`) {
			t.Fatalf("expected repeated field round-trip, got %#v", parsedObj["tags"])
		}
	}
}

func TestFormat_Multipart_NonObject(t *testing.T) {
	_, err := Format(Array{1, 2, 3}, "multipart/form-data")
	if err == nil {
		t.Fatal("expected error for non-object multipart output")
	}
	if !strings.Contains(err.Error(), "multipart output expects an object") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFormat_UnknownMimeType(t *testing.T) {
	// Unknown mime types should return an error rather than silently falling back
	_, err := Format(42, "application/unknown")
	if err == nil {
		t.Fatal("expected error for unknown mime type, got nil")
	}
}

func TestFormat_XML_Namespaces(t *testing.T) {
	t.Run("prefixed namespace declaration", func(t *testing.T) {
		input := Object{
			"soap:Envelope": Object{
				"@xmlns": Object{
					"soap": "http://schemas.xmlsoap.org/soap/envelope/",
				},
				"soap:Body": "content",
			},
		}
		result, err := Format(input, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should contain the namespace declaration
		if !strings.Contains(result, "xmlns:soap=") {
			t.Errorf("expected xmlns:soap declaration, got: %s", result)
		}
		if !strings.Contains(result, "soap:Envelope") {
			t.Errorf("expected soap:Envelope element, got: %s", result)
		}
	})

	t.Run("default namespace declaration", func(t *testing.T) {
		input := Object{
			"root": Object{
				"@xmlns": Object{
					"#default": "http://example.com/ns",
				},
				"child": "value",
			},
		}
		result, err := Format(input, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should contain the default namespace declaration
		if !strings.Contains(result, `xmlns="http://example.com/ns"`) {
			t.Errorf("expected default xmlns declaration, got: %s", result)
		}
	})

	t.Run("attributes", func(t *testing.T) {
		input := Object{
			"root": Object{
				"@id":   "123",
				"child": "value",
			},
		}
		result, err := Format(input, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should contain the attribute
		if !strings.Contains(result, `id="123"`) {
			t.Errorf("expected id attribute, got: %s", result)
		}
	})

	t.Run("text content with attributes", func(t *testing.T) {
		input := Object{
			"element": Object{
				"@attr": "val",
				"#text": "text content",
			},
		}
		result, err := Format(input, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "text content") {
			t.Errorf("expected text content, got: %s", result)
		}
		if !strings.Contains(result, `attr="val"`) {
			t.Errorf("expected attr attribute, got: %s", result)
		}
	})
}

func TestBuildXMLAttributes(t *testing.T) {
	tests := []struct {
		name     string
		input    Object
		wantNS   []string
		wantAttr []string
	}{
		{
			name: "no attributes",
			input: Object{
				"child": "value",
			},
			wantNS:   []string{},
			wantAttr: []string{},
		},
		{
			name: "only attributes",
			input: Object{
				"@id":   "123",
				"@type": "test",
			},
			wantNS:   []string{},
			wantAttr: []string{`id="123"`, `type="test"`},
		},
		{
			name: "only default namespace",
			input: Object{
				"@xmlns": Object{
					"#default": "http://example.com",
				},
			},
			wantNS:   []string{`xmlns="http://example.com"`},
			wantAttr: []string{},
		},
		{
			name: "prefixed namespace",
			input: Object{
				"@xmlns": Object{
					"soap": "http://schemas.xmlsoap.org/soap/envelope/",
					"xs":   "http://www.w3.org/2001/XMLSchema",
				},
			},
			wantNS:   []string{`xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"`, `xmlns:xs="http://www.w3.org/2001/XMLSchema"`},
			wantAttr: []string{},
		},
		{
			name: "mixed attributes and namespaces",
			input: Object{
				"@id":    "123",
				"@xmlns": Object{"#default": "http://example.com"},
				"child":  "value",
			},
			wantNS:   []string{`xmlns="http://example.com"`},
			wantAttr: []string{`id="123"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNS, gotAttr := buildXMLAttributes(tt.input)

			if len(gotNS) != len(tt.wantNS) {
				t.Errorf("namespace count: got %d, want %d", len(gotNS), len(tt.wantNS))
			}
			for _, ns := range tt.wantNS {
				found := false
				for _, g := range gotNS {
					if g == ns {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected namespace %q not found in %v", ns, gotNS)
				}
			}

			if len(gotAttr) != len(tt.wantAttr) {
				t.Errorf("attribute count: got %d, want %d", len(gotAttr), len(tt.wantAttr))
			}
			for _, attr := range tt.wantAttr {
				found := false
				for _, g := range gotAttr {
					if g == attr {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected attribute %q not found in %v", attr, gotAttr)
				}
			}
		})
	}
}

func TestBuildXMLChildren(t *testing.T) {
	tests := []struct {
		name  string
		input Object
		want  int // number of children expected
	}{
		{
			name:  "no children",
			input: Object{"#text": "value"},
			want:  0,
		},
		{
			name:  "single child",
			input: Object{"child": "value"},
			want:  1,
		},
		{
			name:  "multiple children",
			input: Object{"a": "1", "b": "2", "c": "3"},
			want:  3,
		},
		{
			name:  "skips special keys",
			input: Object{"@attr": "val", "@xmlns": Object{}, "#text": "text", "child": "value"},
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildXMLChildren(tt.input)
			if len(got) != tt.want {
				t.Errorf("got %d children, want %d", len(got), tt.want)
			}
		})
	}
}

func TestBuildXMLContent(t *testing.T) {
	tests := []struct {
		name     string
		input    Object
		expected string
	}{
		{
			name:     "empty map",
			input:    Object{},
			expected: "",
		},
		{
			name:     "only text",
			input:    Object{"#text": "hello"},
			expected: "hello",
		},
		{
			name:     "only children",
			input:    Object{"child": "value"},
			expected: "<child>value</child>",
		},
		{
			name:     "children and text",
			input:    Object{"child": "value", "#text": "text"},
			expected: "<child>value</child>text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildXMLContent(tt.input)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBuildXMLOpeningTag(t *testing.T) {
	tests := []struct {
		name  string
		xmlns []string
		attrs []string
		want  string
	}{
		{
			name:  "no attributes",
			xmlns: []string{},
			attrs: []string{},
			want:  ">",
		},
		{
			name:  "only attributes",
			xmlns: []string{},
			attrs: []string{`id="1"`},
			want:  ` id="1">`,
		},
		{
			name:  "only namespaces",
			xmlns: []string{`xmlns="http://example.com"`},
			attrs: []string{},
			want:  ` xmlns="http://example.com">`,
		},
		{
			name:  "both namespaces and attributes",
			xmlns: []string{`xmlns="http://example.com"`},
			attrs: []string{`id="1"`},
			want:  ` xmlns="http://example.com" id="1">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildXMLOpeningTag(tt.xmlns, tt.attrs)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

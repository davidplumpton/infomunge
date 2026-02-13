package formats

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRead_EmptyMimeType(t *testing.T) {
	_, err := Read("content", "")
	if err == nil {
		t.Error("expected error for empty mimeType")
	}
}

func TestRead_EmptyContent(t *testing.T) {
	result, err := Read("", "application/json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty content, got %v", result)
	}
}

func TestRead_UnsupportedMimeType(t *testing.T) {
	_, err := Read("content", "application/unknown")
	if err == nil {
		t.Error("expected error for unsupported mimeType")
	}
}

func TestRead_JSON(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected interface{}
	}{
		{"object", `{"key": "value"}`, Object{"key": "value"}},
		{"array", `[1, 2, 3]`, Array{1.0, 2.0, 3.0}},
		{"string", `"hello"`, "hello"},
		{"number", `42`, 42.0},
		{"boolean", `true`, true},
		{"null", `null`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Read(tt.content, "application/json")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestRead_JSON_Invalid(t *testing.T) {
	_, err := Read("{invalid", "application/json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestRead_YAML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected interface{}
	}{
		{"simple object", "key: value", Object{"key": "value"}},
		{"array", "- 1\n- 2\n- 3", Array{1, 2, 3}},
		{"nested", "parent:\n  child: value", Object{"parent": Object{"child": "value"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Read(tt.content, "application/yaml")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRead_CSV(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected interface{}
	}{
		{
			"simple csv",
			"name,age\nAlice,30\nBob,25",
			Array{
				Object{"name": "Alice", "age": "30"},
				Object{"name": "Bob", "age": "25"},
			},
		},
		{
			"single row",
			"col1,col2\nval1,val2",
			Array{
				Object{"col1": "val1", "col2": "val2"},
			},
		},
		{
			"header only",
			"col1,col2",
			Array(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Read(tt.content, "application/csv")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRead_CSV_Errors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"empty header column", ",col2\nval1,val2"},
		{"mismatched columns", "col1,col2\nval1,val2,val3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Read(tt.content, "application/csv")
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestRead_CSV_Empty(t *testing.T) {
	result, err := Read("", "application/csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty CSV, got %v", result)
	}
}

func TestRead_Binary(t *testing.T) {
	content := "\x00\x01hello\xff"
	result, err := Read(content, "application/octet-stream")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := result.(string); !ok || got != content {
		t.Fatalf("expected binary reader to return original content, got %#v (%T)", result, result)
	}
}

func TestRead_Avro(t *testing.T) {
	content := "\x00Obj\x01avro\xff"
	result, err := Read(content, "application/avro")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := result.(string); !ok || got != content {
		t.Fatalf("expected avro reader to return original content, got %#v (%T)", result, result)
	}
}

func TestRead_DW(t *testing.T) {
	content := "%dw 2.0\noutput application/json\n---\n{answer: 42}"
	result, err := Read(content, "application/dw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := result.(string); !ok || got != content {
		t.Fatalf("expected dw reader to return original content, got %#v (%T)", result, result)
	}
}

func TestRead_Flatfile(t *testing.T) {
	content := "HDR0001ALICE   000030NY \nDTL0002BOB     000025CA "
	result, err := Read(content, "application/flatfile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := result.(string); !ok || got != content {
		t.Fatalf("expected flatfile reader to return original content, got %#v (%T)", result, result)
	}
}

func TestRead_Java(t *testing.T) {
	content := "\xac\xed\x00\x05sr\x00\x10java.lang.String"
	result, err := Read(content, "application/java")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, ok := result.(string); !ok || got != content {
		t.Fatalf("expected java reader to return original content, got %#v (%T)", result, result)
	}
}

func TestRead_Protobuf(t *testing.T) {
	content := "\x0a\x05Alice\x10\x1e"

	tests := []struct {
		name     string
		mimeType string
	}{
		{name: "application/protobuf", mimeType: "application/protobuf"},
		{name: "application/x-protobuf", mimeType: "application/x-protobuf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Read(content, tt.mimeType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got, ok := result.(string); !ok || got != content {
				t.Fatalf("expected protobuf reader to return original content, got %#v (%T)", result, result)
			}
		})
	}
}

func TestRead_Excel(t *testing.T) {
	content := "PK\x03\x04xlsx-content"

	tests := []struct {
		name     string
		mimeType string
	}{
		{name: "application/xlsx", mimeType: "application/xlsx"},
		{name: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", mimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Read(content, tt.mimeType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got, ok := result.(string); !ok || got != content {
				t.Fatalf("expected excel reader to return original content, got %#v (%T)", result, result)
			}
		})
	}
}

func TestRead_XML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected interface{}
	}{
		{
			"simple element",
			"<root><item>value</item></root>",
			Object{"root": Object{"item": "value"}},
		},
		{
			"multiple children",
			"<root><a>1</a><b>2</b></root>",
			Object{"root": Object{"a": "1", "b": "2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Read(tt.content, "application/xml")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRead_XML_Limits(t *testing.T) {
	t.Run("attribute count limit", func(t *testing.T) {
		var b strings.Builder
		b.WriteString("<root")
		for i := 0; i <= MaxXMLAttributesPerElement; i++ {
			b.WriteString(fmt.Sprintf(` a%d="v"`, i))
		}
		b.WriteString("></root>")

		_, err := Read(b.String(), "application/xml")
		if err == nil || !strings.Contains(err.Error(), "attribute count exceeded") {
			t.Fatalf("expected attribute count limit error, got: %v", err)
		}
	})

	t.Run("element count limit", func(t *testing.T) {
		content := "<root>" + strings.Repeat("<n/>", MaxXMLElementCount) + "</root>"

		_, err := Read(content, "application/xml")
		if err == nil || !strings.Contains(err.Error(), "element count exceeded") {
			t.Fatalf("expected element count limit error, got: %v", err)
		}
	})

	t.Run("text bytes per element limit", func(t *testing.T) {
		content := "<root>" + strings.Repeat("a", MaxXMLTextBytesPerElement+1) + "</root>"

		_, err := Read(content, "application/xml")
		if err == nil || !strings.Contains(err.Error(), "text size exceeded") {
			t.Fatalf("expected text size limit error, got: %v", err)
		}
	})
}

func TestRead_XML_Namespaces(t *testing.T) {
	t.Run("prefixed namespace", func(t *testing.T) {
		content := `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body>content</soap:Body></soap:Envelope>`
		result, err := Read(content, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m, ok := result.(Object)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		// Should have soap:Envelope as key
		if _, ok := m["soap:Envelope"]; !ok {
			t.Errorf("expected 'soap:Envelope' key, got keys: %v", m)
		}
	})

	t.Run("default namespace", func(t *testing.T) {
		content := `<root xmlns="http://example.com/ns"><child>value</child></root>`
		result, err := Read(content, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m, ok := result.(Object)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		// Should have root as key
		if _, ok := m["root"]; !ok {
			t.Errorf("expected 'root' key, got keys: %v", m)
		}

		// Check xmlns declaration is captured
		rootMap, ok := m["root"].(Object)
		if !ok {
			t.Fatalf("expected root to be map, got %T", m["root"])
		}
		if _, ok := rootMap["@xmlns"]; !ok {
			t.Errorf("expected '@xmlns' in root, got keys: %v", rootMap)
		}
	})

	t.Run("multiple namespaces", func(t *testing.T) {
		content := `<root xmlns:a="http://a.com" xmlns:b="http://b.com"><a:child>A</a:child><b:child>B</b:child></root>`
		result, err := Read(content, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m, ok := result.(Object)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		rootMap, ok := m["root"].(Object)
		if !ok {
			t.Fatalf("expected root to be map, got %T", m["root"])
		}

		// Should have both prefixed children
		if _, ok := rootMap["a:child"]; !ok {
			t.Errorf("expected 'a:child' in root, got: %v", rootMap)
		}
		if _, ok := rootMap["b:child"]; !ok {
			t.Errorf("expected 'b:child' in root, got: %v", rootMap)
		}
	})

	t.Run("attributes with namespaces", func(t *testing.T) {
		content := `<root xmlns:x="http://x.com" x:attr="value">text</root>`
		result, err := Read(content, "application/xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m, ok := result.(Object)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}

		rootMap, ok := m["root"].(Object)
		if !ok {
			t.Fatalf("expected root to be map, got %T", m["root"])
		}

		// Should have namespaced attribute
		if _, ok := rootMap["@x:attr"]; !ok {
			t.Errorf("expected '@x:attr' in root, got: %v", rootMap)
		}
	})
}

func TestRead_Properties(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected interface{}
	}{
		{
			"equals separator",
			"key=value",
			Object{"key": "value"},
		},
		{
			"colon separator",
			"key:value",
			Object{"key": "value"},
		},
		{
			"comment lines",
			"# comment\nkey=value\n! another comment",
			Object{"key": "value"},
		},
		{
			"empty lines",
			"key1=val1\n\nkey2=val2",
			Object{"key1": "val1", "key2": "val2"},
		},
		{
			"trim whitespace",
			"  key  =  value  ",
			Object{"key": "value"},
		},
		{
			"escaped separators",
			`path\=name\:id=value\=1\:2`,
			Object{"path=name:id": "value=1:2"},
		},
		{
			"multiline continuation",
			"key=hello\\\n world",
			Object{"key": "helloworld"},
		},
		{
			"empty key",
			"=value",
			Object{"": "value"},
		},
		{
			"first separator wins when both present",
			"key:va=lue",
			Object{"key": "va=lue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Read(tt.content, "text/x-java-properties")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldSimplifyNode(t *testing.T) {
	tests := []struct {
		name string
		node Object
		want bool
	}{
		{
			name: "single text node",
			node: Object{XMLTextKey: "text"},
			want: true,
		},
		{
			name: "node with text and attributes",
			node: Object{XMLTextKey: "text", "@attr": "value"},
			want: false,
		},
		{
			name: "node with text and child element",
			node: Object{XMLTextKey: "text", "child": "value"},
			want: false,
		},
		{
			name: "node with text and namespace",
			node: Object{XMLTextKey: "text", XMLNamespaceKey: Object{}},
			want: false,
		},
		{
			name: "empty node",
			node: Object{},
			want: false,
		},
		{
			name: "node with only attribute",
			node: Object{"@id": "123"},
			want: false,
		},
		{
			name: "node with only child",
			node: Object{"child": "value"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSimplifyNode(tt.node)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateXMLBrackets(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid simple XML",
			content: "<root><child>text</child></root>",
			wantErr: false,
		},
		{
			name:    "valid nested XML",
			content: "<root><a><b><c>text</c></b></a></root>",
			wantErr: false,
		},
		{
			name:    "valid self-closing tag",
			content: "<root><empty/><child>text</child></root>",
			wantErr: false,
		},
		{
			name:    "unclosed tag",
			content: "<root><child>text</root>",
			wantErr: true,
		},
		{
			name:    "closing tag without opening",
			content: "<root></child></root>",
			wantErr: true,
		},
		{
			name:    "mismatched closing tag",
			content: "<root><child>text</different></root>",
			wantErr: true,
		},
		{
			name:    "unclosed root",
			content: "<root><child>text</child>",
			wantErr: true,
		},
		{
			name:    "with XML declaration",
			content: `<?xml version="1.0"?><root><child>text</child></root>`,
			wantErr: false,
		},
		{
			name:    "with comment",
			content: `<root><!-- comment --><child>text</child></root>`,
			wantErr: false,
		},
		{
			name:    "with DOCTYPE",
			content: `<!DOCTYPE root><root><child>text</child></root>`,
			wantErr: false,
		},
		{
			name:    "empty root element",
			content: "<root></root>",
			wantErr: false,
		},
		{
			name:    "multiple root elements (allowed for fragments)",
			content: "<a></a><b></b>",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateXMLBracketsWithStateMachine(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateXMLBracketsWithStateMachine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSimplifyXML(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "scalar value",
			input:    "text",
			expected: "text",
		},
		{
			name:     "map with single text node",
			input:    Object{XMLTextKey: "text"},
			expected: "text",
		},
		{
			name: "map with text and child",
			input: Object{
				XMLTextKey: "text",
				"child":    "value",
			},
			expected: map[string]interface{}{
				XMLTextKey: "text",
				"child":    "value",
			},
		},
		{
			name: "nested simplification",
			input: Object{
				"outer": Object{
					XMLTextKey: "content",
				},
			},
			expected: map[string]interface{}{
				"outer": "content",
			},
		},
		{
			name: "array with maps to simplify",
			input: Array{
				Object{XMLTextKey: "first"},
				Object{XMLTextKey: "second"},
			},
			expected: []interface{}{
				"first",
				"second",
			},
		},
		{
			name:     "empty array",
			input:    Array{},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simplifyXML(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNamespaceDeclsNormalization(t *testing.T) {
	decls := namespaceDeclsFromNodeObject(Object{
		"#default": "http://example.com/default",
		"ns":       "http://example.com/ns",
	})

	if decls[""] != "http://example.com/default" {
		t.Fatalf("expected default namespace to normalize to empty prefix, got %q", decls[""])
	}
	if decls["ns"] != "http://example.com/ns" {
		t.Fatalf("expected ns prefix to be preserved, got %q", decls["ns"])
	}

	node := decls.toNodeObject()
	if node["#default"] != "http://example.com/default" {
		t.Fatalf("expected default namespace to round-trip as #default, got %v", node["#default"])
	}
	if node["ns"] != "http://example.com/ns" {
		t.Fatalf("expected ns prefix to round-trip, got %v", node["ns"])
	}
}

func TestXMLNamespaceContextShadowing(t *testing.T) {
	var ctx xmlNamespaceContext
	ctx.Push(xmlNamespaceDecls{"ns": "http://example.com/one"})
	ctx.Push(xmlNamespaceDecls{"ns": "http://example.com/two"})

	name := ctx.ResolveElementName(xml.Name{Space: "http://example.com/two", Local: "item"})
	if name != "ns:item" {
		t.Fatalf("expected child namespace prefix resolution, got %q", name)
	}

	ctx.Pop()
	name = ctx.ResolveElementName(xml.Name{Space: "http://example.com/one", Local: "item"})
	if name != "ns:item" {
		t.Fatalf("expected parent namespace prefix resolution after pop, got %q", name)
	}
}

func TestResult(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		r := Ok("hello")
		if !r.IsOk() {
			t.Error("expected IsOk() to be true")
		}
		if r.IsErr() {
			t.Error("expected IsErr() to be false")
		}
		val, err := r.Unwrap()
		if val != "hello" {
			t.Errorf("expected value 'hello', got %v", val)
		}
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if r.Error() != nil {
			t.Errorf("expected nil error, got %v", r.Error())
		}
	})

	t.Run("Err result", func(t *testing.T) {
		r := Err[string](fmt.Errorf("test error"))
		if r.IsOk() {
			t.Error("expected IsOk() to be false")
		}
		if !r.IsErr() {
			t.Error("expected IsErr() to be true")
		}
		if r.Error() == nil {
			t.Error("expected non-nil error")
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		r := Ok(42)
		val, err := r.Unwrap()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("OrElse", func(t *testing.T) {
		ok := Ok("value")
		if ok.OrElse("default") != "value" {
			t.Error("OrElse should return value for Ok result")
		}

		err := Err[string](fmt.Errorf("error"))
		if err.OrElse("default") != "default" {
			t.Error("OrElse should return default for Err result")
		}
	})
}

func TestReadAsObject(t *testing.T) {
	t.Run("valid object", func(t *testing.T) {
		r := ReadAsObject(`{"key": "value"}`, "application/json")
		if !r.IsOk() {
			t.Fatalf("unexpected error: %v", r.Error())
		}
		obj, _ := r.Unwrap()
		if obj["key"] != "value" {
			t.Errorf("expected key='value', got %v", obj["key"])
		}
	})

	t.Run("non-object returns error", func(t *testing.T) {
		r := ReadAsObject(`[1, 2, 3]`, "application/json")
		if r.IsOk() {
			t.Error("expected error for array input")
		}
	})

	t.Run("empty content returns nil", func(t *testing.T) {
		r := ReadAsObject("", "application/json")
		if !r.IsOk() {
			t.Fatalf("unexpected error: %v", r.Error())
		}
		val, _ := r.Unwrap()
		if val != nil {
			t.Errorf("expected nil, got %v", val)
		}
	})
}

func TestReadAsArray(t *testing.T) {
	t.Run("valid array", func(t *testing.T) {
		r := ReadAsArray(`[1, 2, 3]`, "application/json")
		if !r.IsOk() {
			t.Fatalf("unexpected error: %v", r.Error())
		}
		arr, _ := r.Unwrap()
		if len(arr) != 3 {
			t.Errorf("expected 3 elements, got %d", len(arr))
		}
	})

	t.Run("non-array returns error", func(t *testing.T) {
		r := ReadAsArray(`{"key": "value"}`, "application/json")
		if r.IsOk() {
			t.Error("expected error for object input")
		}
	})
}

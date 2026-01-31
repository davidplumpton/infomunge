package formats

import (
	"strings"
	"testing"
)

func TestFormat_EmptyMimeType(t *testing.T) {
	_, err := Format("content", "")
	if err == nil {
		t.Error("expected error for empty mimeType")
	}
}

func TestFormat_Nil(t *testing.T) {
	result, err := Format(nil, "application/json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "null" {
		t.Errorf("expected 'null', got %q", result)
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

func TestFormat_XML(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			"simple object",
			Object{"root": "value"},
			"<root>value</root>",
		},
		{
			"nested object",
			Object{"root": Object{"child": "value"}},
			"<root><child>value</child></root>",
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

func TestFormat_DefaultJSON(t *testing.T) {
	// Unknown mime types should default to JSON
	result, err := Format(42, "application/unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "42" {
		t.Errorf("expected '42', got %q", result)
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

func TestValidateFormatInput(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "valid map",
			input:   Object{"key": "value"},
			wantErr: false,
		},
		{
			name:    "valid array",
			input:   Array{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "valid string",
			input:   "text",
			wantErr: false,
		},
		{
			name:    "valid number",
			input:   42,
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFormatInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFormatInput() error = %v, wantErr %v", err, tt.wantErr)
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

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

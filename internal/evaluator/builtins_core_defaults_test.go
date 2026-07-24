package evaluator

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseDefaultValueDecodesStringLiteralEscapes(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want Value
	}{
		{
			name: "scalar",
			src:  `"line1\nline2\t\"quoted\"\\path\u263A"`,
			want: "line1\nline2\t\"quoted\"\\path☺",
		},
		{
			name: "nested array",
			src:  `["first\nline", ["tab\tvalue", "snowman: \u2603"]]`,
			want: Array{"first\nline", Array{"tab\tvalue", "snowman: ☃"}},
		},
		{
			name: "nested object",
			src:  `{message: "hello\nworld", nested: {quote: "\"yes\"", slash: "\\"}}`,
			want: Object{
				"message": "hello\nworld",
				"nested": Object{
					"quote": "\"yes\"",
					"slash": "\\",
				},
			},
		},
		{
			name: "rewritten array and object",
			src:  `[]interface{}{"a\tb", map[string]interface{}{"value": "x\u263Ay",},}`,
			want: Array{"a\tb", Object{"value": "x☺y"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDefaultValue(tt.src)
			if err != nil {
				t.Fatalf("parseDefaultValue(%q) returned error: %v", tt.src, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseDefaultValue(%q) = %#v, want %#v", tt.src, got, tt.want)
			}
		})
	}
}

func TestParseDefaultValueRejectsMalformedStringLiteralEscape(t *testing.T) {
	_, err := parseDefaultValue(`"bad\q"`)
	if err == nil {
		t.Fatal("parseDefaultValue accepted malformed string literal escape")
	}
	if !strings.Contains(err.Error(), "invalid string literal") {
		t.Fatalf("error = %q, want invalid string literal context", err)
	}
}

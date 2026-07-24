package evaluator

import (
	"go/ast"
	"testing"
	"unicode/utf8"
)

func TestUnicodeTextTransforms(t *testing.T) {
	tests := []struct {
		name      string
		transform func([]Value, *ast.CallExpr) (Value, error)
		input     string
		want      string
	}{
		{
			name:      "capitalize accented initial",
			transform: callBuiltinCapitalize,
			input:     "éclair",
			want:      "Éclair",
		},
		{
			name:      "capitalize emoji-leading word",
			transform: callBuiltinCapitalize,
			input:     "🙂 smile",
			want:      "🙂 Smile",
		},
		{
			name:      "capitalize non-ASCII camel case",
			transform: callBuiltinCapitalize,
			input:     "déjàVu",
			want:      "Déjà Vu",
		},
		{
			name:      "camelize accented initial",
			transform: callBuiltinCamelize,
			input:     "Éclair_test",
			want:      "éclairTest",
		},
		{
			name:      "camelize emoji-leading part",
			transform: callBuiltinCamelize,
			input:     "🙂_smile",
			want:      "🙂Smile",
		},
		{
			name:      "dasherize non-ASCII camel case",
			transform: callBuiltinDasherize,
			input:     "déjàVu",
			want:      "déjà-vu",
		},
		{
			name:      "underscore non-ASCII camel case",
			transform: callBuiltinUnderscore,
			input:     "déjàVu",
			want:      "déjà_vu",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.transform([]Value{tt.input}, &ast.CallExpr{})
			if err != nil {
				t.Fatalf("transform returned an error: %v", err)
			}

			got, ok := value.(string)
			if !ok {
				t.Fatalf("transform returned %T, want string", value)
			}
			if !utf8.ValidString(got) {
				t.Fatalf("transform returned invalid UTF-8: %q", got)
			}
			if got != tt.want {
				t.Fatalf("transform returned %q, want %q", got, tt.want)
			}
		})
	}
}

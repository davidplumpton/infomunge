package cli

import (
	"errors"
	"fmt"
	"testing"

	unifiederrors "infomunge/internal/errors"
)

func TestIsClientFormattingError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "parse error",
			err:  unifiederrors.ParseError("invalid output option"),
			want: true,
		},
		{
			name: "validation error",
			err:  unifiederrors.ValidationError("unsupported output mimeType"),
			want: true,
		},
		{
			name: "wrapped validation error",
			err:  fmt.Errorf("format output: %w", unifiederrors.ValidationError("invalid shape")),
			want: true,
		},
		{
			name: "internal error",
			err:  unifiederrors.InternalError("writer failed"),
		},
		{
			name: "untyped error",
			err:  errors.New("writer failed"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := isClientFormattingError(test.err); got != test.want {
				t.Fatalf("isClientFormattingError(%v) = %t, want %t", test.err, got, test.want)
			}
		})
	}
}

func TestSanitizeClientErrorMessagePreservesMIMETypeAndRedactsPaths(t *testing.T) {
	t.Parallel()

	err := errors.New(`unsupported output mimeType: application/x-unknown; schema "/tmp/private/schema.json" failed`)
	got := sanitizeClientErrorMessage(err)
	want := `unsupported output mimeType: application/x-unknown; schema "<path>" failed`
	if got != want {
		t.Fatalf("sanitizeClientErrorMessage() = %q, want %q", got, want)
	}
}

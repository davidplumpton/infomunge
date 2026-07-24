package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	unifiederrors "infomunge/internal/errors"
)

type cancelRequestAfterChecksContext struct {
	context.Context
	checks      atomic.Int64
	cancelAfter int64
	cancel      context.CancelFunc
}

func (c *cancelRequestAfterChecksContext) Err() error {
	if c.checks.Add(1) >= c.cancelAfter {
		c.cancel()
	}
	return c.Context.Err()
}

func TestRunHandlerStopsInFlightEvaluationAfterRequestCancellation(t *testing.T) {
	payload, err := json.Marshal(map[string]string{
		"script": "%im 0.1\noutput application/json\n---\nwhile(true) { 1 }",
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	deadlineCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	goCtx := &cancelRequestAfterChecksContext{
		Context:     deadlineCtx,
		cancelAfter: 50,
		cancel:      cancel,
	}
	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(payload)).WithContext(goCtx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	NewApp().ServerHandler(&Config{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("run handler status = %d, want %d; body = %q", rec.Code, http.StatusRequestTimeout, rec.Body.String())
	}
	if checks := goCtx.checks.Load(); checks < goCtx.cancelAfter {
		t.Fatalf("request context checked %d times, want at least %d to prove in-flight cancellation", checks, goCtx.cancelAfter)
	}
}

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

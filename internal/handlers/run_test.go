package handlers

import (
	"context"
	"testing"

	"infomunge/internal/runner"
)

func TestResolveAndFormatExecutionResultFormatsHeaderlessValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		script    string
		requested string
		wantMime  string
		wantBody  string
	}{
		{name: "scalar JSON", script: "42", requested: "json", wantMime: "application/json", wantBody: "42"},
		{name: "array JSON", script: "[1, 2]", requested: "json", wantMime: "application/json", wantBody: "[1,2]"},
		{name: "object JSON", script: `{name: "Alice"}`, requested: "json", wantMime: "application/json", wantBody: `{"name":"Alice"}`},
		{name: "text", script: `"plain text"`, requested: "text/plain", wantMime: "text/plain", wantBody: "plain text"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			execution, err := runner.ExecuteString(context.Background(), test.script, nil, runner.RunnerOptions{})
			if err != nil {
				t.Fatalf("ExecuteString() error = %v", err)
			}

			formatted, mimeType, err := ResolveAndFormatExecutionResult(execution, test.requested)
			if err != nil {
				t.Fatalf("ResolveAndFormatExecutionResult() error = %v", err)
			}
			if mimeType != test.wantMime {
				t.Fatalf("MIME type = %q, want %q", mimeType, test.wantMime)
			}
			if formatted != test.wantBody {
				t.Fatalf("formatted result = %q, want %q", formatted, test.wantBody)
			}
		})
	}
}

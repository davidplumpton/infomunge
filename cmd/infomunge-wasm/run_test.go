package main

import (
	"encoding/json"
	"testing"
)

func TestExecutePayloadFormatsHeaderlessValues(t *testing.T) {
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

			payload, err := json.Marshal(map[string]string{
				"script": test.script,
				"output": test.requested,
			})
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			response := executePayload(string(payload))
			if !response.ok {
				t.Fatalf("executePayload() failed: %s", response.errMessage)
			}
			if response.mimeType != test.wantMime {
				t.Fatalf("MIME type = %q, want %q", response.mimeType, test.wantMime)
			}
			if response.result != test.wantBody {
				t.Fatalf("formatted result = %q, want %q", response.result, test.wantBody)
			}
		})
	}
}

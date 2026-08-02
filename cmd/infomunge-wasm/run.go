package main

import (
	"context"
	"strings"

	"infomunge/internal/handlers"
	"infomunge/internal/runner"
)

type payloadResponse struct {
	ok         bool
	result     string
	mimeType   string
	errMessage string
}

// executePayload contains the platform-independent portion of the WASM
// adapter so its request, execution, and formatting contract can be tested
// without requiring a JavaScript runtime.
func executePayload(payload string) payloadResponse {
	request, err := handlers.DecodeRunRequest(strings.NewReader(payload))
	if err != nil {
		return payloadError(err)
	}

	evalContext, err := handlers.BuildRunContext(request.Inputs)
	if err != nil {
		return payloadError(err)
	}

	execution, err := runner.ExecuteString(context.Background(), request.Script, evalContext, runner.RunnerOptions{BaseDir: "."})
	if err != nil {
		return payloadError(err)
	}

	formatted, outputMimeType, err := handlers.ResolveAndFormatExecutionResult(execution, request.Output)
	if err != nil {
		return payloadError(err)
	}

	return payloadResponse{ok: true, result: formatted, mimeType: outputMimeType}
}

func payloadError(err error) payloadResponse {
	return payloadResponse{errMessage: err.Error()}
}

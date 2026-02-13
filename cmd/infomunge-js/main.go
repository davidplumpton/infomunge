//go:build js
// +build js

package main

import (
	"context"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"infomunge/internal/handlers"
	"infomunge/internal/runner"
)

func main() {
	js.Global.Set("infomungeRun", func(payload string) map[string]interface{} {
		return runPayload(payload)
	})
}

func runPayload(payload string) map[string]interface{} {
	request, err := handlers.DecodeRunRequest(strings.NewReader(payload))
	if err != nil {
		return errorResponse(err)
	}

	evalContext, err := handlers.BuildRunContext(request.Inputs)
	if err != nil {
		return errorResponse(err)
	}

	opts := runner.RunnerOptions{
		BaseDir: ".",
	}
	result, _, headerOutputMimeType, evalCtx, err := runner.RunStringWithGoContextAndOptionsWithOutput(context.Background(), request.Script, evalContext, opts)
	if err != nil {
		return errorResponse(err)
	}

	outputMimeType, err := handlers.ResolveOutputMimeType(request.Output, headerOutputMimeType)
	if err != nil {
		return errorResponse(err)
	}

	formatted, err := handlers.FormatRunResult(result, outputMimeType, evalCtx)
	if err != nil {
		return errorResponse(err)
	}

	return map[string]interface{}{
		"ok":       true,
		"result":   formatted,
		"mimeType": outputMimeType,
	}
}

func errorResponse(err error) map[string]interface{} {
	return map[string]interface{}{
		"ok":    false,
		"error": err.Error(),
	}
}

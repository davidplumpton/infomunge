//go:build js
// +build js

package main

import (
	"context"
	"strings"
	"syscall/js"

	"infomunge/internal/handlers"
	"infomunge/internal/runner"
)

var infomungeRun js.Func

func main() {
	infomungeRun = js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return errorResponseString("missing payload argument")
		}
		return runPayload(args[0].String())
	})

	js.Global().Set("infomungeRun", infomungeRun)
	select {}
}

func runPayload(payload string) js.Value {
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
	execution, err := runner.ExecuteStringWithGoContextAndOptions(context.Background(), request.Script, evalContext, opts)
	if err != nil {
		return errorResponse(err)
	}

	outputMimeType, err := handlers.ResolveOutputMimeType(request.Output, execution.OutputMimeType)
	if err != nil {
		return errorResponse(err)
	}

	execution.OutputMimeType = outputMimeType
	formatted, err := runner.FormatExecutionResult(execution)
	if err != nil {
		return errorResponse(err)
	}

	return successResponse(formatted, outputMimeType)
}

func successResponse(result, mimeType string) js.Value {
	response := js.Global().Get("Object").New()
	response.Set("ok", true)
	response.Set("result", result)
	response.Set("mimeType", mimeType)
	response.Set("error", "")
	return response
}

func errorResponse(err error) js.Value {
	return errorResponseString(err.Error())
}

func errorResponseString(message string) js.Value {
	response := js.Global().Get("Object").New()
	response.Set("ok", false)
	response.Set("result", "")
	response.Set("mimeType", "")
	response.Set("error", message)
	return response
}

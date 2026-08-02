//go:build js
// +build js

package main

import (
	"syscall/js"
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
	response := executePayload(payload)
	if !response.ok {
		return errorResponseString(response.errMessage)
	}

	return successResponse(response.result, response.mimeType)
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

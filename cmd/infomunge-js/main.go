//go:build js
// +build js

package main

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/runner"
	"infomunge/pkg/formats"
)

type runInput struct {
	Name    string `json:"name"`
	Format  string `json:"format,omitempty"`
	Content string `json:"content"`
}

type runRequest struct {
	Script string     `json:"script"`
	Output string     `json:"output"`
	Inputs []runInput `json:"inputs,omitempty"`
}

func main() {
	js.Global.Set("infomungeRun", func(payload string) map[string]interface{} {
		return runPayload(payload)
	})
}

func runPayload(payload string) map[string]interface{} {
	request, err := decodeRunRequest(payload)
	if err != nil {
		return errorResponse(err)
	}

	context, err := buildRunContext(request.Inputs)
	if err != nil {
		return errorResponse(err)
	}

	opts := runner.RunnerOptions{
		BaseDir: ".",
	}
	result, _, headerOutputMimeType, evalCtx, err := runner.RunStringWithContextAndOptionsWithOutput(request.Script, context, opts)
	if err != nil {
		return errorResponse(err)
	}

	outputMimeType, err := resolveOutputMimeType(request.Output, headerOutputMimeType)
	if err != nil {
		return errorResponse(err)
	}

	formatted, err := formatRunResult(result, outputMimeType, evalCtx)
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

func decodeRunRequest(payload string) (*runRequest, error) {
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()

	var request runRequest
	if err := decoder.Decode(&request); err != nil {
		return nil, unifiederrors.ValidationErrorf("invalid JSON body: %v", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return nil, err
	}

	request.Script = strings.TrimSpace(request.Script)
	if request.Script == "" {
		return nil, unifiederrors.ValidationError("script is required")
	}

	request.Output = strings.TrimSpace(request.Output)

	return &request, nil
}

func ensureEOF(decoder *json.Decoder) error {
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return unifiederrors.ValidationError("invalid JSON body: unexpected trailing data")
	}
	return nil
}

func resolveOutputMimeType(requested string, headerOutput string) (string, error) {
	if strings.TrimSpace(requested) != "" {
		return normalizeMimeType(requested, "output")
	}
	return normalizeMimeType(headerOutput, "output")
}

func normalizeMimeType(format string, label string) (string, error) {
	format = strings.TrimSpace(format)
	if format == "" {
		return "text/plain", nil
	}
	if strings.Contains(format, "/") {
		return format, nil
	}
	mimeType := formats.MimeTypeForFormat(format)
	if mimeType == "" {
		return "", unifiederrors.ValidationErrorf("unknown %s format: %s", label, format)
	}
	return mimeType, nil
}

func buildRunContext(inputs []runInput) (map[string]interface{}, error) {
	context := make(map[string]interface{})
	for _, input := range inputs {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return nil, unifiederrors.ValidationError("input name is required")
		}

		mimeType, err := normalizeMimeType(input.Format, "input")
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("input %q: %v", name, err)
		}

		value, err := formats.Read(input.Content, mimeType)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("input %q: %v", name, err)
		}

		context[name] = value
	}
	return context, nil
}

func formatRunResult(result interface{}, mimeType string, evalCtx map[string]interface{}) (string, error) {
	if mimeType == "application/xml" {
		var nsMap map[string]string
		if namespaces, ok := evalCtx["__namespaces__"].(map[string]string); ok {
			nsMap = namespaces
		}
		xmlOpts := formats.XMLOutputOptions{
			DeclaredNamespaces: nsMap,
			NamespaceVars:      extractNamespaceVars(evalCtx),
			WriteDeclaration:   true,
		}
		if rawOpts, ok := evalCtx["__output_options__"].(map[string]string); ok {
			if err := applyXMLOutputOptions(&xmlOpts, rawOpts); err != nil {
				return "", err
			}
		}
		return formats.FormatXMLWithOptions(result, xmlOpts)
	}
	return formats.Format(result, mimeType)
}

func extractNamespaceVars(context map[string]interface{}) map[string]formats.Namespace {
	if context == nil {
		return nil
	}
	vars := make(map[string]formats.Namespace)
	for k, v := range context {
		if ns, ok := v.(formats.Namespace); ok {
			vars[k] = ns
		}
	}
	if len(vars) == 0 {
		return nil
	}
	return vars
}

func applyXMLOutputOptions(opts *formats.XMLOutputOptions, raw map[string]string) error {
	for key, value := range raw {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "writedeclaration":
			parsed, err := strconv.ParseBool(strings.TrimSpace(value))
			if err != nil {
				return unifiederrors.ParseErrorf("output option writeDeclaration: %v", err)
			}
			opts.WriteDeclaration = parsed
		case "writedeclarednamespaces":
			opts.WriteDeclaredNamespaces = value
		case "writenilonnull":
			parsed, err := strconv.ParseBool(strings.TrimSpace(value))
			if err != nil {
				return unifiederrors.ParseErrorf("output option writeNilOnNull: %v", err)
			}
			opts.WriteNilOnNull = parsed
		case "skipnullon":
			opts.SkipNullOn = value
		}
	}
	return nil
}

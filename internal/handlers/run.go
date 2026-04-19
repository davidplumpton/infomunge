package handlers

import (
	"encoding/json"
	"io"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	inputio "infomunge/internal/io"
	"infomunge/internal/runner"
	"infomunge/pkg/formats"
)

const maxRunInputs = 100

// RunInput represents a single named input with optional format.
type RunInput struct {
	Name    string `json:"name"`
	Format  string `json:"format,omitempty"`
	Content string `json:"content"`
}

// RunRequest represents a script execution request.
type RunRequest struct {
	Script string     `json:"script"`
	Output string     `json:"output"`
	Inputs []RunInput `json:"inputs,omitempty"`
}

// DecodeRunRequest reads and validates a JSON run request from the given reader.
func DecodeRunRequest(body io.Reader) (*RunRequest, error) {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	var request RunRequest
	if err := decoder.Decode(&request); err != nil {
		return nil, unifiederrors.WrapValidationf(err, "invalid JSON body: %v", err)
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

// ResolveOutputMimeType picks the output MIME type from the request or header fallback.
func ResolveOutputMimeType(requested string, headerOutput string) (string, error) {
	if strings.TrimSpace(requested) != "" {
		return NormalizeMimeType(requested, "output")
	}
	return NormalizeMimeType(headerOutput, "output")
}

// NormalizeMimeType converts a short format name (e.g. "json") to a full MIME type,
// or passes through an already-qualified MIME type.
func NormalizeMimeType(format string, label string) (string, error) {
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

// BuildRunContext parses named inputs into an evaluation context map.
func BuildRunContext(inputs []RunInput) (evaluator.Context, error) {
	if len(inputs) > maxRunInputs {
		return nil, unifiederrors.ValidationErrorf("too many inputs: maximum %d", maxRunInputs)
	}

	context := make(evaluator.Context)
	for _, input := range inputs {
		name, err := inputio.NormalizeAndValidateInputName(input.Name)
		if err != nil {
			return nil, err
		}
		if _, exists := context[name]; exists {
			return nil, unifiederrors.ValidationErrorf("duplicate input name %q", name)
		}

		mimeType, err := NormalizeMimeType(input.Format, "input")
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

// FormatRunResult formats the evaluation result into the appropriate output string.
func FormatRunResult(result evaluator.Value, mimeType string, evalCtx evaluator.Context) (string, error) {
	if mimeType == "application/xml" {
		var nsMap map[string]string
		if namespaces, ok := evalCtx[runner.ContextKeyNamespaces].(map[string]string); ok {
			nsMap = namespaces
		}
		xmlOpts := formats.XMLOutputOptions{
			DeclaredNamespaces: nsMap,
			NamespaceVars:      runner.ExtractNamespaceVars(evalCtx),
			WriteDeclaration:   true,
		}
		if rawOpts, ok := evalCtx[runner.ContextKeyOutputOptions].(map[string]string); ok {
			if err := runner.ApplyXMLOutputOptions(&xmlOpts, rawOpts); err != nil {
				return "", err
			}
		}
		return formats.FormatXMLWithOptions(result, xmlOpts)
	}
	return formats.Format(result, mimeType)
}

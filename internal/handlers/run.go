package handlers

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/runner"
	"infomunge/pkg/formats"
)

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
func BuildRunContext(inputs []RunInput) (map[string]interface{}, error) {
	context := make(map[string]interface{})
	for _, input := range inputs {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return nil, unifiederrors.ValidationError("input name is required")
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
func FormatRunResult(result interface{}, mimeType string, evalCtx map[string]interface{}) (string, error) {
	if mimeType == "application/xml" {
		var nsMap map[string]string
		if namespaces, ok := evalCtx[runner.ContextKeyNamespaces].(map[string]string); ok {
			nsMap = namespaces
		}
		xmlOpts := formats.XMLOutputOptions{
			DeclaredNamespaces: nsMap,
			NamespaceVars:      ExtractNamespaceVars(evalCtx),
			WriteDeclaration:   true,
		}
		if rawOpts, ok := evalCtx[runner.ContextKeyOutputOptions].(map[string]string); ok {
			if err := ApplyXMLOutputOptions(&xmlOpts, rawOpts); err != nil {
				return "", err
			}
		}
		return formats.FormatXMLWithOptions(result, xmlOpts)
	}
	return formats.Format(result, mimeType)
}

// ExtractNamespaceVars finds all Namespace values in the eval context.
func ExtractNamespaceVars(context map[string]interface{}) map[string]formats.Namespace {
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

// ApplyXMLOutputOptions applies raw string options to XMLOutputOptions.
func ApplyXMLOutputOptions(opts *formats.XMLOutputOptions, raw map[string]string) error {
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

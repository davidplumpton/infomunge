package output

import (
	"fmt"
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/pkg/formats"
)

// Metadata carries output-formatting directives parsed from the script header.
type Metadata struct {
	DeclaredNamespaces map[string]string
	Options            map[string]string
}

// SetDeclaredNamespaces records namespace declarations used by XML output.
func SetDeclaredNamespaces(metadata *Metadata, namespaces map[string]string) {
	if metadata == nil || len(namespaces) == 0 {
		return
	}
	metadata.DeclaredNamespaces = copyStringMap(namespaces)
}

// SetOptions records raw output options parsed from the output directive.
func SetOptions(metadata *Metadata, options map[string]string) {
	if metadata == nil || len(options) == 0 {
		return
	}
	metadata.Options = copyStringMap(options)
}

// FormatResult formats an evaluated result without header output metadata.
func FormatResult(result evaluator.Value, mimeType string, evalCtx evaluator.Context) (string, error) {
	return FormatResultWithMetadata(result, mimeType, evalCtx, Metadata{})
}

// FormatResultWithMetadata formats an evaluated result using explicit output metadata.
func FormatResultWithMetadata(result evaluator.Value, mimeType string, evalCtx evaluator.Context, metadata Metadata) (string, error) {
	if mimeType == "application/xml" {
		xmlOpts, err := XMLOptions(evalCtx, metadata)
		if err != nil {
			return "", err
		}
		return formats.FormatXMLWithOptions(result, xmlOpts)
	}
	return formats.Format(result, mimeType)
}

// XMLOptions builds XML output options from explicit metadata and namespace vars.
func XMLOptions(evalCtx evaluator.Context, metadata Metadata) (formats.XMLOutputOptions, error) {
	opts := formats.XMLOutputOptions{
		DeclaredNamespaces: metadata.DeclaredNamespaces,
		NamespaceVars:      extractNamespaceVars(evalCtx),
		WriteDeclaration:   true,
	}
	if err := applyXMLOutputOptions(&opts, metadata.Options); err != nil {
		return formats.XMLOutputOptions{}, err
	}
	return opts, nil
}

func extractNamespaceVars(context evaluator.Context) map[string]formats.Namespace {
	if context == nil {
		return nil
	}
	vars := make(map[string]formats.Namespace)
	for key, value := range context {
		if namespace, ok := value.(formats.Namespace); ok {
			vars[key] = namespace
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
			parsed, err := parseBoolOption(value)
			if err != nil {
				return unifiederrors.ParseErrorf("output option writeDeclaration: %v", err)
			}
			opts.WriteDeclaration = parsed
		case "writedeclarednamespaces":
			opts.WriteDeclaredNamespaces = value
		case "writenilonnull":
			parsed, err := parseBoolOption(value)
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

func parseBoolOption(value string) (bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false, fmt.Errorf("empty value")
	}
	return strconv.ParseBool(trimmed)
}

func copyStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

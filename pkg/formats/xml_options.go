package formats

import (
	"fmt"
	"regexp"
	"strings"
)

// XMLSchemaInstanceURI is the namespace URI for xsi:nil usage.
const XMLSchemaInstanceURI = "http://www.w3.org/2001/XMLSchema-instance"

// XMLOutputOptions configures XML formatting behavior.
type XMLOutputOptions struct {
	DeclaredNamespaces      map[string]string
	NamespaceVars           map[string]Namespace
	WriteDeclaration        bool
	WriteDeclaredNamespaces string
	WriteNilOnNull          bool
	SkipNullOn              string
}

type xmlRenderOptions struct {
	declaredNamespaces   map[string]string
	rootDeclaredNames    map[string]string
	namespaceVars        map[string]Namespace
	writeDeclaration     bool
	writeNilOnNull       bool
	skipNullOnElements   bool
	skipNullOnAttributes bool
}

func normalizeXMLOptions(opts XMLOutputOptions) (xmlRenderOptions, error) {
	normalized := xmlRenderOptions{
		declaredNamespaces: opts.DeclaredNamespaces,
		namespaceVars:      opts.NamespaceVars,
		writeDeclaration:   opts.WriteDeclaration,
		writeNilOnNull:     opts.WriteNilOnNull,
	}

	var err error
	normalized.skipNullOnElements, normalized.skipNullOnAttributes, err = parseSkipNullOn(opts.SkipNullOn)
	if err != nil {
		return normalized, err
	}

	normalized.rootDeclaredNames, err = selectDeclaredNamespaces(opts.DeclaredNamespaces, opts.WriteDeclaredNamespaces)
	if err != nil {
		return normalized, err
	}

	return normalized, nil
}

func parseSkipNullOn(value string) (skipElements bool, skipAttributes bool, err error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	switch trimmed {
	case "":
		return false, false, nil
	case "elements":
		return true, false, nil
	case "attributes":
		return false, true, nil
	case "everywhere":
		return true, true, nil
	default:
		return false, false, fmt.Errorf("invalid skipNullOn value: %q", value)
	}
}

func selectDeclaredNamespaces(declared map[string]string, rule string) (map[string]string, error) {
	if len(declared) == 0 {
		return nil, nil
	}
	trimmed := strings.TrimSpace(rule)
	if trimmed == "" {
		return nil, nil
	}

	lower := strings.ToLower(trimmed)
	if lower == "all" {
		return copyNamespaceMap(declared), nil
	}
	if strings.HasPrefix(lower, "ids:") {
		list := strings.Split(trimmed[len("ids:"):], ",")
		selected := make(map[string]string)
		for _, item := range list {
			name := strings.TrimSpace(item)
			if name == "" {
				continue
			}
			if uri, ok := declared[name]; ok {
				selected[name] = uri
			}
		}
		return selected, nil
	}
	if strings.HasPrefix(lower, "regex:") {
		pattern := strings.TrimSpace(trimmed[len("regex:"):])
		if pattern == "" {
			return nil, fmt.Errorf("writeDeclaredNamespaces regex cannot be empty")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid writeDeclaredNamespaces regex: %w", err)
		}
		selected := make(map[string]string)
		for prefix, uri := range declared {
			if re.MatchString(prefix) {
				selected[prefix] = uri
			}
		}
		return selected, nil
	}

	return nil, fmt.Errorf("invalid writeDeclaredNamespaces value: %q", rule)
}

func copyNamespaceMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

package formats

import (
	"encoding/xml"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"io"
	"sort"
	"strings"
)

// XML special key constants for attribute and text node representation
const (
	XMLTextKey      = "#text"  // Key for text content in element maps
	XMLNamespaceKey = "@xmlns" // Key for XML namespace declarations
	XMLAttrPrefix   = "@"      // Prefix for XML attributes
	MaxXMLDepth     = 512      // Maximum XML element nesting depth to prevent billion laughs attacks
)

func init() {
	RegisterReader("application/xml", readXML)
	RegisterWriter("application/xml", formatXML)
	RegisterExtension(".xml", "application/xml")
}

func readXML(content string) (interface{}, error) {
	if err := validateXMLBracketsWithStateMachine(content); err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(strings.NewReader(content))
	var stack []Object
	var result Object
	var nsStack []map[string]string // namespace prefix -> URI mappings at each level

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, unifiederrors.WrapValidationf(err, "XML parse error: %v", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Check depth limit to prevent billion laughs attacks
			if len(stack) >= MaxXMLDepth {
				return nil, unifiederrors.ValidationErrorf("XML element nesting depth exceeded (max %d levels)", MaxXMLDepth)
			}

			newNode, elemName, newNsStack := handleXMLStartElement(t, nsStack)
			nsStack = newNsStack

			if len(stack) > 0 {
				addChildToParent(stack[len(stack)-1], elemName, newNode)
			} else {
				result = Object{elemName: newNode}
			}
			stack = append(stack, newNode)

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			if len(nsStack) > 0 {
				nsStack = nsStack[:len(nsStack)-1]
			}

		case xml.CharData:
			if str := strings.TrimSpace(string(t)); str != "" && len(stack) > 0 {
				appendTextContent(stack[len(stack)-1], str)
			}
		}
	}

	return simplifyXML(result), nil
}

func formatXML(result interface{}) (string, error) {
	return formatXMLWithNamespaces(result, nil)
}

// formatXMLWithNamespaces formats result to XML, optionally applying declared namespaces.
func formatXMLWithNamespaces(result interface{}, declaredNs map[string]string) (string, error) {
	return toXMLWithNamespaces(result, "", declaredNs), nil
}

// Note: The old validateXMLBrackets and helper functions (handleClosingTag, handleOpeningTag,
// handleComment, handleProcessingInstruction) have been replaced with an explicit state machine
// implementation in xml_state_machine.go. The state machine approach provides clearer control
// flow and easier maintenance compared to the nested conditional approach.

// handleXMLStartElement processes a start element, extracting namespaces and attributes.
func handleXMLStartElement(elem xml.StartElement, nsStack []map[string]string) (Object, string, []map[string]string) {
	newNode := make(Object)

	// Collect namespace declarations from attributes
	nsDecls := collectNamespaceDecls(elem.Attr)

	// Push namespace context (inherit from parent + new declarations)
	nsStack = pushNamespaceContext(nsStack, nsDecls)

	// Build element name with prefix if namespace is present
	elemName := buildElementName(elem.Name, nsStack)

	// Store namespace declarations in the node
	if len(nsDecls) > 0 {
		xmlnsMap := make(Object)
		for prefix, uri := range nsDecls {
			if prefix == "" {
				xmlnsMap["#default"] = uri
			} else {
				xmlnsMap[prefix] = uri
			}
		}
		newNode[XMLNamespaceKey] = xmlnsMap
	}

	// Store non-namespace attributes
	for _, attr := range elem.Attr {
		if attr.Name.Space != "xmlns" && !(attr.Name.Local == "xmlns" && attr.Name.Space == "") {
			attrName := XMLAttrPrefix + buildElementName(attr.Name, nsStack)
			newNode[attrName] = attr.Value
		}
	}

	return newNode, elemName, nsStack
}

// collectNamespaceDecls extracts namespace declarations from XML attributes.
func collectNamespaceDecls(attrs []xml.Attr) map[string]string {
	nsDecls := make(map[string]string)
	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" {
			nsDecls[attr.Name.Local] = attr.Value // xmlns:prefix="uri"
		} else if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
			nsDecls[""] = attr.Value // xmlns="uri" (default namespace)
		}
	}
	return nsDecls
}

// pushNamespaceContext creates a new namespace context inheriting from parent.
func pushNamespaceContext(nsStack []map[string]string, nsDecls map[string]string) []map[string]string {
	if len(nsStack) > 0 {
		inherited := make(map[string]string)
		for k, v := range nsStack[len(nsStack)-1] {
			inherited[k] = v
		}
		for k, v := range nsDecls {
			inherited[k] = v
		}
		return append(nsStack, inherited)
	}
	return append(nsStack, nsDecls)
}

// addChildToParent adds a child node to a parent, handling repeated elements as arrays.
func addChildToParent(parent Object, elemName string, child Object) {
	if existing, ok := parent[elemName]; ok {
		switch v := existing.(type) {
		case XMLMultiValue:
			parent[elemName] = append(v, child)
		case Array:
			// If it's already a plain Array, it might have come from JSON or elsewhere,
			// but if we're in XML parser, we should probably treat it as MultiValue
			// if it's being added to.
			parent[elemName] = append(XMLMultiValue(v), child)
		default:
			parent[elemName] = XMLMultiValue{existing, child}
		}
	} else {
		parent[elemName] = child
	}
}

// appendTextContent appends text content to a node, joining with spaces.
// Uses strings.Builder internally to avoid O(n²) string concatenation for many text nodes.
func appendTextContent(node Object, text string) {
	if existing, ok := node[XMLTextKey]; ok {
		// If existing is a string, convert to builder and append
		if existingStr, ok := existing.(string); ok {
			var builder strings.Builder
			builder.WriteString(existingStr)
			builder.WriteString(" ")
			builder.WriteString(text)
			node[XMLTextKey] = builder.String()
		}
	} else {
		node[XMLTextKey] = text
	}
}

// buildElementName creates an element name, using prefix if a namespace is present
func buildElementName(name xml.Name, nsStack []map[string]string) string {
	if name.Space == "" {
		return name.Local
	}

	// Look up prefix for this namespace URI
	if len(nsStack) > 0 {
		current := nsStack[len(nsStack)-1]
		for prefix, uri := range current {
			if uri == name.Space && prefix != "" {
				return prefix + ":" + name.Local
			}
		}
	}

	// No prefix found, just use local name
	return name.Local
}

// normalizeTagName converts '#' in a key to ':' for XML tag or attribute names.
func normalizeTagName(name string) string {
	return strings.ReplaceAll(name, "#", ":")
}

// shouldSimplifyNode determines if a map node with a single #text key should be simplified to just the text.
func shouldSimplifyNode(node Object) bool {
	// Only simplify if the node has exactly one key and it's #text
	// Don't simplify if there are attributes (@...) or namespace declarations (@xmlns)
	if len(node) == 1 {
		_, hasText := node[XMLTextKey]
		return hasText
	}
	return false
}

// simplifyXMLMap recursively simplifies a map, converting single text nodes to their string values.
func simplifyXMLMap(node Object) interface{} {
	if shouldSimplifyNode(node) {
		return node[XMLTextKey]
	}
	for k, val := range node {
		node[k] = simplifyXML(val)
	}
	return node
}

// simplifyXMLArray recursively simplifies array elements.
func simplifyXMLArray(arr Array) Array {
	for i, val := range arr {
		arr[i] = simplifyXML(val)
	}
	return arr
}

// simplifyXMLMultiValue recursively simplifies XMLMultiValue elements.
func simplifyXMLMultiValue(arr XMLMultiValue) XMLMultiValue {
	for i, val := range arr {
		arr[i] = simplifyXML(val)
	}
	return arr
}

func simplifyXML(input interface{}) interface{} {
	switch v := input.(type) {
	case Object:
		return simplifyXMLMap(v)
	case XMLMultiValue:
		return simplifyXMLMultiValue(v)
	case Array:
		return simplifyXMLArray(v)
	default:
		return v
	}
}

// toXML converts an internal representation to XML.
func toXML(v interface{}, name string) string {
	return toXMLWithNamespaces(v, name, nil)
}

// toXMLWithNamespaces converts an internal representation to XML with optional declared namespaces.
// declaredNs maps namespace prefixes to URIs (e.g., {"ns0": "http://www.abc.com"}).
func toXMLWithNamespaces(v interface{}, name string, declaredNs map[string]string) string {
	return toXMLWithNamespacesRecursive(v, name, declaredNs, false)
}

func toXMLWithNamespacesRecursive(v interface{}, name string, declaredNs map[string]string, isChild bool) string {
	switch val := v.(type) {
	case Object:
		if name == "" {
			// Root level wrapper - find the first non-special key
			keys := make([]string, 0, len(val))
			for k := range val {
				if !strings.HasPrefix(k, XMLAttrPrefix) && k != XMLTextKey {
					keys = append(keys, k)
				}
			}
			if len(keys) > 0 {
				sort.Strings(keys)
				k := keys[0]
				tagName := normalizeTagName(k)
				return fmt.Sprintf("<%s%s</%s>", tagName, toXMLWithNamespacesRecursive(val[k], k, declaredNs, false), tagName)
			}
			return ""
		}

		// Non-root element
		xmlnsDecls, attrs := buildXMLAttributesWithNamespaces(val, declaredNs, isChild)
		content := buildXMLContentWithNamespaces(val, declaredNs)
		openingTag := buildXMLOpeningTag(xmlnsDecls, attrs)

		return openingTag + content

	default:
		return ">" + fmt.Sprintf("%v", val)
	}
}

// buildXMLAttributes builds XML namespace declarations and attribute strings.
func buildXMLAttributes(val Object) ([]string, []string) {
	return buildXMLAttributesWithNamespaces(val, nil, false)
}

// buildXMLAttributesWithNamespaces builds XML namespace declarations and attribute strings,
// optionally applying declared namespaces from the ns keyword.
func buildXMLAttributesWithNamespaces(val Object, declaredNs map[string]string, isChild bool) ([]string, []string) {
	var xmlnsDecls []string
	var attrs []string

	// Extract and sort attribute-related keys
	keys := extractAndSortAttributeKeys(val)

	// Process explicit namespaces and attributes from element
	for _, k := range keys {
		if k == XMLNamespaceKey {
			if nsMap, ok := val[k].(Object); ok {
				xmlnsDecls = append(xmlnsDecls, buildNamespaceDeclsFromMap(nsMap)...)
			}
		} else if strings.HasPrefix(k, XMLAttrPrefix) {
			attrName := normalizeTagName(k[len(XMLAttrPrefix):])
			attrs = append(attrs, fmt.Sprintf(`%s="%v"`, attrName, val[k]))
		}
	}

	// Apply declared namespaces if no explicit @xmlns in the element AND we are at the root
	if declaredNs != nil && len(declaredNs) > 0 && !isChild {
		if _, hasExplicitNs := val[XMLNamespaceKey]; !hasExplicitNs {
			xmlnsDecls = append(xmlnsDecls, buildNamespaceDeclsFromStringMap(declaredNs)...)
		}
	}

	// Sort for consistent output
	sort.Strings(xmlnsDecls)
	sort.Strings(attrs)

	return xmlnsDecls, attrs
}

// extractAndSortAttributeKeys extracts keys that represent XML attributes and namespaces.
func extractAndSortAttributeKeys(val Object) []string {
	keys := make([]string, 0, len(val))
	for k := range val {
		if k == XMLNamespaceKey || strings.HasPrefix(k, XMLAttrPrefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// buildNamespaceDecls is a generic function that builds namespace declaration strings
// from a map. It handles both interface{} maps (where default key is "#default")
// and string maps (where default key is "").
func buildNamespaceDecls[V any](nsMap map[string]V, defaultKeyValue string) []string {
	var decls []string
	for prefix, uri := range nsMap {
		if prefix == defaultKeyValue {
			decls = append(decls, fmt.Sprintf(`xmlns="%v"`, uri))
		} else {
			decls = append(decls, fmt.Sprintf(`xmlns:%s="%v"`, prefix, uri))
		}
	}
	return decls
}

// buildNamespaceDeclsFromMap builds namespace declaration strings from a map (parsed from @xmlns).
func buildNamespaceDeclsFromMap(nsMap Object) []string {
	return buildNamespaceDecls(nsMap, "#default")
}

// buildNamespaceDeclsFromStringMap builds namespace declaration strings from a string map (from ns keyword).
func buildNamespaceDeclsFromStringMap(nsMap map[string]string) []string {
	return buildNamespaceDecls(nsMap, "")
}

// buildXMLChildren builds child element strings from a map.
func buildXMLChildren(val Object) []string {
	return buildXMLChildrenWithNamespaces(val, nil)
}

// buildXMLChildrenWithNamespaces builds child element strings from a map,
// with optional declared namespaces passed to children.
func buildXMLChildrenWithNamespaces(val Object, declaredNs map[string]string) []string {
	// Extract and sort child element keys (non-special keys)
	keys := extractAndSortChildKeys(val)

	children := make([]string, 0, len(keys))
	for _, k := range keys {
		v := val[k]
		tagName := normalizeTagName(k)
		switch arr := v.(type) {
		case XMLMultiValue:
			for _, item := range arr {
				children = append(children, fmt.Sprintf("<%s%s</%s>", tagName, toXMLWithNamespacesRecursive(item, k, declaredNs, true), tagName))
			}
		case Array:
			for _, item := range arr {
				children = append(children, fmt.Sprintf("<%s%s</%s>", tagName, toXMLWithNamespacesRecursive(item, k, declaredNs, true), tagName))
			}
		default:
			children = append(children, fmt.Sprintf("<%s%s</%s>", tagName, toXMLWithNamespacesRecursive(v, k, declaredNs, true), tagName))
		}
	}

	return children
}

// extractAndSortChildKeys extracts keys that represent child elements (skip special keys).
func extractAndSortChildKeys(val Object) []string {
	keys := make([]string, 0, len(val))
	for k := range val {
		// Skip special keys: namespaces (@xmlns), attributes (@...), and text (#text)
		if k != XMLNamespaceKey && !strings.HasPrefix(k, XMLAttrPrefix) && k != XMLTextKey {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// buildXMLContent builds the inner content (children + text) of an element.
func buildXMLContent(val Object) string {
	return buildXMLContentWithNamespaces(val, nil)
}

// buildXMLContentWithNamespaces builds the inner content (children + text) of an element,
// with optional declared namespaces.
func buildXMLContentWithNamespaces(val Object, declaredNs map[string]string) string {
	var sb strings.Builder

	// Add child elements
	for _, child := range buildXMLChildrenWithNamespaces(val, declaredNs) {
		sb.WriteString(child)
	}

	// Add text content if present
	if text, ok := val[XMLTextKey]; ok {
		sb.WriteString(fmt.Sprintf("%v", text))
	}

	return sb.String()
}

// buildXMLOpeningTag builds the opening tag attributes string.
func buildXMLOpeningTag(xmlnsDecls, attrs []string) string {
	allAttrs := append(xmlnsDecls, attrs...)
	if len(allAttrs) > 0 {
		return " " + strings.Join(allAttrs, " ") + ">"
	}
	return ">"
}

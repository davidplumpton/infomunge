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
	var nsCtx xmlNamespaceContext

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

			newNode, elemName := handleXMLStartElement(t, &nsCtx)

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
			nsCtx.Pop()

		case xml.CharData:
			if str := strings.TrimSpace(string(t)); str != "" && len(stack) > 0 {
				appendTextContent(stack[len(stack)-1], str)
			}
		}
	}

	return simplifyXML(result), nil
}

func formatXML(result interface{}) (string, error) {
	return formatXMLWithOptions(result, XMLOutputOptions{WriteDeclaration: true})
}

// formatXMLWithNamespaces formats result to XML, optionally applying declared namespaces.
func formatXMLWithNamespaces(result interface{}, declaredNs map[string]string) (string, error) {
	return formatXMLWithOptions(result, XMLOutputOptions{
		DeclaredNamespaces: declaredNs,
		WriteDeclaration:   true,
	})
}

// formatXMLWithOptions formats result to XML using the supplied options.
func formatXMLWithOptions(result interface{}, opts XMLOutputOptions) (string, error) {
	renderOpts, err := normalizeXMLOptions(opts)
	if err != nil {
		return "", unifiederrors.ValidationErrorf("invalid XML output options: %v", err)
	}

	xml := toXMLWithOptions(result, "", renderOpts)
	if renderOpts.writeDeclaration {
		xml = "<?xml version='1.0' encoding='UTF-8'?>\n" + xml
	}
	return xml, nil
}

// Note: The old validateXMLBrackets and helper functions (handleClosingTag, handleOpeningTag,
// handleComment, handleProcessingInstruction) have been replaced with an explicit state machine
// implementation in xml_state_machine.go. The state machine approach provides clearer control
// flow and easier maintenance compared to the nested conditional approach.

// handleXMLStartElement processes a start element, extracting namespaces and attributes.
func handleXMLStartElement(elem xml.StartElement, nsCtx *xmlNamespaceContext) (Object, string) {
	newNode := make(Object)

	// Collect namespace declarations from attributes
	nsDecls := namespaceDeclsFromAttrs(elem.Attr)

	// Push namespace context (inherit from parent + new declarations)
	nsCtx.Push(nsDecls)

	// Build element name with prefix if namespace is present
	elemName := nsCtx.ResolveElementName(elem.Name)

	// Store namespace declarations in the node
	if len(nsDecls) > 0 {
		newNode[XMLNamespaceKey] = nsDecls.toNodeObject()
	}

	// Store non-namespace attributes
	for _, attr := range elem.Attr {
		if attr.Name.Space != "xmlns" && !(attr.Name.Local == "xmlns" && attr.Name.Space == "") {
			attrName := XMLAttrPrefix + nsCtx.ResolveElementName(attr.Name)
			newNode[attrName] = attr.Value
		}
	}

	return newNode, elemName
}

type xmlNamespaceDecls map[string]string

func newXMLNamespaceDecls() xmlNamespaceDecls {
	return make(xmlNamespaceDecls)
}

// namespaceDeclsFromAttrs extracts namespace declarations from XML attributes.
func namespaceDeclsFromAttrs(attrs []xml.Attr) xmlNamespaceDecls {
	nsDecls := newXMLNamespaceDecls()
	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" {
			nsDecls[attr.Name.Local] = attr.Value // xmlns:prefix="uri"
		} else if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
			nsDecls[""] = attr.Value // xmlns="uri" (default namespace)
		}
	}
	return nsDecls
}

func namespaceDeclsFromNodeObject(nsMap Object) xmlNamespaceDecls {
	nsDecls := newXMLNamespaceDecls()
	for prefix, uri := range nsMap {
		p := prefix
		if p == "#default" {
			p = ""
		}
		nsDecls[p] = fmt.Sprintf("%v", uri)
	}
	return nsDecls
}

func namespaceDeclsFromStringMap(nsMap map[string]string) xmlNamespaceDecls {
	nsDecls := newXMLNamespaceDecls()
	for prefix, uri := range nsMap {
		nsDecls[prefix] = uri
	}
	return nsDecls
}

func (nsDecls xmlNamespaceDecls) toNodeObject() Object {
	node := make(Object)
	for prefix, uri := range nsDecls {
		if prefix == "" {
			node["#default"] = uri
			continue
		}
		node[prefix] = uri
	}
	return node
}

func (nsDecls xmlNamespaceDecls) mergeInto(dst map[string]string) {
	for prefix, uri := range nsDecls {
		dst[prefix] = uri
	}
}

func (nsDecls xmlNamespaceDecls) declarationStrings() []string {
	decls := make([]string, 0, len(nsDecls))
	for prefix, uri := range nsDecls {
		if prefix == "" {
			decls = append(decls, fmt.Sprintf(`xmlns="%v"`, uri))
			continue
		}
		decls = append(decls, fmt.Sprintf(`xmlns:%s="%v"`, prefix, uri))
	}
	return decls
}

type xmlNamespaceContext struct {
	stack []map[string]string
}

func (ctx *xmlNamespaceContext) Push(nsDecls xmlNamespaceDecls) {
	merged := make(map[string]string)
	if current := ctx.current(); current != nil {
		for prefix, uri := range current {
			merged[prefix] = uri
		}
	}
	nsDecls.mergeInto(merged)
	ctx.stack = append(ctx.stack, merged)
}

func (ctx *xmlNamespaceContext) Pop() {
	if len(ctx.stack) == 0 {
		return
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
}

func (ctx *xmlNamespaceContext) current() map[string]string {
	if len(ctx.stack) == 0 {
		return nil
	}
	return ctx.stack[len(ctx.stack)-1]
}

func (ctx *xmlNamespaceContext) ResolveElementName(name xml.Name) string {
	return buildElementName(name, ctx.current())
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
func buildElementName(name xml.Name, current map[string]string) string {
	if name.Space == "" {
		return name.Local
	}

	// Look up prefix for this namespace URI
	if current != nil {
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

func simplifyXMLObject(node Object) interface{} {
	if shouldSimplifyNode(node) {
		return node[XMLTextKey]
	}
	for k, val := range node {
		node[k] = simplifyXML(val)
	}
	return node
}

func simplifyXMLSlice(length int, get func(int) interface{}, set func(int, interface{})) {
	for i := 0; i < length; i++ {
		set(i, simplifyXML(get(i)))
	}
}

func simplifyXML(input interface{}) interface{} {
	switch v := input.(type) {
	case Object:
		return simplifyXMLObject(v)
	case XMLMultiValue:
		simplifyXMLSlice(len(v), func(i int) interface{} { return v[i] }, func(i int, value interface{}) { v[i] = value })
		return v
	case Array:
		simplifyXMLSlice(len(v), func(i int) interface{} { return v[i] }, func(i int, value interface{}) { v[i] = value })
		return v
	default:
		return v
	}
}

// toXML converts an internal representation to XML.
func toXML(v interface{}, name string) string {
	return toXMLWithOptions(v, name, xmlRenderOptions{})
}

func toXMLWithOptions(v interface{}, name string, opts xmlRenderOptions) string {
	return toXMLWithOptionsRecursive(v, name, opts, false)
}

func toXMLWithOptionsRecursive(v interface{}, name string, opts xmlRenderOptions, isChild bool) string {
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
				return buildXMLElement(k, val[k], opts, false)
			}
			return ""
		}
		return buildElementContent(name, val, opts, isChild)
	default:
		if name == "" {
			return fmt.Sprintf("%v", v)
		}
		return buildElementContent(name, v, opts, isChild)
	}
}

func buildXMLElement(name string, value interface{}, opts xmlRenderOptions, isChild bool) string {
	if value == nil && opts.skipNullOnElements {
		return ""
	}
	content := buildElementContent(name, value, opts, isChild)
	if content == "" {
		return ""
	}
	resolvedName, _, _, _ := resolveName(name, opts.namespaceVars)
	tagName := normalizeTagName(resolvedName)
	return fmt.Sprintf("<%s%s</%s>", tagName, content, tagName)
}

func buildElementContent(name string, value interface{}, opts xmlRenderOptions, isChild bool) string {
	valObj, isObj := value.(Object)
	if !isObj {
		valObj = Object{}
	}
	elementIsNil := value == nil

	xmlnsDecls, attrs := buildXMLAttributesWithOptions(valObj, name, opts, isChild, elementIsNil)
	openingTag := buildXMLOpeningTag(xmlnsDecls, attrs)

	if elementIsNil {
		return openingTag
	}

	if isObj {
		return openingTag + buildXMLContentWithOptions(valObj, opts)
	}

	return openingTag + fmt.Sprintf("%v", value)
}

// buildXMLAttributes builds XML namespace declarations and attribute strings.
func buildXMLAttributes(val Object) ([]string, []string) {
	xmlnsDecls, attrs := buildXMLAttributesWithOptions(val, "", xmlRenderOptions{}, false, false)
	return xmlnsDecls, attrs
}

// buildXMLAttributesWithOptions builds XML namespace declarations and attribute strings.
func buildXMLAttributesWithOptions(val Object, elementName string, opts xmlRenderOptions, isChild bool, elementIsNil bool) ([]string, []string) {
	var attrs []string
	xmlnsMap := newXMLNamespaceDecls()

	_, elementPrefix, elementURI, elementHasPrefix := resolveName(elementName, opts.namespaceVars)

	// Extract and sort attribute-related keys
	keys := extractAndSortAttributeKeys(val)

	// Process explicit namespaces and attributes from element
	for _, k := range keys {
		if k == XMLNamespaceKey {
			if nsMap, ok := val[k].(Object); ok {
				namespaceDeclsFromNodeObject(nsMap).mergeInto(xmlnsMap)
			}
		} else if strings.HasPrefix(k, XMLAttrPrefix) {
			if opts.skipNullOnAttributes && val[k] == nil {
				continue
			}
			attrName := k[len(XMLAttrPrefix):]
			resolvedAttr, attrPrefix, attrURI, attrHasPrefix := resolveName(attrName, opts.namespaceVars)
			if attrHasPrefix {
				if attrURI == "" && opts.declaredNamespaces != nil {
					attrURI = opts.declaredNamespaces[attrPrefix]
				}
				if attrURI != "" {
					ensureNamespace(xmlnsMap, attrPrefix, attrURI)
				}
			}
			attrs = append(attrs, fmt.Sprintf(`%s="%s"`, normalizeTagName(resolvedAttr), formatXMLAttrValue(val[k])))
		}
	}

	// Root declared namespaces based on writeDeclaredNamespaces
	if !isChild {
		for prefix, uri := range namespaceDeclsFromStringMap(opts.rootDeclaredNames) {
			ensureNamespace(xmlnsMap, prefix, uri)
		}
	}

	// Namespaces needed for element name
	if elementHasPrefix {
		uri := elementURI
		if uri == "" && opts.declaredNamespaces != nil {
			uri = opts.declaredNamespaces[elementPrefix]
		}
		if uri != "" {
			ensureNamespace(xmlnsMap, elementPrefix, uri)
		}
	}

	if elementIsNil && opts.writeNilOnNull {
		attrs = append(attrs, `xsi:nil="true"`)
		ensureNamespace(xmlnsMap, "xsi", XMLSchemaInstanceURI)
	}

	xmlnsDecls := xmlnsMap.declarationStrings()
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

// buildXMLChildren builds child element strings from a map.
func buildXMLChildren(val Object) []string {
	return buildXMLChildrenWithOptions(val, xmlRenderOptions{})
}

// buildXMLChildrenWithOptions builds child element strings from a map.
func buildXMLChildrenWithOptions(val Object, opts xmlRenderOptions) []string {
	// Extract and sort child element keys (non-special keys)
	keys := extractAndSortChildKeys(val)

	children := make([]string, 0, len(keys))
	for _, k := range keys {
		v := val[k]
		switch arr := v.(type) {
		case XMLMultiValue:
			for _, item := range arr {
				if item == nil && opts.skipNullOnElements {
					continue
				}
				if child := buildXMLElement(k, item, opts, true); child != "" {
					children = append(children, child)
				}
			}
		case Array:
			for _, item := range arr {
				if item == nil && opts.skipNullOnElements {
					continue
				}
				if child := buildXMLElement(k, item, opts, true); child != "" {
					children = append(children, child)
				}
			}
		default:
			if v == nil && opts.skipNullOnElements {
				continue
			}
			if child := buildXMLElement(k, v, opts, true); child != "" {
				children = append(children, child)
			}
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
	return buildXMLContentWithOptions(val, xmlRenderOptions{})
}

// buildXMLContentWithOptions builds the inner content (children + text) of an element.
func buildXMLContentWithOptions(val Object, opts xmlRenderOptions) string {
	var sb strings.Builder

	// Add child elements
	for _, child := range buildXMLChildrenWithOptions(val, opts) {
		sb.WriteString(child)
	}

	// Add text content if present
	if text, ok := val[XMLTextKey]; ok && text != nil {
		sb.WriteString(fmt.Sprintf("%v", text))
	}

	return sb.String()
}

func resolveName(name string, nsVars map[string]Namespace) (resolvedName string, usedPrefix string, usedURI string, hasPrefix bool) {
	if name == "" {
		return name, "", "", false
	}
	idx := strings.Index(name, "#")
	if idx == -1 {
		return name, "", "", false
	}
	prefix := name[:idx]
	local := name[idx+1:]
	if nsVars != nil {
		if ns, ok := nsVars[prefix]; ok {
			if ns.Prefix == "" {
				return local, "", ns.URI, true
			}
			return ns.Prefix + "#" + local, ns.Prefix, ns.URI, true
		}
	}
	return name, prefix, "", true
}

func ensureNamespace(xmlns map[string]string, prefix string, uri string) {
	if uri == "" {
		return
	}
	if _, exists := xmlns[prefix]; exists {
		return
	}
	xmlns[prefix] = uri
}

func formatXMLAttrValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// buildXMLOpeningTag builds the opening tag attributes string.
func buildXMLOpeningTag(xmlnsDecls, attrs []string) string {
	allAttrs := append(xmlnsDecls, attrs...)
	if len(allAttrs) > 0 {
		return " " + strings.Join(allAttrs, " ") + ">"
	}
	return ">"
}

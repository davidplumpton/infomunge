package formats

import (
	"encoding/xml"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

// XML special key constants for attribute and text node representation
const (
	XMLTextKey = "#text" // Key for text content in element maps

	XMLNamespaceKey = "@xmlns" // Key for XML namespace declarations
	XMLAttrPrefix   = "@"      // Prefix for XML attributes

	MaxXMLDepth                = 512     // Maximum XML element nesting depth to prevent deep nesting attacks
	MaxXMLElementCount         = 100000  // Maximum number of elements in a document
	MaxXMLAttributesPerElement = 128     // Maximum number of attributes allowed on a single element
	MaxXMLTextBytesPerElement  = 1 << 20 // Maximum text bytes allowed in one element (1 MiB)
)

func init() {
	RegisterReader("application/xml", readXML)
	RegisterWriter("application/xml", formatXML)
	RegisterExtension(".xml", "application/xml")
}

func readXML(content string) (interface{}, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	var stack []Object
	var textSizes []int
	var result Object
	var nsCtx xmlNamespaceContext
	elementCount := 0

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
			if len(t.Attr) > MaxXMLAttributesPerElement {
				return nil, unifiederrors.ValidationErrorf("XML element attribute count exceeded (max %d attributes)", MaxXMLAttributesPerElement)
			}
			elementCount++
			if elementCount > MaxXMLElementCount {
				return nil, unifiederrors.ValidationErrorf("XML element count exceeded (max %d elements)", MaxXMLElementCount)
			}

			newNode, elemName := handleXMLStartElement(t, &nsCtx)

			if len(stack) > 0 {
				addChildToParent(stack[len(stack)-1], elemName, newNode)
			} else {
				result = values.NewObject(1)
				values.SetObjectValue(result, elemName, newNode)
			}
			stack = append(stack, newNode)
			textSizes = append(textSizes, 0)

		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			if len(textSizes) > 0 {
				textSizes = textSizes[:len(textSizes)-1]
			}
			nsCtx.Pop()

		case xml.CharData:
			if str := strings.TrimSpace(string(t)); str != "" && len(stack) > 0 {
				last := len(textSizes) - 1
				if last >= 0 {
					nextSize := textSizes[last] + len(str)
					if textSizes[last] > 0 {
						nextSize++
					}
					if nextSize > MaxXMLTextBytesPerElement {
						return nil, unifiederrors.ValidationErrorf("XML text size exceeded for element (max %d bytes)", MaxXMLTextBytesPerElement)
					}
					textSizes[last] = nextSize
				}
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

	if err := validateXMLOutput(result, renderOpts); err != nil {
		return "", err
	}

	xml := toXMLWithOptions(result, "", renderOpts)
	if renderOpts.writeDeclaration {
		xml = "<?xml version='1.0' encoding='UTF-8'?>\n" + xml
	}
	return xml, nil
}

const (
	xmlNamespaceURI      = "http://www.w3.org/XML/1998/namespace"
	xmlnsNamespaceURI    = "http://www.w3.org/2000/xmlns/"
	defaultNamespaceName = "#default"
)

type xmlOutputNamespaceContext map[string]string

func validateXMLOutput(result interface{}, opts xmlRenderOptions) error {
	root, ok := result.(Object)
	if !ok {
		return unifiederrors.ValidationErrorf(
			"XML output expects an object containing exactly one root element, got %T",
			result,
		)
	}

	rootKeys := extractAndSortChildKeys(root)
	if len(rootKeys) != 1 {
		return unifiederrors.ValidationErrorf(
			"XML output expects exactly one root element, got %d",
			len(rootKeys),
		)
	}
	if len(root) != 1 {
		for _, key := range values.ObjectKeys(root) {
			if key != rootKeys[0] {
				return unifiederrors.ValidationErrorf(
					"XML output root wrapper contains unsupported metadata key %q",
					key,
				)
			}
		}
	}
	if root[rootKeys[0]] == nil && opts.skipNullOnElements {
		return unifiederrors.ValidationError("XML output root element cannot be skipped")
	}

	namespaceCtx := xmlOutputNamespaceContext{"xml": xmlNamespaceURI}
	return validateXMLOutputElement(rootKeys[0], root[rootKeys[0]], opts, namespaceCtx, false)
}

func validateXMLOutputElement(
	sourceName string,
	value interface{},
	opts xmlRenderOptions,
	parentNamespaces xmlOutputNamespaceContext,
	isChild bool,
) error {
	resolvedName, elementPrefix, elementURI, elementHasPrefix := resolveName(sourceName, opts.namespaceVars)
	elementName := normalizeTagName(resolvedName)
	if err := validateXMLQName(elementName); err != nil {
		return unifiederrors.ValidationErrorf("invalid XML element name %q: %v", elementName, err)
	}

	namespaceCtx := copyXMLOutputNamespaceContext(parentNamespaces)
	localNamespaces := make(map[string]struct{})
	valueObject, isObject := value.(Object)
	if isObject {
		if err := addXMLNodeNamespaceDeclarations(namespaceCtx, localNamespaces, valueObject); err != nil {
			return err
		}
	}
	if !isChild {
		if err := addXMLOutputNamespacesIfAbsent(namespaceCtx, localNamespaces, opts.rootDeclaredNames); err != nil {
			return err
		}
	}
	if elementHasPrefix {
		if elementURI == "" && opts.declaredNamespaces != nil {
			elementURI = opts.declaredNamespaces[elementPrefix]
		}
		if elementURI != "" {
			if err := addXMLOutputNamespaceIfAbsent(namespaceCtx, localNamespaces, elementPrefix, elementURI); err != nil {
				return err
			}
		}
	}

	if isObject {
		renderedAttributeNames := make(map[string]struct{})
		for _, key := range extractAndSortAttributeKeys(valueObject) {
			if key == XMLNamespaceKey || opts.skipNullOnAttributes && valueObject[key] == nil {
				continue
			}
			sourceAttrName := key[len(XMLAttrPrefix):]
			resolvedAttrName, attrPrefix, attrURI, attrHasPrefix := resolveName(sourceAttrName, opts.namespaceVars)
			attrName := normalizeTagName(resolvedAttrName)
			if err := validateXMLQName(attrName); err != nil {
				return unifiederrors.ValidationErrorf("invalid XML attribute name %q: %v", attrName, err)
			}
			if _, duplicate := renderedAttributeNames[attrName]; duplicate {
				return unifiederrors.ValidationErrorf(
					"XML element %q contains duplicate resolved attribute name %q",
					elementName,
					attrName,
				)
			}
			renderedAttributeNames[attrName] = struct{}{}
			if attrHasPrefix {
				if attrURI == "" && opts.declaredNamespaces != nil {
					attrURI = opts.declaredNamespaces[attrPrefix]
				}
				if attrURI != "" {
					if err := addXMLOutputNamespaceIfAbsent(namespaceCtx, localNamespaces, attrPrefix, attrURI); err != nil {
						return err
					}
				}
			}
			if err := validateXMLQNamePrefixBinding(attrName, namespaceCtx); err != nil {
				return unifiederrors.ValidationErrorf("invalid XML attribute name %q: %v", attrName, err)
			}
		}
	}

	if err := validateXMLQNamePrefixBinding(elementName, namespaceCtx); err != nil {
		return unifiederrors.ValidationErrorf("invalid XML element name %q: %v", elementName, err)
	}
	if !isObject {
		return nil
	}

	for _, childName := range extractAndSortChildKeys(valueObject) {
		childValue := valueObject[childName]
		switch items := childValue.(type) {
		case XMLMultiValue:
			for _, item := range items {
				if item == nil && opts.skipNullOnElements {
					continue
				}
				if err := validateXMLOutputElement(childName, item, opts, namespaceCtx, true); err != nil {
					return err
				}
			}
		case Array:
			for _, item := range items {
				if item == nil && opts.skipNullOnElements {
					continue
				}
				if err := validateXMLOutputElement(childName, item, opts, namespaceCtx, true); err != nil {
					return err
				}
			}
		default:
			if childValue == nil && opts.skipNullOnElements {
				continue
			}
			if err := validateXMLOutputElement(childName, childValue, opts, namespaceCtx, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyXMLOutputNamespaceContext(input xmlOutputNamespaceContext) xmlOutputNamespaceContext {
	output := make(xmlOutputNamespaceContext, len(input))
	for prefix, uri := range input {
		output[prefix] = uri
	}
	return output
}

func addXMLNodeNamespaceDeclarations(
	namespaceCtx xmlOutputNamespaceContext,
	localNamespaces map[string]struct{},
	node Object,
) error {
	rawNamespaces, exists := node[XMLNamespaceKey]
	if !exists {
		return nil
	}
	namespaces, ok := rawNamespaces.(Object)
	if !ok {
		return unifiederrors.ValidationErrorf(
			"XML namespace declarations must be an object, got %T",
			rawNamespaces,
		)
	}
	for _, sourcePrefix := range values.ObjectKeys(namespaces) {
		prefix := sourcePrefix
		if prefix == defaultNamespaceName {
			prefix = ""
		}
		if err := addXMLOutputNamespace(namespaceCtx, prefix, fmt.Sprintf("%v", namespaces[sourcePrefix])); err != nil {
			return err
		}
		localNamespaces[prefix] = struct{}{}
	}
	return nil
}

func addXMLOutputNamespacesIfAbsent(
	namespaceCtx xmlOutputNamespaceContext,
	localNamespaces map[string]struct{},
	namespaces map[string]string,
) error {
	for prefix, uri := range namespaces {
		if err := addXMLOutputNamespaceIfAbsent(namespaceCtx, localNamespaces, prefix, uri); err != nil {
			return err
		}
	}
	return nil
}

func addXMLOutputNamespaceIfAbsent(
	namespaceCtx xmlOutputNamespaceContext,
	localNamespaces map[string]struct{},
	prefix string,
	uri string,
) error {
	if _, exists := localNamespaces[prefix]; exists {
		return nil
	}
	if err := addXMLOutputNamespace(namespaceCtx, prefix, uri); err != nil {
		return err
	}
	localNamespaces[prefix] = struct{}{}
	return nil
}

func addXMLOutputNamespace(namespaceCtx xmlOutputNamespaceContext, prefix, uri string) error {
	if prefix != "" && !isXMLNCName(prefix) {
		return unifiederrors.ValidationErrorf("invalid XML namespace prefix %q", prefix)
	}
	switch {
	case prefix == "xmlns":
		return unifiederrors.ValidationError("XML namespace prefix \"xmlns\" is reserved")
	case prefix == "xml" && uri != xmlNamespaceURI:
		return unifiederrors.ValidationErrorf("XML namespace prefix \"xml\" must use URI %q", xmlNamespaceURI)
	case prefix != "xml" && uri == xmlNamespaceURI:
		return unifiederrors.ValidationError("the XML namespace URI may only use prefix \"xml\"")
	case uri == xmlnsNamespaceURI:
		return unifiederrors.ValidationError("the xmlns namespace URI cannot be declared")
	case prefix != "" && uri == "":
		return unifiederrors.ValidationErrorf("XML namespace prefix %q cannot use an empty URI", prefix)
	}
	namespaceCtx[prefix] = uri
	return nil
}

func validateXMLQNamePrefixBinding(name string, namespaceCtx xmlOutputNamespaceContext) error {
	prefix, _, hasPrefix := strings.Cut(name, ":")
	if !hasPrefix {
		return nil
	}
	if _, exists := namespaceCtx[prefix]; !exists {
		return fmt.Errorf("namespace prefix %q is not declared", prefix)
	}
	return nil
}

func validateXMLQName(name string) error {
	if !utf8.ValidString(name) {
		return fmt.Errorf("name is not valid UTF-8")
	}
	if strings.Count(name, ":") > 1 {
		return fmt.Errorf("QName must contain at most one colon")
	}
	prefix, local, hasPrefix := strings.Cut(name, ":")
	if !hasPrefix {
		if !isXMLNCName(name) {
			return fmt.Errorf("name must be a valid XML NCName")
		}
		return nil
	}
	if !isXMLNCName(prefix) || !isXMLNCName(local) {
		return fmt.Errorf("QName prefix and local name must be valid XML NCNames")
	}
	if prefix == "xmlns" {
		return fmt.Errorf("prefix %q is reserved", prefix)
	}
	return nil
}

func isXMLNCName(name string) bool {
	if name == "" {
		return false
	}
	for index, r := range name {
		if index == 0 {
			if !isXMLNCNameStartRune(r) {
				return false
			}
			continue
		}
		if !isXMLNCNameRune(r) {
			return false
		}
	}
	return true
}

func isXMLNCNameStartRune(r rune) bool {
	return r == '_' ||
		r >= 'A' && r <= 'Z' ||
		r >= 'a' && r <= 'z' ||
		r >= 0xC0 && r <= 0xD6 ||
		r >= 0xD8 && r <= 0xF6 ||
		r >= 0xF8 && r <= 0x2FF ||
		r >= 0x370 && r <= 0x37D ||
		r >= 0x37F && r <= 0x1FFF ||
		r >= 0x200C && r <= 0x200D ||
		r >= 0x2070 && r <= 0x218F ||
		r >= 0x2C00 && r <= 0x2FEF ||
		r >= 0x3001 && r <= 0xD7FF ||
		r >= 0xF900 && r <= 0xFDCF ||
		r >= 0xFDF0 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0xEFFFF
}

func isXMLNCNameRune(r rune) bool {
	return isXMLNCNameStartRune(r) ||
		r == '-' ||
		r == '.' ||
		r >= '0' && r <= '9' ||
		r == 0xB7 ||
		r >= 0x0300 && r <= 0x036F ||
		r >= 0x203F && r <= 0x2040
}

// Note: The old validateXMLBrackets and helper functions (handleClosingTag, handleOpeningTag,
// handleComment, handleProcessingInstruction) have been replaced with an explicit state machine
// implementation in xml_state_machine.go. The state machine approach provides clearer control
// flow and easier maintenance compared to the nested conditional approach.

// handleXMLStartElement processes a start element, extracting namespaces and attributes.
func handleXMLStartElement(elem xml.StartElement, nsCtx *xmlNamespaceContext) (Object, string) {
	newNode := values.NewObject(len(elem.Attr))

	// Collect namespace declarations from attributes
	nsDecls := namespaceDeclsFromAttrs(elem.Attr)

	// Push namespace context (inherit from parent + new declarations)
	nsCtx.Push(nsDecls)

	// Build element name with prefix if namespace is present
	elemName := nsCtx.ResolveElementName(elem.Name)

	// Store namespace declarations in the node
	if len(nsDecls) > 0 {
		values.SetObjectValue(newNode, XMLNamespaceKey, namespaceDeclsNodeFromAttrs(elem.Attr))
	}

	// Store non-namespace attributes
	for _, attr := range elem.Attr {
		if attr.Name.Space != "xmlns" && !(attr.Name.Local == "xmlns" && attr.Name.Space == "") {
			attrName := XMLAttrPrefix + nsCtx.ResolveElementName(attr.Name)
			values.SetObjectValue(newNode, attrName, attr.Value)
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
	for _, prefix := range values.ObjectKeys(nsMap) {
		uri := nsMap[prefix]
		p := prefix
		if p == "#default" {
			p = ""
		}
		nsDecls[p] = fmt.Sprintf("%v", uri)
	}
	return nsDecls
}

func namespaceDeclsNodeFromAttrs(attrs []xml.Attr) Object {
	node := values.NewObject(len(attrs))
	for _, attr := range attrs {
		switch {
		case attr.Name.Space == "xmlns":
			values.SetObjectValue(node, attr.Name.Local, attr.Value)
		case attr.Name.Local == "xmlns" && attr.Name.Space == "":
			values.SetObjectValue(node, "#default", attr.Value)
		}
	}
	return node
}

func namespaceDeclsFromStringMap(nsMap map[string]string) xmlNamespaceDecls {
	nsDecls := newXMLNamespaceDecls()
	for prefix, uri := range nsMap {
		nsDecls[prefix] = uri
	}
	return nsDecls
}

func (nsDecls xmlNamespaceDecls) toNodeObject() Object {
	node := values.NewObject(len(nsDecls))
	prefixes := make([]string, 0, len(nsDecls))
	for prefix := range nsDecls {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)
	for _, prefix := range prefixes {
		uri := nsDecls[prefix]
		if prefix == "" {
			values.SetObjectValue(node, "#default", uri)
			continue
		}
		values.SetObjectValue(node, prefix, uri)
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
		escapedURI := xmlEscape(uri)
		if prefix == "" {
			decls = append(decls, fmt.Sprintf(`xmlns="%s"`, escapedURI))
			continue
		}
		decls = append(decls, fmt.Sprintf(`xmlns:%s="%s"`, prefix, escapedURI))
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
			values.SetObjectValue(parent, elemName, append(v, child))
		case Array:
			// If it's already a plain Array, it might have come from JSON or elsewhere,
			// but if we're in XML parser, we should probably treat it as MultiValue
			// if it's being added to.
			values.SetObjectValue(parent, elemName, append(XMLMultiValue(v), child))
		default:
			values.SetObjectValue(parent, elemName, XMLMultiValue{existing, child})
		}
	} else {
		values.SetObjectValue(parent, elemName, child)
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
			values.SetObjectValue(node, XMLTextKey, builder.String())
		}
	} else {
		values.SetObjectValue(node, XMLTextKey, text)
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
	for _, key := range values.ObjectKeys(node) {
		values.SetObjectValue(node, key, simplifyXML(node[key]))
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
			for _, k := range values.ObjectKeys(val) {
				if !strings.HasPrefix(k, XMLAttrPrefix) && k != XMLTextKey {
					keys = append(keys, k)
				}
			}
			if len(keys) > 0 {
				k := keys[0]
				return buildXMLElement(k, val[k], opts, false)
			}
			return ""
		}
		return buildElementContent(name, val, opts, isChild)
	default:
		if name == "" {
			return xmlEscape(fmt.Sprintf("%v", v))
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

	return openingTag + xmlEscape(fmt.Sprintf("%v", value))
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
	for _, k := range values.ObjectKeys(val) {
		if k == XMLNamespaceKey || strings.HasPrefix(k, XMLAttrPrefix) {
			keys = append(keys, k)
		}
	}
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
	for _, k := range values.ObjectKeys(val) {
		// Skip special keys: namespaces (@xmlns), attributes (@...), and text (#text)
		if k != XMLNamespaceKey && !strings.HasPrefix(k, XMLAttrPrefix) && k != XMLTextKey {
			keys = append(keys, k)
		}
	}
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
		sb.WriteString(xmlEscape(fmt.Sprintf("%v", text)))
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
	return xmlEscape(fmt.Sprintf("%v", value))
}

// buildXMLOpeningTag builds the opening tag attributes string.
func buildXMLOpeningTag(xmlnsDecls, attrs []string) string {
	allAttrs := append(xmlnsDecls, attrs...)
	if len(allAttrs) > 0 {
		return " " + strings.Join(allAttrs, " ") + ">"
	}
	return ">"
}

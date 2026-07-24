package evaluator

import (
	"fmt"
	"go/token"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
)

// TypeDef represents a custom type definition declared with the 'type' keyword.
// Example: type Currency = String { format: "##.00" }
type TypeDef struct {
	Name       string // The type name (e.g., "Currency")
	BaseType   string // The base type (e.g., "String", "Number")
	Properties Object // Optional properties (e.g., format, locale)
}

// MarshalJSON implements json.Marshaler for TypeDef.
func (t *TypeDef) MarshalJSON() ([]byte, error) {
	repr := fmt.Sprintf("(type: %s = %s)", t.Name, t.BaseType)
	return []byte(fmt.Sprintf("\"%s\"", repr)), nil
}

// getTypeName returns a DataWeave-compatible type name for a value.
func getTypeName(v Value) string {
	switch v.(type) {
	case nil:
		return "Null"
	case bool:
		return "Boolean"
	case int:
		return "Number"
	case float64:
		return "Number"
	case string:
		return "String"
	case Array:
		return "Array"
	case Object:
		return "Object"
	case *Lambda:
		return "Function"
	case *Regex:
		return "Regex"
	case *TypeDef:
		return "Type"
	case Namespace:
		return "Namespace"
	default:
		return "Unknown"
	}
}

// Regex represents a compiled regular expression.
type Regex struct {
	Pattern string
	Flags   string
	Re      *regexp.Regexp
}

// parseOptional strips trailing '?' and returns (base, isOptional).
// Returns an error message if the optional syntax is invalid (e.g., "?").
func parseOptional(typeExpr string) (base string, isOptional bool, errMsg string) {
	trimmed := strings.TrimSpace(typeExpr)
	if !strings.HasSuffix(trimmed, "?") {
		return trimmed, false, ""
	}
	base = strings.TrimSpace(trimmed[:len(trimmed)-1])
	if base == "" {
		return "", true, fmt.Sprintf("invalid optional type: %q", typeExpr)
	}
	return base, true, ""
}

// splitUnion splits "T1 | T2 | T3" into cleaned, trimmed type names.
func splitUnion(typeExpr string) []string {
	parts := strings.Split(typeExpr, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// coerceToType converts a value to the specified type.
func coerceToType(value Value, typeName string, context Context, pos token.Pos) (Value, error) {
	return coerceToTypeWithVisited(value, typeName, context, pos, nil, nil)
}

// coerceToTypeWithVisited handles type coercion with cycle detection.
// Properties are passed down from custom types for formatting/locale support.
// Supports union types (String | Number) and optional types (String?).
// Note: all members of a union share the same property map. Custom types in unions
// merge their TypeDef.Properties with the shared properties (custom properties take precedence).
func coerceToTypeWithVisited(value Value, typeName string, context Context, pos token.Pos, currentPath []string, properties Object) (Value, error) {
	trimmed, isOptional, errMsg := parseOptional(typeName)
	if errMsg != "" {
		return nil, newPosError(errMsg, pos)
	}
	if isOptional {
		return coerceToUnion(value, []string{trimmed, "Null"}, context, pos, currentPath, properties)
	}

	if strings.Contains(trimmed, "|") {
		return coerceToUnion(value, splitUnion(trimmed), context, pos, currentPath, properties)
	}

	switch trimmed {
	case "String":
		return coerceToStringWithProps(value, properties), nil
	case "Number":
		return coerceToNumber(value, pos)
	case "Boolean":
		return coerceToBoolean(value, pos)
	case "Null":
		if value == nil {
			return nil, nil
		}
		return nil, newCoercionTypeError(value, "Null", pos)
	case "Namespace":
		return coerceToNamespace(value, pos)
	}

	// Check for cycles in custom type definitions
	for _, p := range currentPath {
		if p == trimmed {
			return nil, newCyclicTypeError(trimmed, pos)
		}
	}

	if typeDef, ok := context[trimmed]; ok {
		if td, ok := typeDef.(*TypeDef); ok {
			mergedProps := mergeProperties(td.Properties, properties)
			newPath := append(currentPath, trimmed)
			return coerceToTypeWithVisited(value, td.BaseType, context, pos, newPath, mergedProps)
		}
	}

	return nil, newUnknownTypeError(trimmed, pos)
}

// coerceToUnion attempts to coerce a value to one of multiple types in a union.
// It first checks for exact type matches (no coercion needed), then tries coercion.
// Note: all union members share the same property map. For custom types inside the union,
// their own TypeDef.Properties are merged with these shared properties.
func coerceToUnion(value Value, typeNames []string, context Context, pos token.Pos, currentPath []string, properties Object) (Value, error) {
	cleaned := splitUnion(strings.Join(typeNames, "|"))
	if len(cleaned) == 0 {
		return nil, newPosError("empty union type", pos)
	}

	for _, name := range cleaned {
		if matchesTypeExactly(value, name, context) {
			return value, nil
		}
	}

	var firstErr error
	for _, name := range cleaned {
		v, err := coerceToTypeWithVisited(value, name, context, pos, currentPath, properties)
		if err == nil {
			return v, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	msg := fmt.Sprintf("cannot coerce to any of union types: %s", strings.Join(cleaned, " | "))
	if firstErr != nil {
		if unifiedErr, ok := firstErr.(*unifiederrors.Error); ok {
			msg += fmt.Sprintf(" (%s)", unifiedErr.Message)
		}
	}
	return nil, newPosError(msg, pos)
}

// matchesTypeExactly returns true if the value already matches the given type without coercion.
func matchesTypeExactly(value Value, typeName string, context Context) bool {
	base, isOptional, errMsg := parseOptional(typeName)
	if errMsg != "" {
		// Invalid optional syntax - doesn't match
		return false
	}

	// Handle optional (T?) - value matches if it matches T or is nil
	if isOptional {
		if value == nil {
			return true
		}
		return matchesTypeExactly(value, base, context)
	}

	// Handle union (T1 | T2 | ...)
	if strings.Contains(base, "|") {
		parts := splitUnion(base)
		for _, part := range parts {
			if matchesTypeExactly(value, part, context) {
				return true
			}
		}
		return false
	}
	trimmed := base

	switch trimmed {
	case "String":
		_, ok := value.(string)
		return ok
	case "Number":
		switch value.(type) {
		case int, float64:
			return true
		}
		return false
	case "Boolean":
		_, ok := value.(bool)
		return ok
	case "Namespace":
		_, ok := value.(Namespace)
		return ok
	case "Null":
		return value == nil
	case "Array":
		_, ok := value.(Array)
		return ok
	case "Object":
		_, ok := value.(Object)
		return ok
	case "Function":
		_, ok := value.(*Lambda)
		return ok
	case "Regex":
		_, ok := value.(*Regex)
		return ok
	}

	// Check custom types
	if typeDef, ok := context[trimmed]; ok {
		if td, ok := typeDef.(*TypeDef); ok {
			return matchesTypeExactly(value, td.BaseType, context)
		}
	}

	return false
}

func coerceToNamespace(value Value, pos token.Pos) (Value, error) {
	switch v := value.(type) {
	case Namespace:
		return v, nil
	case Object:
		uriVal, ok := v["uri"]
		if !ok {
			return nil, newPosError("Namespace requires a uri field", pos)
		}
		uri, ok := uriVal.(string)
		if !ok || strings.TrimSpace(uri) == "" {
			return nil, newPosError("Namespace uri must be a non-empty string", pos)
		}
		prefix := ""
		if prefixVal, exists := v["prefix"]; exists && prefixVal != nil {
			prefixStr, ok := prefixVal.(string)
			if !ok {
				return nil, newPosError("Namespace prefix must be a string", pos)
			}
			prefix = prefixStr
		}
		return Namespace{Prefix: prefix, URI: uri}, nil
	default:
		return nil, newCoercionTypeError(value, "Namespace", pos)
	}
}

// mergeProperties merges type properties, with child properties overriding parent.
func mergeProperties(parent, child Object) Object {
	if parent == nil && child == nil {
		return nil
	}

	return values.MergeObjects(parent, child)
}

// coerceToString converts any value to a string.
func coerceToString(value Value) string {
	if value == nil {
		return "null"
	}
	if _, ok := value.(*Lambda); ok {
		return "<function>"
	}
	return fmt.Sprintf("%v", value)
}

// coerceToStringWithProps converts a value to a string, applying format properties.
func coerceToStringWithProps(value Value, properties Object) string {
	if properties == nil {
		return coerceToString(value)
	}

	if format, ok := properties["format"].(string); ok {
		return applyStringFormat(value, format)
	}

	return coerceToString(value)
}

// applyStringFormat applies a format pattern to a value.
// Supports DataWeave-style patterns like "##.00" for numbers.
func applyStringFormat(value Value, format string) string {
	switch v := value.(type) {
	case int:
		return formatNumber(float64(v), format)
	case float64:
		return formatNumber(v, format)
	case string:
		// Try to parse as date/time and reformat
		if t, err := parseDateOrTime(v); err == nil {
			goLayout := convertJavaDatePatternToGo(format)
			return t.Format(goLayout)
		}
		return v
	default:
		return coerceToString(value)
	}
}

// parseDateOrTime attempts to parse a string into a time.Time using common layouts.
func parseDateOrTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, unifiederrors.EvalErrorf("cannot parse date/time: %s", s)
}

// convertJavaDatePatternToGo converts a Java SimpleDateFormat pattern to a Go time layout.
// Supported Java tokens:
//
//	yyyy, yy
//	MMMM, MMM, MM, M
//	dd, d
//	EEEE, EEE, EE, E
//	HH, H, hh, h
//	mm, m, ss, s
//	SSS, SS, S
//	a
//	z, Z, XXX, XX, X
//
// Notes:
//   - Single quotes escape literals; doubled quotes ” emit a single quote.
//   - Unsupported tokens are passed through unchanged.
//   - This is a pragmatic subset for DataWeave-style formats, not a full SimpleDateFormat implementation.
func convertJavaDatePatternToGo(pattern string) string {
	// Common replacements sorted by length (longest first) to prevent partial matches
	replacements := []struct {
		java string
		goOp string
	}{
		{"MMMM", "January"},
		{"EEEE", "Monday"},
		{"SSS", "000"},
		{"XXX", "Z07:00"},
		{"yyyy", "2006"},
		{"MMM", "Jan"},
		{"MM", "01"},
		{"EEE", "Mon"},
		{"EE", "Mon"},
		{"HH", "15"},
		{"hh", "03"},
		{"mm", "04"},
		{"ss", "05"},
		{"SS", "00"},
		{"yy", "06"},
		{"XX", "Z0700"},
		{"M", "1"},
		{"dd", "02"},
		{"d", "2"},
		{"H", "15"},
		{"h", "3"},
		{"m", "4"},
		{"s", "5"},
		{"S", "0"},
		{"E", "Mon"},
		{"a", "PM"},
		{"z", "MST"},
		{"Z", "-0700"},
		{"X", "Z07"},
	}

	var result strings.Builder
	runes := []rune(pattern)
	i := 0
	inQuote := false

	for i < len(runes) {
		// Handle single quotes (literals)
		if runes[i] == '\'' {
			if i+1 < len(runes) && runes[i+1] == '\'' {
				// Double single quote '' means one single quote literal
				result.WriteRune('\'')
				i += 2
				continue
			}
			// Toggle quote state
			inQuote = !inQuote
			i++
			continue
		}

		if inQuote {
			result.WriteRune(runes[i])
			i++
			continue
		}

		matched := false
		for _, rep := range replacements {
			if matchesRunesAt(runes, i, rep.java) {
				result.WriteString(rep.goOp)
				i += len(rep.java)
				matched = true
				break
			}
		}
		if !matched {
			result.WriteRune(runes[i])
			i++
		}
	}

	return result.String()
}

func matchesRunesAt(runes []rune, index int, pattern string) bool {
	patRunes := []rune(pattern)
	if index+len(patRunes) > len(runes) {
		return false
	}
	for j := 0; j < len(patRunes); j++ {
		if runes[index+j] != patRunes[j] {
			return false
		}
	}
	return true
}

// formatNumber formats a number using a pattern like "##.00" or "###,###.00".
func formatNumber(num float64, pattern string) string {
	decimalIdx := strings.LastIndex(pattern, ".")
	var decimals int
	var useThousandsSep bool

	if decimalIdx >= 0 {
		decimals = len(pattern) - decimalIdx - 1
	}

	if strings.Contains(pattern, ",") {
		useThousandsSep = true
	}

	if decimals > 0 {
		num = roundToDecimals(num, decimals)
	}

	intPart := int64(num)
	fracPart := num - float64(intPart)
	if fracPart < 0 {
		fracPart = -fracPart
	}

	var result strings.Builder

	if num < 0 {
		result.WriteString("-")
		intPart = -intPart
	}

	intStr := strconv.FormatInt(intPart, 10)

	if useThousandsSep {
		for i, ch := range intStr {
			if i > 0 && (len(intStr)-i)%3 == 0 {
				result.WriteRune(',')
			}
			result.WriteRune(ch)
		}
	} else {
		result.WriteString(intStr)
	}

	if decimals > 0 {
		result.WriteString(".")
		fracStr := strconv.FormatFloat(fracPart, 'f', decimals, 64)
		if len(fracStr) > 2 {
			result.WriteString(fracStr[2:])
		} else {
			for i := 0; i < decimals; i++ {
				result.WriteString("0")
			}
		}
	}

	return result.String()
}

// roundToDecimals rounds a float to the specified number of decimal places.
func roundToDecimals(num float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))
	return math.Round(num*shift) / shift
}

// coerceToNumber converts a value to a number (int or float64).
func coerceToNumber(value Value, pos token.Pos) (Value, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return v, nil
	case string:
		if num, ok := values.ParseNumericLiteral(v); ok {
			return num, nil
		}
		return nil, newCoercionError(v, "Number", pos)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return nil, newCoercionTypeError(value, "Number", pos)
	}
}

// coerceToBoolean converts a value to a boolean.
func coerceToBoolean(value Value, pos token.Pos) (Value, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		if b, ok := values.ParseBoolLiteral(v); ok {
			return b, nil
		}
		return nil, newCoercionError(v, "Boolean", pos)
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case nil:
		return false, nil
	default:
		return nil, newCoercionTypeError(value, "Boolean", pos)
	}
}

// Numeric coercion helpers for arithmetic operations

// ToFloat coerces a value to float64 if it's numeric (int or float64).
func ToFloat(v Value) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	default:
		return 0, false
	}
}

// BothInts returns true if both values are ints, along with their values.
func BothInts(left, right Value) (l, r int, ok bool) {
	l, okL := left.(int)
	r, okR := right.(int)
	return l, r, okL && okR
}

// BothFloats coerces both values to float64 if they are numeric.
func BothFloats(left, right Value) (l, r float64, ok bool) {
	l, okL := ToFloat(left)
	r, okR := ToFloat(right)
	return l, r, okL && okR
}

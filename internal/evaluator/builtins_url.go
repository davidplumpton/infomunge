package evaluator

import (
	"fmt"
	"go/ast"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"infomunge/pkg/values"
)

// callBuiltinParseURI implements parseURI(uri).
func callBuiltinParseURI(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("parseURI requires exactly 1 argument: uri", e.Pos())
	}

	rawURI, err := assertStringArg(args[0], 1, "parseURI", e)
	if err != nil {
		return nil, err
	}

	parsed, err := url.Parse(rawURI)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("parseURI expects a valid URI: %v", err), e.Pos())
	}

	query := values.NewObject(0)
	result := values.NewObject(8)
	values.SetObjectValue(result, "scheme", parsed.Scheme)
	values.SetObjectValue(result, "host", parsed.Hostname())
	values.SetObjectValue(result, "path", parsed.Path)
	values.SetObjectValue(result, "fragment", parsed.Fragment)
	values.SetObjectValue(result, "query", query)
	values.SetObjectValue(result, "user", nil)
	values.SetObjectValue(result, "password", nil)
	values.SetObjectValue(result, "port", nil)

	if parsed.User != nil {
		values.SetObjectValue(result, "user", parsed.User.Username())
		if pw, hasPw := parsed.User.Password(); hasPw {
			values.SetObjectValue(result, "password", pw)
		}
	}

	if parsed.Port() != "" {
		port, convErr := strconv.Atoi(parsed.Port())
		if convErr != nil {
			return nil, newPosError(fmt.Sprintf("parseURI could not parse URI port %q", parsed.Port()), e.Pos())
		}
		values.SetObjectValue(result, "port", port)
	}

	queryValues, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("parseURI failed to parse query: %v", err), e.Pos())
	}

	queryKeys := make([]string, 0, len(queryValues))
	for key := range queryValues {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)
	for _, key := range queryKeys {
		queryValuesForKey := queryValues[key]
		if len(queryValuesForKey) == 1 {
			values.SetObjectValue(query, key, queryValuesForKey[0])
			continue
		}
		items := make(Array, len(queryValuesForKey))
		for i, value := range queryValuesForKey {
			items[i] = value
		}
		values.SetObjectValue(query, key, items)
	}

	return result, nil
}

// callBuiltinCompose implements compose(parts).
func callBuiltinCompose(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("compose requires exactly 1 argument: parts", e.Pos())
	}

	parts, ok := args[0].(Object)
	if !ok {
		return nil, newPosError(fmt.Sprintf("compose expects an object, got %T", args[0]), e.Pos())
	}

	scheme, err := optionalStringField(parts, "scheme", "compose", e)
	if err != nil {
		return nil, err
	}
	host, err := optionalStringField(parts, "host", "compose", e)
	if err != nil {
		return nil, err
	}
	path, err := optionalStringField(parts, "path", "compose", e)
	if err != nil {
		return nil, err
	}
	fragment, err := optionalStringField(parts, "fragment", "compose", e)
	if err != nil {
		return nil, err
	}

	port, err := optionalPortField(parts, "port", "compose", e)
	if err != nil {
		return nil, err
	}
	if port != "" && host == "" {
		return nil, newPosError("compose expects host when port is provided", e.Pos())
	}

	query, err := composeQuery(parts, e)
	if err != nil {
		return nil, err
	}

	composed := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawQuery: query,
		Fragment: fragment,
	}

	if port != "" {
		composed.Host = net.JoinHostPort(host, port)
	}

	user, err := optionalStringField(parts, "user", "compose", e)
	if err != nil {
		return nil, err
	}
	password, err := optionalStringField(parts, "password", "compose", e)
	if err != nil {
		return nil, err
	}

	if user != "" {
		if password != "" {
			composed.User = url.UserPassword(user, password)
		} else {
			composed.User = url.User(user)
		}
	} else if password != "" {
		return nil, newPosError("compose expects user when password is provided", e.Pos())
	}

	return composed.String(), nil
}

// callBuiltinEncodeURI implements encodeURI(uri).
func callBuiltinEncodeURI(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("encodeURI requires exactly 1 argument: uri", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "encodeURI", e)
	if err != nil {
		return nil, err
	}

	const allowed = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~:/?#[]@!$&'()*+,;="
	return percentEncode(text, allowed), nil
}

// callBuiltinDecodeURI implements decodeURI(uri).
func callBuiltinDecodeURI(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("decodeURI requires exactly 1 argument: uri", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "decodeURI", e)
	if err != nil {
		return nil, err
	}

	decoded, err := url.PathUnescape(text)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("decodeURI expects a valid encoded URI: %v", err), e.Pos())
	}
	return decoded, nil
}

// callBuiltinEncodeURIComponent implements encodeURIComponent(component).
func callBuiltinEncodeURIComponent(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("encodeURIComponent requires exactly 1 argument: component", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "encodeURIComponent", e)
	if err != nil {
		return nil, err
	}

	return url.PathEscape(text), nil
}

// callBuiltinDecodeURIComponent implements decodeURIComponent(component).
func callBuiltinDecodeURIComponent(args []Value, e *ast.CallExpr) (Value, error) {
	if len(args) != 1 {
		return nil, newPosError("decodeURIComponent requires exactly 1 argument: component", e.Pos())
	}

	text, err := assertStringArg(args[0], 1, "decodeURIComponent", e)
	if err != nil {
		return nil, err
	}

	decoded, err := url.PathUnescape(text)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("decodeURIComponent expects a valid encoded component: %v", err), e.Pos())
	}
	return decoded, nil
}

func optionalStringField(parts Object, key, funcName string, e *ast.CallExpr) (string, error) {
	val, ok := parts[key]
	if !ok || val == nil {
		return "", nil
	}
	str, ok := val.(string)
	if !ok {
		return "", newPosError(fmt.Sprintf("%s expects %q to be a string, got %T", funcName, key, val), e.Pos())
	}
	return str, nil
}

func optionalPortField(parts Object, key, funcName string, e *ast.CallExpr) (string, error) {
	val, ok := parts[key]
	if !ok || val == nil {
		return "", nil
	}

	switch v := val.(type) {
	case int:
		if v < 0 || v > 65535 {
			return "", newPosError(fmt.Sprintf("%s expects %q to be between 0 and 65535", funcName, key), e.Pos())
		}
		return strconv.Itoa(v), nil
	case string:
		if strings.TrimSpace(v) == "" {
			return "", nil
		}
		port, err := strconv.Atoi(v)
		if err != nil || port < 0 || port > 65535 {
			return "", newPosError(fmt.Sprintf("%s expects %q string to be a valid port number", funcName, key), e.Pos())
		}
		return v, nil
	default:
		return "", newPosError(fmt.Sprintf("%s expects %q to be an integer or string, got %T", funcName, key, val), e.Pos())
	}
}

func composeQuery(parts Object, e *ast.CallExpr) (string, error) {
	val, ok := parts["query"]
	if !ok || val == nil {
		return "", nil
	}

	if raw, ok := val.(string); ok {
		return strings.TrimPrefix(raw, "?"), nil
	}

	queryObj, ok := val.(Object)
	if !ok {
		return "", newPosError(fmt.Sprintf("compose expects \"query\" to be a string or object, got %T", val), e.Pos())
	}

	values := url.Values{}
	for key, raw := range queryObj {
		switch v := raw.(type) {
		case Array:
			for _, item := range v {
				values.Add(key, coerceToString(item))
			}
		default:
			values.Add(key, coerceToString(v))
		}
	}

	return values.Encode(), nil
}

func percentEncode(input, allowed string) string {
	allowedSet := map[byte]struct{}{}
	for i := 0; i < len(allowed); i++ {
		allowedSet[allowed[i]] = struct{}{}
	}

	var b strings.Builder
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if _, ok := allowedSet[ch]; ok {
			b.WriteByte(ch)
			continue
		}
		b.WriteString(fmt.Sprintf("%%%02X", ch))
	}
	return b.String()
}

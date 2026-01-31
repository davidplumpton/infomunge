package evaluator

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/preprocessor"
)

// callBuiltinUpdateExpr implements the __updateExpr(value, casesString) function.
func callBuiltinUpdateExpr(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("update expression requires exactly 2 arguments: value and cases", e.Pos())
	}

	// First argument is evaluated (the value to update)
	value, err := evalASTWithDepth(e.Args[0], context, depth+1)
	if err != nil {
		return nil, err
	}

	// Second argument must be a string literal containing case statements
	casesExpr, ok := e.Args[1].(*ast.BasicLit)
	if !ok || casesExpr.Kind != token.STRING {
		return nil, newPosError("update expression cases must be a string literal", e.Args[1].Pos())
	}

	casesStr, err := strconv.Unquote(casesExpr.Value)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("invalid cases string: %s", err), e.Args[1].Pos())
	}

	// Parse and apply the case statements
	res, err := applyUpdateCases(value, casesStr, context, depth)
	if err != nil {
		return nil, newPosError(err.Error(), e.Args[1].Pos())
	}
	return res, nil
}

// applyUpdateCases parses and applies update case statements to a value
func applyUpdateCases(value interface{}, casesStr string, context map[string]interface{}, depth int) (interface{}, error) {
	// Parse case statements line by line
	lines := strings.Split(casesStr, "\n")
	result := deepCopy(value) // Start with a copy of the value

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Parse: case varName at .selector -> expression
		if !strings.HasPrefix(trimmed, "case ") {
			return nil, unifiederrors.EvalErrorf("update case must start with 'case': %s", line)
		}

		rest := strings.TrimPrefix(trimmed, "case ")

		// Find " at " separator
		atIdx := strings.Index(rest, " at ")
		if atIdx == -1 {
			return nil, unifiederrors.EvalErrorf("update case missing 'at' keyword: %s", line)
		}

		varName := strings.TrimSpace(rest[:atIdx])
		afterAt := rest[atIdx+4:]

		// Find " -> " separator
		arrowIdx := strings.Index(afterAt, " -> ")
		if arrowIdx == -1 {
			return nil, unifiederrors.EvalErrorf("update case missing '->': %s", line)
		}

		selector := strings.TrimSpace(afterAt[:arrowIdx])
		expression := strings.TrimSpace(afterAt[arrowIdx+4:])

		// Apply this case
		var applyErr error
		result, applyErr = applyUpdateCase(result, varName, selector, expression, context, depth)
		if applyErr != nil {
			return nil, applyErr
		}
	}

	return result, nil
}

// applyUpdateCase applies a single update case to a value
func applyUpdateCase(value interface{}, varName, selector, expression string, context map[string]interface{}, depth int) (interface{}, error) {
	// Parse the selector path
	path, err := parseSelectorPath(selector)
	if err != nil {
		return nil, err
	}

	if len(path) == 0 {
		return nil, unifiederrors.EvalError("empty selector path")
	}

	// Get the current value at the selector path
	currentValue, err := getValueAtPath(value, path)
	if err != nil {
		// Path doesn't exist, return original value unchanged
		return value, nil
	}

	// Create local context with the variable binding
	localContext := copyContext(context)
	localContext[varName] = currentValue

	// Evaluate the expression in the local context
	preparedExpr, _, err := preprocessor.PrepareForParsing(expression, preprocessor.Options{})
	if err != nil {
		return nil, unifiederrors.EvalErrorf("preprocessing error in update case expression: %s", err)
	}
	parsedExpr, err := goparser.ParseExpr(preparedExpr)
	if err != nil {
		return nil, unifiederrors.EvalErrorf("parse error in update case expression: %s", err)
	}

	newValue, err := evalASTWithDepth(parsedExpr, localContext, depth+1)
	if err != nil {
		return nil, err
	}

	// Set the new value at the path
	return setValueAtPath(value, path, newValue, 0)
}

// selectorSegment represents a part of a selector path
type selectorSegment struct {
	fieldName string // for .field selectors
	index     int    // for [index] selectors (-1 if not an index)
	isIndex   bool
}

// parseSelectorPath parses a selector like ".age" or ".address.street" or ".items[0]"
func parseSelectorPath(selector string) ([]selectorSegment, error) {
	var path []selectorSegment
	s := strings.TrimSpace(selector)

	i := 0
	for i < len(s) {
		if s[i] == '.' {
			i++ // skip the dot
			// Read field name
			start := i
			for i < len(s) && s[i] != '.' && s[i] != '[' {
				i++
			}
			if start == i {
				return nil, unifiederrors.EvalErrorf("empty field name in selector: %s", selector)
			}
			path = append(path, selectorSegment{fieldName: s[start:i], index: -1, isIndex: false})
		} else if s[i] == '[' {
			i++ // skip the [
			// Read index
			start := i
			for i < len(s) && s[i] != ']' {
				i++
			}
			if i >= len(s) {
				return nil, unifiederrors.EvalErrorf("unclosed bracket in selector: %s", selector)
			}
			indexStr := s[start:i]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, unifiederrors.EvalErrorf("invalid index in selector: %s", indexStr)
			}
			path = append(path, selectorSegment{index: index, isIndex: true})
			i++ // skip the ]
		} else {
			return nil, unifiederrors.EvalErrorf("unexpected character in selector: %c at position %d", s[i], i)
		}
	}

	return path, nil
}

// getValueAtPath retrieves a value at the given selector path
func getValueAtPath(value interface{}, path []selectorSegment) (interface{}, error) {
	current := value
	for _, seg := range path {
		if seg.isIndex {
			arr, ok := current.([]interface{})
			if !ok {
				return nil, unifiederrors.EvalError("expected array for index selector")
			}
			if seg.index < 0 || seg.index >= len(arr) {
				return nil, unifiederrors.EvalErrorf("index out of bounds: %d", seg.index)
			}
			current = arr[seg.index]
		} else {
			obj, ok := current.(map[string]interface{})
			if !ok {
				return nil, unifiederrors.EvalError("expected object for field selector")
			}
			val, exists := obj[seg.fieldName]
			if !exists {
				return nil, unifiederrors.EvalErrorf("field not found: %s", seg.fieldName)
			}
			current = val
		}
	}
	return current, nil
}

// setValueAtPath sets a value at the given selector path, returning a new value with the update applied
// updateArrayElement updates an element in an array at the given index.
func updateArrayElement(arr []interface{}, index int, newValue interface{}) ([]interface{}, error) {
	if index < 0 || index >= len(arr) {
		return nil, unifiederrors.EvalErrorf("index out of bounds: %d", index)
	}
	newArr := make([]interface{}, len(arr))
	copy(newArr, arr)
	newArr[index] = newValue
	return newArr, nil
}

// updateObjectField updates a field in an object.
func updateObjectField(obj map[string]interface{}, fieldName string, newValue interface{}) map[string]interface{} {
	newObj := make(map[string]interface{})
	for k, v := range obj {
		newObj[k] = v
	}
	newObj[fieldName] = newValue
	return newObj
}

func setValueAtPath(value interface{}, path []selectorSegment, newValue interface{}, depth int) (interface{}, error) {
	if depth > MaxEvalDepth {
		return nil, unifiederrors.EvalErrorf("update path depth limit exceeded (max %d)", MaxEvalDepth)
	}

	if len(path) == 0 {
		return newValue, nil
	}

	seg := path[0]
	isTerminal := len(path) == 1

	if seg.isIndex {
		arr, ok := value.([]interface{})
		if !ok {
			return nil, unifiederrors.EvalError("expected array for index selector")
		}
		if isTerminal {
			return updateArrayElement(arr, seg.index, newValue)
		}
		childValue, err := setValueAtPath(arr[seg.index], path[1:], newValue, depth+1)
		if err != nil {
			return nil, err
		}
		return updateArrayElement(arr, seg.index, childValue)
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil, unifiederrors.EvalError("expected object for field selector")
	}
	if isTerminal {
		return updateObjectField(obj, seg.fieldName, newValue), nil
	}
	childVal, exists := obj[seg.fieldName]
	if !exists {
		return nil, unifiederrors.EvalErrorf("field not found: %s", seg.fieldName)
	}
	childValue, err := setValueAtPath(childVal, path[1:], newValue, depth+1)
	if err != nil {
		return nil, err
	}
	return updateObjectField(obj, seg.fieldName, childValue), nil
}

// deepCopy creates a deep copy of a value
func deepCopy(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for k, val := range v {
			newMap[k] = deepCopy(val)
		}
		return newMap
	case []interface{}:
		newArr := make([]interface{}, len(v))
		for i, val := range v {
			newArr[i] = deepCopy(val)
		}
		return newArr
	default:
		return value
	}
}

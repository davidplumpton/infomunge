package evaluator

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"sort"
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
)

// ControlFlowSignal is used to signal break or continue in loops
type ControlFlowSignal struct {
	Type string // "break" or "continue"
}

// Lambda represents a lambda function with its parameter definitions and body.
type Lambda struct {
	Params  []ParamDef // Parameter definitions with names, types, and defaults
	Body    string     // The body expression as a string (unevaluated)
	BodyAST ast.Expr   // The body expression as an AST (for evaluation with parameters)
	Env     Context    // Optional: module or closure environment for lexical scoping
}

// ParamNames returns the names of all parameters in order.
func (l *Lambda) ParamNames() []string {
	names := make([]string, len(l.Params))
	for i, p := range l.Params {
		names[i] = p.Name
	}
	return names
}

// ParamName returns the name of the parameter at the given index.
func (l *Lambda) ParamName(idx int) string {
	if idx < 0 || idx >= len(l.Params) {
		return ""
	}
	return l.Params[idx].Name
}

// ParamCount returns the number of parameters.
func (l *Lambda) ParamCount() int {
	return len(l.Params)
}

// GetDefault returns the default value for a parameter by name.
// Returns the value and true if a default exists, nil and false otherwise.
func (l *Lambda) GetDefault(name string) (Value, bool) {
	for _, p := range l.Params {
		if p.Name == name && p.HasDefault {
			return p.DefaultValue, true
		}
	}
	return nil, false
}

// HasDefaults returns true if any parameter has a default value.
func (l *Lambda) HasDefaults() bool {
	for _, p := range l.Params {
		if p.HasDefault {
			return true
		}
	}
	return false
}

// MarshalJSON implements json.Marshaler for Lambda.
func (l *Lambda) MarshalJSON() ([]byte, error) {
	// Represent lambda as a string in JSON output for display purposes
	paramList := strings.Join(l.ParamNames(), ", ")
	repr := fmt.Sprintf("(lambda: [%s] => %s)", paramList, l.Body)
	return []byte(fmt.Sprintf("\"%s\"", repr)), nil
}

// ErrorContext holds information needed for error reporting and location mapping
type ErrorContext struct {
	ExprStr    string // The original expression string (transformed)
	Mapping    []int  // Mapping from transformed indices back to original
	BodyOffset int    // Offset of the body in the full raw input
	FullRaw    string // The full raw input for line/column calculation
}

// posError represents an error with position information
type posError struct {
	msg string
	pos token.Pos
}

func (e *posError) Error() string {
	return e.msg
}

func (e *posError) Pos() token.Pos {
	return e.pos
}

// newPosError creates a new positional error
func newPosError(msg string, pos token.Pos) error {
	return unifiederrors.EvalPositional(msg, pos, token.NewFileSet())
}

// Evaluate parses and evaluates the expression.
func Evaluate(exprStr string, context Context, mapping []int, bodyOffset int, fullRaw string) (Value, error) {
	ctx := &ErrorContext{
		ExprStr:    exprStr,
		Mapping:    mapping,
		BodyOffset: bodyOffset,
		FullRaw:    fullRaw,
	}
	return EvaluateWithContext(exprStr, context, ctx)
}

// EvaluateWithContext parses and evaluates the expression with error context.
func EvaluateWithContext(exprStr string, context Context, errCtx *ErrorContext) (Value, error) {
	expr, err := parser.ParseExpr(exprStr)
	if err != nil {
		return nil, errCtx.FormatParseError(err, exprStr)
	}

	result, err := evalASTWithDepth(expr, context, 0)
	if err != nil {
		return nil, errCtx.FormatEvalError(err)
	}
	return result, nil
}

// FormatEvalError converts a unified error to a user-friendly error with line/column information
func (ec *ErrorContext) FormatEvalError(err error) error {
	// Check if it's a unified error
	if unifiedErr, ok := err.(*unifiederrors.Error); ok && unifiedErr.Position != nil {
		// Use the existing position information
		return fmt.Errorf("%s:%d:%d: %s", unifiedErr.Position.Filename, unifiedErr.Position.Line, unifiedErr.Position.Column, unifiedErr.Message)
	}

	// Fallback for legacy posError (shouldn't occur with migration)
	var pe interface{ Pos() token.Pos }
	if errors.As(err, &pe) {
		offset := int(pe.Pos()) - 1
		if offset >= 0 && offset < len(ec.Mapping) {
			originalOffset := ec.Mapping[offset] + ec.BodyOffset
			origLine, origCol := offsetToLineCol(ec.FullRaw, originalOffset)
			return fmt.Errorf("%d:%d: %s", origLine, origCol, err.Error())
		}
	}

	return err
}

// FormatParseError converts a parser error to a user-friendly error with line/column information
func (ec *ErrorContext) FormatParseError(err error, exprStr string) error {
	if len(ec.Mapping) == 0 {
		return err
	}

	// Try to extract position from error message (Go parser errors: "1:12: expected ...")
	errMsg := err.Error()
	var line, col int
	if _, scanErr := fmt.Sscanf(errMsg, "%d:%d:", &line, &col); scanErr != nil {
		return err
	}

	// Calculate offset in transformed string (line 1 based)
	offset := 0
	lines := strings.Split(exprStr, "\n")
	for i := 0; i < line-1; i++ {
		offset += len(lines[i]) + 1
	}
	offset += col - 1

	if offset >= len(ec.Mapping) {
		offset = len(ec.Mapping) - 1
	}

	originalOffset := ec.Mapping[offset] + ec.BodyOffset
	origLine, origCol := offsetToLineCol(ec.FullRaw, originalOffset)

	return fmt.Errorf("%d:%d: %s", origLine, origCol, errMsg[strings.Index(errMsg, ": ")+2:])
}

// offsetToLineCol converts a byte offset in a string to 1-based line and column numbers.
func offsetToLineCol(s string, offset int) (line, col int) {
	line, col = 1, 1
	for i := 0; i < offset && i < len(s); i++ {
		if s[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// evalAST evaluates an expression with initial depth of 0 using the visitor pattern.
func evalAST(expr ast.Expr, context Context) (Value, error) {
	visitor := NewDefaultVisitor(context, 0)
	return visitor.Visit(expr)
}

// evalASTWithDepth evaluates an expression using the visitor pattern with the given depth.
func evalASTWithDepth(expr ast.Expr, context Context, depth int) (Value, error) {
	visitor := NewDefaultVisitor(context, depth)
	return visitor.Visit(expr)
}

// copyContext creates a shallow copy of a context map for scope isolation.
//
// This function is essential for maintaining proper variable scoping during:
//   - Lambda evaluation (map, filter, reduce, etc.)
//   - Function calls
//   - Pattern matching (case/match expressions)
//   - Local variable bindings (do blocks, using expressions)
//
// The shallow copy ensures that:
//   - New bindings (like lambda parameters) don't pollute the parent scope
//   - Each iteration in collection operations gets isolated parameter bindings
//   - Parent scope variables remain accessible in the child scope
//
// Important: This is intentionally a SHALLOW copy. The map keys are copied,
// but the Values themselves are shared references. This means:
//   - Reading parent variables works correctly (they're accessible)
//   - Mutating a referenced object (array/map) affects the original
//   - Reassigning a variable in child scope doesn't affect parent scope
//
// Naming conventions used across the codebase:
//   - lambdaContext: for lambda/closure evaluation
//   - fnContext: for user-defined function calls
//   - matchContext: for pattern matching expressions
//   - localContext: for local variable scopes (do blocks)
//
// Example usage:
//
//	lambdaContext := copyContext(context)
//	lambdaContext[lambda.ParamName(0)] = element  // add lambda parameter
//	result, err := evalASTWithDepth(lambda.BodyAST, lambdaContext, depth+1)
func copyContext(context Context) Context {
	newCtx := make(Context, len(context))
	for k, v := range context {
		newCtx[k] = v
	}
	return newCtx
}

// Core evaluation functions needed by visitor pattern

// evalBasicLit evaluates basic literals (strings, numbers, etc.)
func evalBasicLit(e *ast.BasicLit) (Value, error) {
	switch e.Kind {
	case token.STRING:
		v, err := strconv.Unquote(e.Value)
		if err != nil {
			return nil, newPosError(fmt.Sprintf("invalid string literal: %s", err), e.Pos())
		}
		return v, nil
	case token.INT:
		v, err := strconv.Atoi(e.Value)
		if err != nil {
			return nil, newPosError(fmt.Sprintf("invalid integer literal: %s", err), e.Pos())
		}
		return v, nil
	case token.FLOAT:
		v, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			return nil, newPosError(fmt.Sprintf("invalid float literal: %s", err), e.Pos())
		}
		return v, nil
	default:
		return e.Value, nil
	}
}

// evalIdent evaluates identifiers (variables, constants)
func evalIdent(e *ast.Ident, context map[string]interface{}) (interface{}, error) {
	switch e.Name {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "nil", "null":
		return nil, nil
	}

	if val, ok := context[e.Name]; ok {
		return val, nil
	}
	return e.Name, nil
}

// Binary operations

// add performs addition with type coercion and concatenation support
func add(left, right Value) (Value, error) {
	// Special case: string concatenation (supports string + any and any + string)
	if l, okL := left.(string); okL {
		if r, okR := right.(string); okR {
			return l + r, nil
		}
		// String + non-string: convert right to string
		return l + fmt.Sprintf("%v", right), nil
	}
	if r, okR := right.(string); okR {
		// Non-string + string: convert left to string
		return fmt.Sprintf("%v", left) + r, nil
	}

	// Special case: array concatenation (supports array + array)
	if l, okL := left.([]interface{}); okL {
		if r, okR := right.([]interface{}); okR {
			result := make([]interface{}, 0, len(l)+len(r))
			result = append(result, l...)
			result = append(result, r...)
			return result, nil
		}
	}

	return numericOp(left, right,
		func(l, r int) int { return l + r },
		func(l, r float64) float64 { return l + r },
		"addition", "add")
}

// sub performs subtraction or object key removal
func sub(left, right Value) (Value, error) {
	// Special case: object key removal (object - "key" → object without that key)
	if obj, ok := left.(map[string]interface{}); ok {
		if key, ok := right.(string); ok {
			result := make(map[string]interface{}, len(obj))
			for k, v := range obj {
				if k != key {
					result[k] = v
				}
			}
			return result, nil
		}
	}

	return numericOp(left, right,
		func(l, r int) int { return l - r },
		func(l, r float64) float64 { return l - r },
		"subtraction", "subtract")
}

// mul performs multiplication
func mul(left, right Value) (Value, error) {
	return numericOp(left, right,
		func(l, r int) int { return l * r },
		func(l, r float64) float64 { return l * r },
		"multiplication", "multiply")
}

// quo performs division with zero-check
func quo(left, right Value) (Value, error) {
	// Check for division by zero before calling numericOp
	if r, ok := right.(int); ok && r == 0 {
		return nil, unifiederrors.EvalError("division by zero")
	}
	if r, ok := right.(float64); ok && r == 0 {
		return nil, unifiederrors.EvalError("division by zero")
	}
	return numericOp(left, right,
		func(l, r int) int { return l / r },
		func(l, r float64) float64 { return l / r },
		"division", "divide")
}

// land performs logical AND operation
func land(left, right Value) (Value, error) {
	l, okL := left.(bool)
	r, okR := right.(bool)
	if okL && okR {
		return l && r, nil
	}
	return nil, unifiederrors.EvalErrorf("logical AND requires booleans, got %T and %T", left, right)
}

// lor performs logical OR operation
func lor(left, right Value) (Value, error) {
	l, okL := left.(bool)
	r, okR := right.(bool)
	if okL && okR {
		return l || r, nil
	}
	return nil, unifiederrors.EvalErrorf("logical OR requires booleans, got %T and %T", left, right)
}

// lss performs < comparison
func lss(left, right Value) (Value, error) {
	return numericCompare(left, right, func(l, r float64) bool { return l < r }, "<")
}

// gtr performs > comparison
func gtr(left, right Value) (Value, error) {
	return numericCompare(left, right, func(l, r float64) bool { return l > r }, ">")
}

// leq performs <= comparison
func leq(left, right Value) (Value, error) {
	return numericCompare(left, right, func(l, r float64) bool { return l <= r }, "<=")
}

// geq performs >= comparison
func geq(left, right Value) (Value, error) {
	return numericCompare(left, right, func(l, r float64) bool { return l >= r }, ">=")
}

// appendRight implements the >> operator: array >> element → [...array, element]
func appendRight(left, right Value) (Value, error) {
	arr, ok := left.([]interface{})
	if !ok {
		return nil, unifiederrors.EvalErrorf(">> requires array on left side, got %T", left)
	}
	result := make([]interface{}, 0, len(arr)+1)
	result = append(result, arr...)
	result = append(result, right)
	return result, nil
}

// arrayAppend implements the << operator: array << element → [...array, element]
func arrayAppend(left, right Value) (Value, error) {
	arr, ok := left.([]interface{})
	if !ok {
		return nil, unifiederrors.EvalErrorf("<< requires array on left side, got %T", left)
	}
	result := make([]interface{}, 0, len(arr)+1)
	result = append(result, arr...)
	result = append(result, right)
	return result, nil
}

// Unary operations

// negate performs unary negation
func negate(val Value) (Value, error) {
	switch v := val.(type) {
	case float64:
		return -v, nil
	case int:
		return -v, nil
	default:
		return nil, unifiederrors.EvalErrorf("cannot negate %T", val)
	}
}

// not performs logical NOT
func not(val Value) (Value, error) {
	if v, ok := val.(bool); ok {
		return !v, nil
	}
	return nil, unifiederrors.EvalErrorf("cannot apply NOT to %T", val)
}

// Helper functions

// numericOp performs a numeric operation with automatic type coercion.
func numericOp(left, right Value, intOp func(int, int) int, floatOp func(float64, float64) float64, opName, verb string) (Value, error) {
	// Prefer int arithmetic for precision when both operands are ints
	if l, r, ok := BothInts(left, right); ok {
		return intOp(l, r), nil
	}

	// Fall back to float64 for mixed int/float operations
	if l, r, ok := BothFloats(left, right); ok {
		result := floatOp(l, r)
		if err := validateFloat(result, opName); err != nil {
			return nil, err
		}
		return result, nil
	}

	return nil, unifiederrors.EvalErrorf("cannot %s %T and %T", verb, left, right)
}

// numericCompare performs a numeric comparison with automatic type coercion.
func numericCompare(left, right Value, op func(float64, float64) bool, opName string) (Value, error) {
	if l, r, ok := BothFloats(left, right); ok {
		return op(l, r), nil
	}
	return nil, unifiederrors.EvalErrorf("cannot compare %T and %T with %s", left, right, opName)
}

// validateFloat checks if float result is valid
func validateFloat(v float64, op string) error {
	if math.IsNaN(v) {
		return unifiederrors.EvalErrorf("%s resulted in NaN", op)
	}
	if math.IsInf(v, 0) {
		return unifiederrors.EvalErrorf("%s resulted in infinity", op)
	}
	return nil
}

// evalBinaryOp evaluates binary operations
func evalBinaryOp(left, right Value, op token.Token, pos token.Pos) (Value, error) {
	switch op {
	case token.ADD:
		return add(left, right)
	case token.SUB:
		return sub(left, right)
	case token.MUL:
		return mul(left, right)
	case token.QUO:
		result, err := quo(left, right)
		if err != nil {
			// If it's a unified error, upgrade it with position info using posError to preserve pos
			if unifiedErr, ok := err.(*unifiederrors.Error); ok && unifiedErr.Message == "division by zero" {
				return nil, &posError{msg: "division by zero", pos: pos}
			}
		}
		return result, err
	case token.EQL:
		return numericEquals(left, right), nil
	case token.NEQ:
		return !numericEquals(left, right), nil
	case token.LSS:
		return lss(left, right)
	case token.GTR:
		return gtr(left, right)
	case token.LEQ:
		return leq(left, right)
	case token.GEQ:
		return geq(left, right)
	case token.LAND:
		return land(left, right)
	case token.LOR:
		return lor(left, right)
	case token.SHR: // >> append: array >> element
		return appendRight(left, right)
	case token.SHL: // << append: array << element
		return arrayAppend(left, right)
	case token.XOR: // ~= coercion equality
		return CoerceEquals(left, right), nil
	default:
		return nil, newPosError(fmt.Sprintf("unsupported operation: %v %v %v", left, op, right), pos)
	}
}

// evalUnaryOp evaluates unary operations
func evalUnaryOp(val Value, op token.Token, pos token.Pos) (Value, error) {
	switch op {
	case token.SUB:
		return negate(val)
	case token.NOT:
		return not(val)
	default:
		return nil, newPosError(fmt.Sprintf("unsupported unary operation: %v %v", op, val), pos)
	}
}

// evalIndex evaluates index operations: obj[idx]
func evalIndex(obj Value, idx Value, pos token.Pos) (Value, error) {
	if obj == nil {
		return nil, nil
	}

	switch v := obj.(type) {
	case []interface{}, XMLMultiValue:
		// Convert to []interface{} if it's XMLMultiValue for uniform processing
		var arr []interface{}
		if mv, ok := v.(XMLMultiValue); ok {
			arr = []interface{}(mv)
		} else {
			arr = v.([]interface{})
		}

		switch i := idx.(type) {
		case int:
			if i < 0 {
				i += len(arr)
			}
			if i < 0 || i >= len(arr) {
				return nil, newPosError(fmt.Sprintf("array index out of bounds: %d", i), pos)
			}
			return arr[i], nil
		case string:
			if strings.HasSuffix(i, "?") {
				key := strings.TrimSuffix(i, "?")
				// Safe access: extract field from all elements, return nil if not map or field missing
				result := make([]interface{}, 0, len(arr))
				for _, item := range arr {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if val, ok := itemMap[key]; ok {
							result = append(result, val)
						} else {
							result = append(result, nil)
						}
					} else {
						result = append(result, nil)
					}
				}
				return result, nil
			}
			// DataWeave compatibility: applying a string index to an array extracts that field from all elements
			result := make([]interface{}, 0, len(arr))
			for _, item := range arr {
				if itemMap, ok := item.(map[string]interface{}); ok {
					result = append(result, itemMap[i])
				} else {
					result = append(result, nil)
				}
			}
			return result, nil
		default:
			return nil, newPosError(fmt.Sprintf("array index must be an integer or string, got %T", idx), pos)
		}
	case string:
		switch i := idx.(type) {
		case int:
			if i < 0 {
				i += len(v)
			}
			if i < 0 || i >= len(v) {
				return nil, newPosError(fmt.Sprintf("string index out of bounds: %d", i), pos)
			}
			return v[i : i+1], nil
		default:
			return nil, newPosError(fmt.Sprintf("string index must be an integer, got %T", idx), pos)
		}
	case map[string]interface{}:
		switch i := idx.(type) {
		case string:
			if i == "@" {
				// Return all attributes (keys starting with @)
				attrs := make(map[string]interface{})
				for k, val := range v {
					if strings.HasPrefix(k, "@") {
						attrs[k] = val
					}
				}
				return attrs, nil
			}
			if strings.HasSuffix(i, "?") {
				key := strings.TrimSuffix(i, "?")
				if val, ok := v[key]; ok {
					return val, nil
				}
				return nil, nil
			}
			return v[i], nil
		case int:
			// Support ordinal indexing on objects
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			if i < 0 || i >= len(keys) {
				return nil, newPosError(fmt.Sprintf("object index out of bounds: %d (object has %d keys)", i, len(keys)), pos)
			}
			return v[keys[i]], nil
		default:
			return nil, newPosError(fmt.Sprintf("map key must be a string or int, got %T", idx), pos)
		}
	default:
		return nil, newPosError(fmt.Sprintf("cannot index into %T", obj), pos)
	}
}

// CoerceEquals implements the ~= coercion equality operator
func CoerceEquals(left, right Value) bool {
	if left == right {
		return true
	}

	leftNum, leftIsNum := tryParseNumber(left)
	rightNum, rightIsNum := tryParseNumber(right)
	if leftIsNum && rightIsNum {
		return leftNum == rightNum
	}

	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	return leftStr == rightStr
}

// tryParseNumber attempts to convert a value to float64 for coercion equality
func tryParseNumber(v Value) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case float64:
		return val, true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// numericEquals compares two values for equality, treating int and float64 as comparable
func numericEquals(left, right Value) bool {
	if left == right {
		return true
	}

	leftNum, leftIsNum := asNumber(left)
	rightNum, rightIsNum := asNumber(right)
	if leftIsNum && rightIsNum {
		return leftNum == rightNum
	}

	return false
}

// asNumber converts int or float64 to float64 for numeric comparison
func asNumber(v Value) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// callUserDefinedFunction calls a user-defined function (Lambda) with the provided arguments
func callUserDefinedFunction(lambda *Lambda, args []Value, e *ast.CallExpr, context Context, depth int) (Value, error) {
	if len(args) != lambda.ParamCount() {
		return nil, newPosError(fmt.Sprintf("function expects %d arguments, got %d", lambda.ParamCount(), len(args)), e.Pos())
	}

	fnContext := copyContext(context)
	for i, param := range lambda.Params {
		fnContext[param.Name] = args[i]
	}

	return evalASTWithDepth(lambda.BodyAST, fnContext, depth+1)
}

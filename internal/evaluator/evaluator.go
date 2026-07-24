package evaluator

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/sourcemap"
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

func (l *Lambda) repr() string {
	if l == nil {
		return "<function>"
	}
	paramList := strings.Join(l.ParamNames(), ", ")
	return fmt.Sprintf("(lambda: [%s] => %s)", paramList, l.Body)
}

// String implements fmt.Stringer so nested lambda values render deterministically.
func (l *Lambda) String() string {
	return l.repr()
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
	return []byte(fmt.Sprintf("\"%s\"", l.repr())), nil
}

// ErrorContext holds information needed for error reporting and location mapping
type ErrorContext struct {
	ExprStr    string
	Mapping    []int
	BodyOffset int
	FullRaw    string
	SourceMap  sourcemap.Map
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
func Evaluate(exprStr string, evalCtx Context, mapping []int, bodyOffset int, fullRaw string) (Value, error) {
	return EvaluateWithGoContext(exprStr, evalCtx, context.Background(), mapping, bodyOffset, fullRaw)
}

// EvaluateWithGoContext parses and evaluates the expression with Go context support.
func EvaluateWithGoContext(exprStr string, evalCtx Context, goCtx context.Context, mapping []int, bodyOffset int, fullRaw string) (Value, error) {
	return EvaluateWithGoContextAndSourceMap(exprStr, evalCtx, goCtx, sourcemap.New(fullRaw, exprStr, mapping).WithBaseOffset(bodyOffset))
}

// EvaluateWithGoContextAndSourceMap parses and evaluates an expression with an explicit source map.
func EvaluateWithGoContextAndSourceMap(exprStr string, evalCtx Context, goCtx context.Context, sourceMap sourcemap.Map) (Value, error) {
	scope := NewScope(evalCtx).WithGoContext(goCtx)
	ctx := &ErrorContext{
		ExprStr:    exprStr,
		Mapping:    sourceMap.Positions(),
		BodyOffset: 0,
		FullRaw:    "",
		SourceMap:  sourceMap,
	}
	return EvaluateWithScopeAndContext(exprStr, scope, ctx)
}

// EvaluateWithContext parses and evaluates the expression with error context.
func EvaluateWithContext(exprStr string, evalCtx Context, errCtx *ErrorContext) (Value, error) {
	return EvaluateWithContextAndGoContext(exprStr, evalCtx, context.Background(), errCtx)
}

// EvaluateWithContextAndGoContext parses and evaluates the expression with error and Go contexts.
func EvaluateWithContextAndGoContext(exprStr string, evalCtx Context, goCtx context.Context, errCtx *ErrorContext) (Value, error) {
	scope := NewScope(evalCtx).WithGoContext(goCtx)
	return EvaluateWithScopeAndContext(exprStr, scope, errCtx)
}

// EvaluateWithScopeAndContext parses and evaluates an expression with an explicit scope.
func EvaluateWithScopeAndContext(exprStr string, scope *Scope, errCtx *ErrorContext) (Value, error) {
	expr, err := parser.ParseExpr(exprStr)
	if err != nil {
		return nil, errCtx.FormatParseError(err, exprStr)
	}

	result, err := evalASTInScopeWithDepth(expr, scope, 0)
	if err != nil {
		return nil, errCtx.FormatEvalError(err)
	}
	return result, nil
}

// FormatEvalError converts a unified error to a user-friendly error with line/column information
func (ec *ErrorContext) FormatEvalError(err error) error {
	return ec.resolvedSourceMap().FormatEvalError(err)
}

// FormatParseError converts a parser error to a user-friendly error with line/column information
func (ec *ErrorContext) FormatParseError(err error, exprStr string) error {
	sm := ec.resolvedSourceMap()
	if sm.Generated() == "" {
		sm = sourcemap.New(ec.FullRaw, exprStr, ec.Mapping).WithBaseOffset(ec.BodyOffset)
	}
	return sm.FormatParseError(err)
}

func (ec *ErrorContext) resolvedSourceMap() sourcemap.Map {
	if ec.SourceMap.Generated() != "" || len(ec.SourceMap.Positions()) > 0 {
		return ec.SourceMap
	}
	return sourcemap.New(ec.FullRaw, ec.ExprStr, ec.Mapping).WithBaseOffset(ec.BodyOffset)
}

// evalAST evaluates an expression with initial depth of 0 using the visitor pattern.
func evalAST(expr ast.Expr, context Context) (Value, error) {
	return evalASTInScopeWithDepth(expr, NewScope(context), 0)
}

// evalASTWithDepth evaluates an expression using the visitor pattern with the given depth.
func evalASTWithDepth(expr ast.Expr, context Context, depth int) (Value, error) {
	return evalASTInScopeWithDepth(expr, NewScope(context), depth)
}

func evalASTInScopeWithDepth(expr ast.Expr, scope *Scope, depth int) (Value, error) {
	visitor := NewDefaultVisitorForScope(scope, depth)
	return visitor.Visit(expr)
}

// newLambdaInvocationScope returns the lexical scope captured when the lambda
// was defined. Caller locals do not implicitly leak into lambda/function bodies;
// only explicit invocation bindings should be layered on top of this scope.
func newLambdaInvocationScope(lambda *Lambda, caller *Scope) *Scope {
	if lambda != nil && lambda.Env != nil {
		runtime := Runtime{}
		if caller != nil {
			runtime = caller.Runtime
		}
		return NewScopeWithRuntime(copyContext(lambda.Env), runtime)
	}
	if caller != nil {
		return NewScopeWithRuntime(make(Context), caller.Runtime)
	}
	return NewScope(nil)
}

func evalLambdaWithBindingsAtDepth(lambda *Lambda, caller *Scope, depth int, bind func(Context)) (Value, error) {
	lambdaScope := newLambdaInvocationScope(lambda, caller)
	if bind != nil {
		bind(lambdaScope.Vars)
	}
	return evalASTInScopeWithDepth(lambda.BodyAST, lambdaScope, depth)
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
func evalIdent(e *ast.Ident, context Context) (Value, error) {
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
	return nil, &posError{
		msg: fmt.Sprintf("unresolved reference: %s", e.Name),
		pos: e.Pos(),
	}
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
		return l + coerceToString(right), nil
	}
	if r, okR := right.(string); okR {
		// Non-string + string: convert left to string
		return coerceToString(left) + r, nil
	}

	// Special case: array concatenation (supports array + array)
	if l, okL := left.(Array); okL {
		if r, okR := right.(Array); okR {
			result := make(Array, 0, len(l)+len(r))
			result = append(result, l...)
			result = append(result, r...)
			return result, nil
		}
	}

	if l, r, ok := BothInts(left, right); ok {
		return checkedIntOp(l, r, "addition")
	}
	return numericOp(left, right,
		func(l, r float64) float64 { return l + r },
		"addition", "add")
}

// sub performs subtraction or object key removal
func sub(left, right Value) (Value, error) {
	// Special case: object key removal (object - "key" → object without that key)
	if obj, ok := left.(Object); ok {
		if key, ok := right.(string); ok {
			result := make(Object, len(obj))
			for k, v := range obj {
				if k != key {
					result[k] = v
				}
			}
			return result, nil
		}
	}

	if l, r, ok := BothInts(left, right); ok {
		return checkedIntOp(l, r, "subtraction")
	}
	return numericOp(left, right,
		func(l, r float64) float64 { return l - r },
		"subtraction", "subtract")
}

// mul performs multiplication
func mul(left, right Value) (Value, error) {
	if l, r, ok := BothInts(left, right); ok {
		return checkedIntOp(l, r, "multiplication")
	}
	return numericOp(left, right,
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
	if l, r, ok := BothInts(left, right); ok {
		return exactIntDivision(l, r)
	}
	return numericOp(left, right, func(l, r float64) float64 { return l / r }, "division", "divide")
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
	return numericCompare(left, right, func(cmp int) bool { return cmp < 0 }, "<")
}

// gtr performs > comparison
func gtr(left, right Value) (Value, error) {
	return numericCompare(left, right, func(cmp int) bool { return cmp > 0 }, ">")
}

// leq performs <= comparison
func leq(left, right Value) (Value, error) {
	return numericCompare(left, right, func(cmp int) bool { return cmp <= 0 }, "<=")
}

// geq performs >= comparison
func geq(left, right Value) (Value, error) {
	return numericCompare(left, right, func(cmp int) bool { return cmp >= 0 }, ">=")
}

// prepend implements the >> operator: element >> array → [element, ...array]
func prepend(left, right Value) (Value, error) {
	arr, ok := right.(Array)
	if !ok {
		return nil, unifiederrors.EvalErrorf(">> requires array on right side, got %T", right)
	}
	result := make(Array, 0, len(arr)+1)
	result = append(result, left)
	result = append(result, arr...)
	return result, nil
}

// arrayAppend implements the << operator: array << element → [...array, element]
func arrayAppend(left, right Value) (Value, error) {
	arr, ok := left.(Array)
	if !ok {
		return nil, unifiederrors.EvalErrorf("<< requires array on left side, got %T", left)
	}
	result := make(Array, 0, len(arr)+1)
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
		if v == minInt() {
			return nil, unifiederrors.EvalError("integer overflow during negation")
		}
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

// numericOp performs a float or mixed numeric operation. Integer-only
// arithmetic is handled separately so it cannot silently wrap.
func numericOp(left, right Value, floatOp func(float64, float64) float64, opName, verb string) (Value, error) {
	if hasInexactFloatInt(left, right) {
		return exactMixedNumericOp(left, right, opName)
	}

	if l, r, ok := BothFloats(left, right); ok {
		result := floatOp(l, r)
		if err := validateFloat(result, opName); err != nil {
			return nil, err
		}
		return result, nil
	}

	return nil, unifiederrors.EvalErrorf("cannot %s %T and %T", verb, left, right)
}

// numericCompare uses the exact rational values represented by ints and
// float64s, avoiding lossy int-to-float coercion above 2^53.
func numericCompare(left, right Value, op func(int) bool, opName string) (Value, error) {
	l, okL := exactNumericRat(left)
	r, okR := exactNumericRat(right)
	if okL && okR {
		return op(l.Cmp(r)), nil
	}
	return nil, unifiederrors.EvalErrorf("cannot compare %T and %T with %s", left, right, opName)
}

func checkedIntOp(left, right int, opName string) (Value, error) {
	l := big.NewInt(int64(left))
	r := big.NewInt(int64(right))
	result := new(big.Int)
	switch opName {
	case "addition":
		result.Add(l, r)
	case "subtraction":
		result.Sub(l, r)
	case "multiplication":
		result.Mul(l, r)
	default:
		return nil, unifiederrors.EvalErrorf("unsupported integer operation %s", opName)
	}
	if !result.IsInt64() || int64(int(result.Int64())) != result.Int64() {
		return nil, unifiederrors.EvalErrorf("integer overflow during %s", opName)
	}
	return int(result.Int64()), nil
}

func exactIntDivision(left, right int) (Value, error) {
	result := new(big.Rat).SetFrac(big.NewInt(int64(left)), big.NewInt(int64(right)))
	if result.IsInt() {
		if !result.Num().IsInt64() {
			return nil, unifiederrors.EvalError("integer overflow during division")
		}
		n := result.Num().Int64()
		if int64(int(n)) != n {
			return nil, unifiederrors.EvalError("integer overflow during division")
		}
		return int(n), nil
	}
	quotient, exact := result.Float64()
	if !exact && (!intExactlyRepresentableAsFloat(left) || !intExactlyRepresentableAsFloat(right)) {
		return nil, unifiederrors.EvalError("numeric precision loss during division")
	}
	if err := validateFloat(quotient, "division"); err != nil {
		return nil, err
	}
	return quotient, nil
}

func exactMixedNumericOp(left, right Value, opName string) (Value, error) {
	l, okL := exactNumericRat(left)
	r, okR := exactNumericRat(right)
	if !okL || !okR {
		return nil, unifiederrors.EvalErrorf("cannot apply %s to %T and %T", opName, left, right)
	}

	result := new(big.Rat)
	switch opName {
	case "addition":
		result.Add(l, r)
	case "subtraction":
		result.Sub(l, r)
	case "multiplication":
		result.Mul(l, r)
	case "division":
		if r.Sign() == 0 {
			return nil, unifiederrors.EvalError("division by zero")
		}
		result.Quo(l, r)
	default:
		return nil, unifiederrors.EvalErrorf("unsupported numeric operation %s", opName)
	}

	if result.IsInt() && result.Num().IsInt64() {
		n := result.Num().Int64()
		if int64(int(n)) == n {
			return int(n), nil
		}
	}
	f, exact := result.Float64()
	if !exact {
		return nil, unifiederrors.EvalErrorf("numeric precision loss during %s", opName)
	}
	if err := validateFloat(f, opName); err != nil {
		return nil, err
	}
	return f, nil
}

func hasInexactFloatInt(left, right Value) bool {
	for _, value := range []Value{left, right} {
		if n, ok := value.(int); ok && !intExactlyRepresentableAsFloat(n) {
			return true
		}
	}
	return false
}

func intExactlyRepresentableAsFloat(value int) bool {
	asInt := new(big.Rat).SetInt64(int64(value))
	asFloat := new(big.Rat)
	asFloat.SetFloat64(float64(value))
	return asInt.Cmp(asFloat) == 0
}

func exactNumericRat(value Value) (*big.Rat, bool) {
	switch n := value.(type) {
	case int:
		return new(big.Rat).SetInt64(int64(n)), true
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return nil, false
		}
		return new(big.Rat).SetFloat64(n), true
	default:
		return nil, false
	}
}

func minInt() int {
	return -int(^uint(0)>>1) - 1
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
	case token.SHR: // >> prepend: element >> array
		return prepend(left, right)
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
	case Array:
		return evalArrayIndex(v, idx, pos)
	case XMLMultiValue:
		return evalArrayIndex(Array(v), idx, pos)
	case string:
		return evalStringIndex(v, idx, pos)
	case Object:
		return evalObjectIndex(v, idx, pos)
	default:
		return nil, newPosError(fmt.Sprintf("cannot index into %T", obj), pos)
	}
}

// getObjectValue retrieves a key from an object, with namespace fallback (# -> :).
func getObjectValue(obj Object, key string) (Value, bool) {
	if val, ok := obj[key]; ok {
		return val, true
	}
	if strings.Contains(key, "#") {
		if val, ok := obj[strings.ReplaceAll(key, "#", ":")]; ok {
			return val, true
		}
	}
	return nil, false
}

// extractNamespaceValue returns a best-effort namespace URI for an XML-like object.
func extractNamespaceValue(obj Object) Value {
	nsVal, ok := obj["@xmlns"]
	if !ok {
		return nil
	}
	nsMap, ok := nsVal.(Object)
	if !ok {
		return nil
	}
	if def, ok := nsMap["#default"]; ok {
		return def
	}
	if len(nsMap) == 1 {
		for _, v := range nsMap {
			return v
		}
	}
	return nil
}

// CoerceEquals implements the ~= coercion equality operator
func CoerceEquals(left, right Value) bool {
	// Guard against uncomparable types (slices, maps) that would panic with ==.
	switch left.(type) {
	case Array, Object:
		return reflect.DeepEqual(left, right)
	}
	switch right.(type) {
	case Array, Object:
		return reflect.DeepEqual(left, right)
	}
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
	if f, ok := ToFloat(v); ok {
		return f, true
	}
	if s, ok := v.(string); ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// numericEquals compares two values for equality, treating int and float64 as comparable.
// Handles uncomparable types (slices, maps) via reflect.DeepEqual.
func numericEquals(left, right Value) bool {
	// Handle uncomparable types (slices, maps) that would panic with ==.
	switch left.(type) {
	case Array, Object:
		return reflect.DeepEqual(left, right)
	}
	switch right.(type) {
	case Array, Object:
		return reflect.DeepEqual(left, right)
	}

	if left == right {
		return true
	}

	leftNum, leftOK := exactNumericRat(left)
	rightNum, rightOK := exactNumericRat(right)
	if leftOK && rightOK {
		return leftNum.Cmp(rightNum) == 0
	}

	return false
}

// callUserDefinedFunction calls a user-defined function (Lambda) with the provided arguments
func callUserDefinedFunction(lambda *Lambda, args []Value, e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(args) != lambda.ParamCount() {
		return nil, newPosError(fmt.Sprintf("function expects %d arguments, got %d", lambda.ParamCount(), len(args)), e.Pos())
	}

	fnScope := scope.Copy()
	for i, param := range lambda.Params {
		fnScope.Vars[param.Name] = args[i]
	}

	return evalASTInScopeWithDepth(lambda.BodyAST, fnScope, depth+1)
}

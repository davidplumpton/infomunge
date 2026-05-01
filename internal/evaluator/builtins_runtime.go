package evaluator

import (
	"crypto/rand"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	unifiederrors "infomunge/internal/errors"
)

// UserError represents an error explicitly thrown by the user via fail().
type UserError struct {
	Message  string
	Kind     string
	Location string
	pos      token.Pos
}

func (e *UserError) Error() string {
	return e.Message
}

func (e *UserError) Pos() token.Pos {
	return e.pos
}

// callBuiltinUUID implements the uuid() function.
func callBuiltinUUID(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 0, "uuid takes no arguments", e); err != nil {
		return nil, err
	}

	// Generate 16 random bytes for a v4 UUID
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("failed to generate UUID: %s", err), e.Pos())
	}

	// Set version 4 and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	// Format as UUID string
	uuid := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return uuid, nil
}

// callBuiltinEvaluateCompatibilityFlag implements the evaluateCompatibilityFlag(flagName) function.
func callBuiltinEvaluateCompatibilityFlag(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "evaluateCompatibilityFlag requires exactly 1 argument: flagName", e); err != nil {
		return nil, err
	}

	flagName, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("evaluateCompatibilityFlag expects flagName to be a string, got %T", args[0]), e.Pos())
	}

	// Define default compatibility flags
	compatibilityFlags := Object{
		"allowUndefinedProperties": true,
		"allowNullArithmetic":      false,
		"allowImplicitConversion":  true,
		"strictTypeChecking":       false,
		"allowDynamicProperties":   true,
	}

	if value, ok := compatibilityFlags[flagName]; ok {
		return value, nil
	}

	// Return null if flag not found
	return nil, nil
}

// callBuiltinEnvVar implements the envVar(name) function.
func callBuiltinEnvVar(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "envVar requires exactly 1 argument: variable name", e); err != nil {
		return nil, err
	}

	varName, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("envVar expects a string argument, got %T", args[0]), e.Pos())
	}

	value, exists := os.LookupEnv(varName)
	if !exists {
		return nil, nil
	}
	return value, nil
}

// callBuiltinEnvVars implements the envVars() function.
func callBuiltinEnvVars(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 0, "envVars takes no arguments", e); err != nil {
		return nil, err
	}

	envMap := make(Object)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	return envMap, nil
}

// callBuiltinLog implements the log(value) function.
// Logs the value to stderr and returns it unchanged.
func callBuiltinLog(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "log requires exactly 1 argument: value", e); err != nil {
		return nil, err
	}

	fmt.Fprintln(os.Stderr, args[0])
	return args[0], nil
}

func logWithPrefix(prefix string, value Value) {
	fmt.Fprintf(os.Stderr, "[%s] %v\n", prefix, value)
}

// callBuiltinLogDebug implements the logDebug(value) function.
// Logs the value at DEBUG level and returns it unchanged.
func callBuiltinLogDebug(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "logDebug requires exactly 1 argument: value", e); err != nil {
		return nil, err
	}

	logWithPrefix("DEBUG", args[0])
	return args[0], nil
}

// callBuiltinLogInfo implements the logInfo(value) function.
// Logs the value at INFO level and returns it unchanged.
func callBuiltinLogInfo(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "logInfo requires exactly 1 argument: value", e); err != nil {
		return nil, err
	}

	logWithPrefix("INFO", args[0])
	return args[0], nil
}

// callBuiltinLogWarn implements the logWarn(value) function.
// Logs the value at WARN level and returns it unchanged.
func callBuiltinLogWarn(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "logWarn requires exactly 1 argument: value", e); err != nil {
		return nil, err
	}

	logWithPrefix("WARN", args[0])
	return args[0], nil
}

// callBuiltinLogError implements the logError(value) function.
// Logs the value at ERROR level and returns it unchanged.
func callBuiltinLogError(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 1, "logError requires exactly 1 argument: value", e); err != nil {
		return nil, err
	}

	logWithPrefix("ERROR", args[0])
	return args[0], nil
}

// callBuiltinLogWith implements the logWith(value, prefix) function.
// Logs the value with a custom prefix and returns it unchanged.
func callBuiltinLogWith(args []Value, e *ast.CallExpr) (Value, error) {
	if err := requireExactArgs(args, 2, "logWith requires exactly 2 arguments: value, prefix", e); err != nil {
		return nil, err
	}

	prefix, err := assertStringArg(args[1], 2, "logWith", e)
	if err != nil {
		return nil, err
	}

	logWithPrefix(prefix, args[0])
	return args[0], nil
}

// callBuiltinTry implements the try(delegate) function.
func callBuiltinTry(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 1 {
		return nil, newPosError("try requires exactly 1 argument: a delegate function", e.Pos())
	}

	// Evaluate the delegate expression
	result, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)

	if err != nil {
		// Build error object based on error type
		errorObj := buildErrorObject(err)
		return Object{
			"success": false,
			"error":   errorObj,
		}, nil
	}

	// If result is a zero-argument lambda, invoke it
	if lambda, ok := result.(*Lambda); ok && lambda.ParamCount() == 0 {
		result, err = callUserDefinedFunction(lambda, Array{}, e, scope, depth)
		if err != nil {
			errorObj := buildErrorObject(err)
			return Object{
				"success": false,
				"error":   errorObj,
			}, nil
		}
	}

	return Object{
		"success": true,
		"result":  result,
	}, nil
}

// callBuiltinOrElse implements the orElse(previous, orElse) function.
func callBuiltinOrElse(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("orElse requires exactly 2 arguments: previous TryResult, orElse value or function", e.Pos())
	}

	// Evaluate the previous argument (should be a TryResult)
	prev, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}

	tryResult, ok := prev.(Object)
	if !ok {
		return nil, newPosError("orElse: first argument must be a TryResult object", e.Pos())
	}

	success, hasSuccess := tryResult["success"].(bool)
	if !hasSuccess {
		return nil, newPosError("orElse: first argument must be a TryResult object with success field", e.Pos())
	}

	// If previous succeeded, return its result
	if success {
		return tryResult["result"], nil
	}

	// Previous failed, evaluate the orElse argument
	orElseVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		return nil, err
	}

	// If orElse is a zero-argument lambda, invoke it
	if lambda, ok := orElseVal.(*Lambda); ok && lambda.ParamCount() == 0 {
		return callUserDefinedFunction(lambda, Array{}, e, scope, depth)
	}

	return orElseVal, nil
}

// callBuiltinOrElseTry implements the orElseTry(previous, orElse) function.
func callBuiltinOrElseTry(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("orElseTry requires exactly 2 arguments: previous TryResult, orElse function", e.Pos())
	}

	// Evaluate the previous argument (should be a TryResult)
	prev, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}

	tryResult, ok := prev.(Object)
	if !ok {
		return nil, newPosError("orElseTry: first argument must be a TryResult object", e.Pos())
	}

	success, hasSuccess := tryResult["success"].(bool)
	if !hasSuccess {
		return nil, newPosError("orElseTry: first argument must be a TryResult object with success field", e.Pos())
	}

	// If previous succeeded, return it as-is
	if success {
		return tryResult, nil
	}

	// Previous failed, evaluate the orElse argument and wrap in TryResult
	orElseVal, err := evalASTInScopeWithDepth(e.Args[1], scope, depth)
	if err != nil {
		// orElse evaluation itself failed
		errorObj := buildErrorObject(err)
		return Object{
			"success": false,
			"error":   errorObj,
		}, nil
	}

	// If orElse is a zero-argument lambda, invoke it
	if lambda, ok := orElseVal.(*Lambda); ok && lambda.ParamCount() == 0 {
		result, err := callUserDefinedFunction(lambda, Array{}, e, scope, depth)
		if err != nil {
			errorObj := buildErrorObject(err)
			return Object{
				"success": false,
				"error":   errorObj,
			}, nil
		}
		return Object{
			"success": true,
			"result":  result,
		}, nil
	}

	return Object{
		"success": true,
		"result":  orElseVal,
	}, nil
}

// buildErrorObject creates the error object structure for TryResult.
func buildErrorObject(err error) Object {
	errorObj := Object{
		"message":  err.Error(),
		"location": "Unknown location",
		"stack":    Array{},
	}

	// Check for UserError (from fail())
	if userErr, ok := err.(*UserError); ok {
		errorObj["kind"] = userErr.Kind
		errorObj["message"] = userErr.Message
		errorObj["location"] = userErr.Location
	} else if unifiedErr, ok := err.(*unifiederrors.Error); ok && unifiedErr.Position != nil {
		errorObj["kind"] = "RuntimeException"
		errorObj["location"] = fmt.Sprintf("%s:%d:%d", unifiedErr.Position.Filename, unifiedErr.Position.Line, unifiedErr.Position.Column)
	} else if posErr, ok := err.(interface{ Pos() token.Pos }); ok {
		errorObj["kind"] = "RuntimeException"
		errorObj["location"] = fmt.Sprintf("position %d", posErr.Pos())
	} else {
		errorObj["kind"] = "RuntimeException"
	}

	return errorObj
}

// callBuiltinFail implements the fail(message) function.
func callBuiltinFail(args []Value, e *ast.CallExpr) (Value, error) {
	message := "Error" // default message
	if len(args) > 1 {
		return nil, newPosError("fail accepts at most 1 argument: message", e.Pos())
	}
	if len(args) == 1 {
		msg, ok := args[0].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("fail expects a string argument, got %T", args[0]), e.Pos())
		}
		message = msg
	}
	return nil, &UserError{
		Message:  message,
		Kind:     "UserException",
		Location: fmt.Sprintf("line %d", e.Pos()),
		pos:      e.Pos(),
	}
}

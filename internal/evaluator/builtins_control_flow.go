package evaluator

import (
	"fmt"
	"go/ast"
	"time"
)

// callBuiltinIfElse implements the __ifelse(condition, trueValue, falseValue) function.
func callBuiltinIfElse(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 3 {
		return nil, newPosError("if/else requires exactly 3 arguments: condition, trueValue, falseValue", e.Pos())
	}
	// Evaluate condition
	cond, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return nil, err
	}
	// Convert condition to boolean
	condBool, ok := cond.(bool)
	if !ok {
		return nil, newPosError(fmt.Sprintf("if condition must be a boolean, got %T", cond), e.Args[0].Pos())
	}
	// Evaluate and return appropriate branch
	if condBool {
		return evalASTInScopeWithDepth(e.Args[1], scope, depth)
	}
	return evalASTInScopeWithDepth(e.Args[2], scope, depth)
}

// callBuiltinWhile implements the __while(condition, body) function.
func callBuiltinWhile(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 2 {
		return nil, newPosError("while loop requires exactly 2 arguments: condition, body", e.Pos())
	}

	var result Value
	startTime := time.Now()
	deadline := getDeadline(scope, startTime)

loop:
	for i := 0; ; i++ {
		// Check for timeout
		if err := checkTimeout(i, startTime, deadline, e); err != nil {
			return nil, err
		}

		// Evaluate condition
		condBool, err := evaluateCondition(e, scope, depth)
		if err != nil {
			return nil, err
		}

		// Exit loop if condition is false
		if !condBool {
			break
		}

		// Execute body
		bodyResult, err := evaluateBody(e, scope, depth)
		if err != nil {
			return nil, err
		}

		// Check for control flow signals
		signal, isSignal := processBodyResult(bodyResult)
		if isSignal {
			switch signal.Type {
			case "break":
				break loop
			case "continue":
				continue loop
			}
		}

		// Only update result if it's not a signal
		if !isSignal {
			result = bodyResult
		}
	}

	return result, nil
}

// getDeadline retrieves the loop deadline from context or uses default timeout
func getDeadline(scope *Scope, startTime time.Time) time.Time {
	return scope.LoopDeadline(startTime)
}

// checkTimeout verifies the loop hasn't exceeded the deadline
func checkTimeout(i int, startTime, deadline time.Time, e *ast.CallExpr) error {
	// Check for timeout every TimeoutCheckInterval iterations to balance overhead and responsiveness
	if i%TimeoutCheckInterval == 0 && time.Now().After(deadline) {
		return newPosError(fmt.Sprintf("while loop timed out after %v", time.Since(startTime)), e.Pos())
	}
	return nil
}

// evaluateCondition evaluates the while condition and converts it to boolean
func evaluateCondition(e *ast.CallExpr, scope *Scope, depth int) (bool, error) {
	cond, err := evalASTInScopeWithDepth(e.Args[0], scope, depth)
	if err != nil {
		return false, err
	}

	condBool, ok := cond.(bool)
	if !ok {
		return false, newPosError(fmt.Sprintf("while condition must be a boolean, got %T", cond), e.Args[0].Pos())
	}
	return condBool, nil
}

// evaluateBody executes the while loop body and returns the result or control flow signal
func evaluateBody(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	return evalASTInScopeWithDepth(e.Args[1], scope, depth)
}

// processBodyResult handles control flow signals from the loop body
func processBodyResult(bodyResult Value) (*ControlFlowSignal, bool) {
	if signal, ok := bodyResult.(*ControlFlowSignal); ok {
		return signal, true
	}
	return nil, false
}

// callBuiltinBreak implements the __break function.
func callBuiltinBreak(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 0 {
		return nil, newPosError("break takes no arguments", e.Pos())
	}
	return &ControlFlowSignal{Type: "break"}, nil
}

// callBuiltinContinue implements the __continue function.
func callBuiltinContinue(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) != 0 {
		return nil, newPosError("continue takes no arguments", e.Pos())
	}
	return &ControlFlowSignal{Type: "continue"}, nil
}

// callBuiltinSeq implements the __seq(expr1, expr2, ...) function.
func callBuiltinSeq(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	if len(e.Args) < 1 {
		return nil, newPosError("seq requires at least 1 argument", e.Pos())
	}

	var result Value
	var err error

	// Evaluate all arguments in order
	for _, arg := range e.Args {
		result, err = evalASTInScopeWithDepth(arg, scope, depth)
		if err != nil {
			return nil, err
		}

		// If a control flow signal is encountered, propagate it immediately
		if _, ok := result.(*ControlFlowSignal); ok {
			return result, nil
		}
	}

	// Return the last evaluated result
	return result, nil
}

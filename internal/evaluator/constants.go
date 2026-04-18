package evaluator

import "time"

// Evaluation depth and recursion limits
const (
	// MaxEvalDepth is the maximum recursion depth for expression evaluation
	// to prevent infinite recursion and stack overflow.
	MaxEvalDepth = 100

	// MaxDeepDepth is the maximum recursion depth for deep search operations
	// in nested data structures. Higher than MaxEvalDepth to allow deep traversal.
	MaxDeepDepth = 1000
)

// Loop control constants
const (
	// DefaultLoopTimeout is the default maximum duration for while loops
	// if no deadline is set in the execution context.
	DefaultLoopTimeout = 30 * time.Second

	// TimeoutCheckInterval controls how often (in iterations) we check
	// if a loop has exceeded its timeout. Checking every iteration would
	// be expensive, so we check periodically to balance responsiveness and overhead.
	TimeoutCheckInterval = 100
)

# While Loop Implementation Design

## Overview

While loops will be added to InfoMunge to support iteration with state management. The implementation follows the same pattern as `if/else` expressions, transforming DataWeave-like syntax into Go-compatible function calls.

## Architecture

### 1. Syntax

```
while(condition) {
  body;
  var = var + 1
}
```

### 2. Transformation Pipeline

**Preprocessor** (`internal/preprocessor/preprocessor.go`) will transform:
```
while(x < 10) {
  x = x + 1
}
```

Into a function call:
```
__while(x < 10, (): {
  x = x + 1
})
```

Note: The body is passed as a lambda closure to maintain variable scope.

### 3. Evaluator Implementation

**Evaluator** (`internal/evaluator/evaluator.go` and `internal/evaluator/builtins.go`) will implement `__while` as a special builtin that:

1. Evaluates the condition
2. If true, executes the body (which updates variables in context)
3. Re-evaluates the condition
4. Repeats until condition is false or a `break` statement is encountered
5. Returns the result of the last body expression

### 4. Key Design Decisions

- **State Management**: Variables modified in the loop body must persist across iterations. This requires mutable context variables.
- **Break/Continue**: These control flow statements must be implemented as special error/signal types that propagate up.
- **Timeout Protection**: Tests enforce a 5-second timeout by default to prevent infinite loops.
- **Lambda Closure**: The body is represented as a lambda to handle variable scoping naturally.

### 5. Related Features for Future Work

- `do...while` - executes body at least once before checking condition
- Counter variable shorthand (e.g., `for(i = 0; i < 10; i++)`)

## Testing Strategy

Feature file: `test/features/while_loops.feature` includes scenarios for:
1. Basic while loop counting
2. Collection accumulation
3. Break/continue statements  
4. Early termination
5. Infinite loop timeout protection

All tests run with default 5-second timeout via the test harness in `test/godog_test.go`.

## Implementation Order

1. ✅ Test infrastructure with timeout (COMPLETE)
2. ✅ Preprocessor handler for `while` syntax (COMPLETE - basic parsing done)
3. ✅ `__while` builtin in evaluator
4. ✅ Assignment expression support
5. ✅ Break/continue control flow signals
6. ✅ Feature test scenarios

## Status Notes

Assignment expressions are supported by rewriting `x = y` into `__assign("x", y)` and mutating the evaluation context. Break/continue are implemented via control-flow signals that the while builtin handles.

## Complexity Considerations

- **Variable Mutation**: The biggest complexity is handling mutable state. Variables must be modified in-place in the context map during loop execution.
  - Solution: Add `__assign(varName, value)` builtin that mutates context
- **Control Flow**: Break/continue require a signal mechanism (similar to exceptions).
- **Nested Loops**: Must support loops within loops with proper variable scoping.
- **Infinite Loop Protection**: Timeout in test harness prevents test hangs, but long-running scripts will need progress tracking.

# Refactoring Plan for infomunge-bmp.3: Replace interface{} with Concrete Types and Generics

## Overview
Replace generic `interface{}` types throughout the codebase with concrete type aliases and Go 1.18+ generics where applicable. This reduces runtime panics and improves type safety.

## Key Findings

### Current Usage Patterns
1. **pkg/formats**: Heavy use of `interface{}` for generic data representation
   - `Reader` and `Writer` function types use `interface{}`
   - Maps: `map[string]interface{}` for objects
   - Slices: `[]interface{}` for arrays
   - This is appropriate here since these are the API boundaries

2. **internal/evaluator**: Uses `interface{}` for expression values
   - Return types from evaluations: `(interface{}, error)`
   - Context maps: `map[string]interface{}`
   - This is appropriate for a dynamic expression evaluator

3. **internal/runner**: Module loading uses `map[string]interface{}`

### Type Alias Strategy
Create common type aliases to make code more semantic:

```go
// Core value types
type Value = interface{}           // Generic evaluated value
type Object = map[string]interface{}
type Array = []interface{}

// For static APIs that can be more specific
type Config = map[string]interface{}
```

### Priority Refactoring Areas (from low to high impact)

#### Low Impact (Documentation/Clarity)
- [ ] Add type aliases in central types file
- [ ] Update function signatures to use aliases where semantically appropriate
- [ ] Comment where and why interface{} is necessary

#### Medium Impact (Type Safety)
- [ ] Replace type assertions with helper functions that return errors
- [ ] Add validation helpers for common patterns
- [ ] Create specific interfaces for matchers and visitors

#### High Impact (Behavior Changes - requires testing)
- [ ] Consider using generics for collection operations
- [ ] Refactor assertion patterns to reduce panics

## Implementation Steps

1. **Create types.go in pkg/formats** (if not exists)
   - Define Object, Array type aliases
   - Define Reader and Writer function types with these aliases
   
2. **Add assertion helpers in internal/evaluator**
   - `AsObject(v Value) (Object, error)`
   - `AsArray(v Value) (Array, error)`
   - `AsString(v Value) (string, error)`
   - etc.

3. **Update key function signatures**
   - Start with public APIs
   - Then internal functions
   - Update tests to validate

4. **Add tests for assertion helpers**
   - Verify error handling
   - Test type coercion in type system

## Notes
- Go 1.18 generics may not be needed initially; focus on aliases first
- The project already has a type system (TypeDef, coerceToType, etc.)
- Many assertions are in critical paths; need careful testing

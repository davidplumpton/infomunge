# Operator Chaining Body Boundary Detection Audit

## Executive Summary

**CONFIRMED BUGS FOUND:**

1. **Missing operators in stopping lists**: `maxBy`, `minBy`, `distinctBy`, `filterObject`, `mapObject` are NOT recognized as body boundary markers. Chaining with these operators fails.

2. **Implicit lambda chaining broken**: Even operators that ARE in the stopping list (like `map`, `filter`) don't work correctly when chaining implicit $ lambdas without parentheses.

3. **Existing test uses workaround**: The existing "Chained map and filter with $" test uses parentheses to work around the issue.

## Verified Test Results

### Bug 1: Missing operators in stopping lists

```bash
# These all fail - the preceding operator's body doesn't stop at maxBy/minBy/distinctBy
payload filter $ > 1 maxBy $      # Error: maxBy expects an array, got int
payload filter $ > 1 minBy $      # Error: minBy expects an array, got int
payload map $ * 2 distinctBy $    # Error: distinctBy expects an array, got int
payload filterObject ... mapObject ... # Error: mapObject expects an object, got int
```

### Bug 2: Chaining implicit lambdas without parentheses

```bash
# These fail even though map/filter ARE in the stopping list
payload filter $ > 1 map $ * 2    # Error: illegal character U+0024 '$'
payload map $ * 2 filter $ > 5    # Error: cannot compare string and int with >
```

### What works:

```bash
# Explicit lambdas with chaining work
payload filter (x) -> x > 1 map (y) -> y * 2  # Works: [4,6,8,10]

# Parenthesized expressions work
(payload filter $ > 1) maxBy $                 # Works: 5
(payload map $ * 2) filter $ > 5               # Works: [6,8,10]

# Single operator with implicit $ works
payload maxBy $                                # Works: 5
payload filter $ > 1                           # Works
```

## Root Cause Analysis

### Missing operators issue

In `scanLambdaBody` (line 201), `replaceArrowFunctions` (lines 311-334), and `replaceCollectionOperator` (line 128), the `collOps` list is:
```go
[]string{"map ", "filter ", "reduce ", "flatMap ", "groupBy ", "pluck ", "sort ", "orderBy "}
```

Missing from this list:
- `maxBy `
- `minBy `
- `distinctBy `
- `filterObject `
- `mapObject `

### Implicit lambda chaining issue

The `replaceImplicitLambdas` function iterates operators and calls `replaceImplicitLambdaForOp`. When multiple operators are chained, the body boundary detection isn't correctly stopping at the subsequent operator in all cases.

## Files Involved

1. `internal/preprocessor/transformers_lambda.go`
   - `scanLambdaBody` (line 168) - needs complete operator list
   - `replaceImplicitLambdaForOp` (line 69) - needs to detect chained operators
   - `replaceArrowFunctions` (line 268) - needs complete operator list

2. `internal/preprocessor/transformers_collection.go`
   - `replaceCollectionOperator` (line 75) - needs complete operator list

## Recommendations

### Immediate Fix

1. Create a single constant list of all collection operators:
```go
var allCollectionOperators = []string{
    "map ", "filter ", "reduce ", "flatMap ", "groupBy ",
    "pluck ", "sort ", "orderBy ", "maxBy ", "minBy ",
    "distinctBy ", "filterObject ", "mapObject ",
}
```

2. Use this list in all three locations:
   - `scanLambdaBody`
   - `replaceArrowFunctions`
   - `replaceCollectionOperator`

### Additional Investigation Needed

1. Debug why implicit lambda chaining fails even with operators in the list
2. Consider whether `default` and other binary operators should be stopping operators
3. Add comprehensive test coverage for all chaining scenarios

## Test Cases to Add

```gherkin
Scenario: Chain filter and maxBy with implicit $
  payload filter $ > 1 maxBy $
  # Expected: 5 (from [2,3,4,5])

Scenario: Chain filter and minBy with implicit $
  payload filter $ > 1 minBy $
  # Expected: 2 (from [2,3,4,5])

Scenario: Chain map and distinctBy with implicit $
  [1,1,2,2,3] map $ * 2 distinctBy $
  # Expected: [2,4,6]

Scenario: Chain filterObject and mapObject
  {"a":1,"b":2} filterObject (k,v) -> v > 1 mapObject (k,v) -> {(k): v*2}
  # Expected: {"b": 4}

Scenario: Chain filter and map with implicit $ (no parens)
  [1,2,3,4,5] filter $ > 2 map $ * 2
  # Expected: [6,8,10]
```

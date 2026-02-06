# DataWeave to InfoMunge Conversion Summary

## Overview

This document summarizes DataWeave conversion coverage in InfoMunge and reflects the current compatibility direction.

Compatibility target:
- Behavior should match DataWeave by default.
- The primary intentional surface difference is script header token (`%im 0.1` vs `%dw 2.0`).

## Deliverables

### 1. Documentation
- `docs/DATAWEAVE_COOKBOOK_EQUIVALENTS.md`

### 2. Test Coverage
- `test/features/dataweave_cookbook_examples.feature`
- Additional compatibility-focused scenarios across:
  - `test/features/array_indexing.feature`
  - `test/features/object_values.feature`
  - `test/features/implicit_lambda.feature`
  - `test/features/reduce_operator.feature`
  - `test/features/object_functions.feature`
  - `test/features/pluck_function.feature`

## Compatibility Snapshot

| Aspect | DataWeave | InfoMunge |
|--------|-----------|-----------|
| Header directive | `%dw 2.0` | `%im 0.1` |
| Array field selection | `items.name` | `items.name` |
| Multi-value selectors | `.*`, `.*field` | `.*`, `.*field` |
| Object expression values | `{k: a ++ b}` | `{k: a ++ b}` |
| mapObject output forms | object / pair patterns | object / pair patterns |
| mapObject/filterObject/pluck object-lambda order | DataWeave semantics | DataWeave-default semantics, explicit legacy aliases preserved |
| Lambda shortcuts | `$`, `$$` | `$`, `$$` |
| Reduce initial value syntax | supported | supported |

## Notes

Earlier versions of this summary listed multiple syntax gaps that have since been implemented.
Current docs should treat compatibility as the default runtime behavior, not an optional mode.

## Running Compatibility-Focused Tests

```bash
# Full suite
go test -v ./test

# Targeted compatibility scenarios
GODOG_PATHS=features/object_functions.feature,features/pluck_function.feature go test -v ./test -run TestFeatures
```

## If You Find a Mismatch

Create a beads issue and treat it as a default-compatibility bug/feature task.
Do not add a compatibility-mode flag as a workaround.

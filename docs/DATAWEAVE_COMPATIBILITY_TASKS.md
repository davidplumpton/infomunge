# DataWeave Compatibility Status

## Goal

InfoMunge should be compatible with DataWeave by default.

Target state:
- DataWeave scripts should run in InfoMunge with equivalent behavior.
- The primary intentional syntax difference is the header token:
  - DataWeave: `%dw 2.0`
  - InfoMunge: `%im 0.1`
- Do not add or rely on a separate compatibility mode flag.

## Current Status

Compatibility work originally tracked under epic `infomunge-vv1` is complete and closed.
Recent follow-up tasks for object-lambda parameter-order behavior and regression coverage are also complete.

## Implemented Compatibility Areas

The following previously identified gaps are implemented and covered by tests:

1. Array field extraction with dot syntax (`items.name`) works.
2. Object literal expressions no longer require extra parentheses.
3. DataWeave multi-value selector syntax (`.*`, `.*field`) works.
4. Implicit lambda parameters (`$`, `$$`) work.
5. `reduce` supports DataWeave-style initial/default accumulator values.
6. Nested `mapObject` in object literals works.
7. `mapObject` supports object-return style and `[key, value]` pair style.
8. Object-lambda order defaults to DataWeave semantics, with explicit legacy compatibility for `(key, value)` / `(k, v)`.

Representative coverage lives in:
- `test/features/array_indexing.feature`
- `test/features/object_values.feature`
- `test/features/implicit_lambda.feature`
- `test/features/reduce_operator.feature`
- `test/features/object_functions.feature`
- `test/features/pluck_function.feature`

## Known Remaining Difference

- Header directive token remains `%im 0.1` instead of `%dw 2.0`.

## Policy

- Compatibility is default behavior.
- `--dataweave-compat` (or similar mode flags) is not the direction for this project.
- Any newly discovered incompatibility should be filed as a normal beads issue and prioritized for default behavior alignment.

## Tracking

To inspect open compatibility work:
```bash
bd ready
```

To inspect compatibility-related issues:
```bash
bd show <issue-id>
```

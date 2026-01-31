# Error Handling Guidelines

InfoMunge uses a unified error type in `internal/errors` so user-facing failures are consistent
and include context like error type and position.

## Principles

- Prefer `internal/errors` for any error that can reach the CLI or HTTP server.
- Use positional helpers for evaluator and parser errors when a `token.Pos` is available.
- Wrap underlying errors at package boundaries; avoid raw `fmt.Errorf`/`errors.New` in
  user-facing paths.
- Internal helper functions can return raw errors, but they must be wrapped before crossing
  package boundaries.

## Error Types

- Parse errors: syntax or header parsing issues in scripts (`ParseError`, `ParsePositional`).
- Evaluation errors: semantic/runtime evaluation failures (`EvalError`, `EvalPositional`).
- Validation errors: invalid input data, unsupported formats, or invalid arguments (`ValidationError`).
- I/O errors: file, stdin/stdout, or external read/write failures (`IOError`).
- Internal errors: unexpected invariants or bugs (`InternalError`).
- User errors: explicit user-thrown failures (`UserError`).

## Recommended Helpers

- `ParseError` / `ParseErrorf`, `ParsePositional` / `ParsePositionalf`
- `EvalError` / `EvalErrorf`, `EvalPositional` / `EvalPositionalf`
- `ValidationError` / `ValidationErrorf`
- `IOError` / `IOErrorf`
- `WrapParse*`, `WrapEval*`, `WrapValidation*`, `WrapIO*`, `WrapInternal*`

## Examples

```go
// Validation error with a root cause included in the message
return errors.ValidationErrorf("JSON parse error: %v", err)

// I/O error with context
return errors.WrapIOf(err, "error reading script file: %v", err)

// Evaluation error with position
return errors.EvalPositional("array index out of bounds", expr.Pos(), fset)
```

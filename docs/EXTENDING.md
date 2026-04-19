# Extending InfoMunge

Quick guide for adding features without spelunking.

## Add a New Operator

1. Add a transformer in `internal/preprocessor/transformers_*.go`.
2. Wire it into the pipeline in `internal/preprocessor/pipeline.go`.
3. Add tests in `internal/preprocessor/transformers_test.go` or a new test file.

Notes:
- Ensure the transform updates the mapping for error positions.
- Prefer minimal, targeted rewrites to keep parsing stable.

## Add a New Builtin Function

1. Implement in `internal/evaluator/builtins_*.go`.
2. Register in the builtin registry in `internal/evaluator/modular_registry.go`.
3. Add unit tests in `internal/evaluator/*_test.go`.

## Add a New Format

1. Add reader/writer in `pkg/formats/*`.
2. Register in `pkg/formats/registry.go`.
3. Cover parsing/serialization in `pkg/formats/*_test.go`.

## Module Loading Behavior

Update `internal/runner/module_loader.go` to change module search paths or import
resolution.

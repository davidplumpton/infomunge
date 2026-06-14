# Architecture

This document is a high-level walkthrough of the InfoMunge execution pipeline and
the core packages involved.

## Pipeline Overview

```
CLI -> Inputs -> Header -> Preprocess -> Evaluate -> Format output
```

1. CLI and arguments
   - `cmd/infomunge/main.go`
   - `internal/cli/app.go`
2. Input parsing
   - `internal/io/input.go`
3. Header parsing, context assembly, module loading
   - `internal/runner/runner.go`
   - `internal/runner/module_loader.go`
4. Preprocessing / syntax transforms
   - `internal/preprocessor/*`
5. Evaluation (AST walking, builtins, lazy values)
   - `internal/evaluator/*`
6. Read/write formatting
   - `pkg/formats/*`

## Header Handling

The header block is parsed before evaluation to build the context. This includes
output format directives, namespace declarations, function/type definitions, and
module imports. Header `input` directives are accepted as compatibility
metadata, but input bytes are parsed by CLI, server, or embedding adapter layers
before the runner receives an evaluation context. See `internal/runner/runner.go`.

## Preprocessor and Source Mapping

The preprocessor rewrites DataWeave-like syntax into a Go-parseable expression.
It also keeps a mapping from transformed indices back to original source for
error positioning. See `internal/preprocessor/*` and
`internal/evaluator/evaluator.go`.

## Evaluation

The evaluator parses the rewritten expression into a Go AST and visits nodes to
produce values. Builtins and core types live in `internal/evaluator/*`. Lazy
values are resolved after evaluation if requested.

## Error Reporting

Errors are wrapped with unified context that points to original source line and
column. See `internal/errors/*` and `internal/evaluator/evaluator.go`.

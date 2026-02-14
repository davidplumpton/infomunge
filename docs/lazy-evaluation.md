# Lazy Evaluation in InfoMunge

InfoMunge supports lazy evaluation for efficient processing of large datasets.

## Overview

Lazy evaluation defers computation until the result is actually needed. This is particularly useful for large data streams where eager evaluation would consume excessive memory.

This document reflects the current implementation in the codebase.

## Key Concepts

- **LazyValue**: A wrapper that defers evaluation of an expression.
- **Stream**: A channel-based sequence of values that can be processed lazily.
- **Pipeline Operations**: Map, filter, and reduce operations that work on streams without materializing intermediate results.

## Built-in Functions

### lazy_eval(expr)
Creates a lazy value that evaluates the expression when forced.

### force_eval(lazyValue)
Forces evaluation of a lazy value.

### __toStream(array)
Converts an array to a lazy stream.

### __lazyMap(stream, lambda)
Applies a mapping function lazily to each element in the stream.

### __lazyFilter(stream, lambda)
Filters elements in the stream using a predicate.

### __lazyReduce(stream, lambda, initial)
Reduces the stream to a single value.

## CLI Usage

The `--lazy` flag is currently unsupported and returns a validation error.
Use lazy builtins directly in scripts instead.

```bash
infomunge -i payload data.json "%im 0.1
output application/json
---
__lazyFilter(__lazyMap(__toStream(payload), (x) -> x.field), (x) -> x > 10)"
```

Example of unsupported flag behavior:

```bash
infomunge --lazy -f script.im
# error: --lazy is currently unsupported; use lazy_eval/force_eval and __toStream/__lazyMap/__lazyFilter/__lazyReduce builtins directly
```

## Examples

```infomunge
// Create a lazy stream from an array
stream = __toStream([1,2,3,4,5])

// Apply lazy operations
result = stream |> __lazyMap((x) -> x * x) |> __lazyFilter((x) -> x > 10)

// Force evaluation
output application/json --- force_eval(result)
```

This will output `[16,25]` without creating intermediate arrays.

## Current Semantics

- `lazy_eval(expr)` returns a `LazyValue` that caches the first evaluation result.
- `force_eval(lazyValue)` forces evaluation and returns the underlying value.
- `__toStream(array)` returns a lazy stream backed by a Go channel.
- `__lazyMap` and `__lazyFilter` return lazy streams; errors are delivered via a side-channel and surface when results are consumed.
- `__lazyReduce` consumes the input stream and returns a single value (not a stream).

## Execution Details

- Lazy streams use Go channels with backpressure; consumers control the rate.
- Cancellation is supported through the evaluation context's Go `context.Context`.
- Output formatting materializes streams into arrays before writing results, so stream output is not truly incremental.

## Limitations and Gaps

- The `--lazy` flag is intentionally rejected while lazy mode semantics are not implemented end-to-end.
- There is no dependency graph or cycle detection for lazy expressions beyond the builtin evaluation.

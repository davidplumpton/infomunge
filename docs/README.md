# Documentation

This directory contains current user and contributor guidance for InfoMunge.
Completed plans, point-in-time audits, and code-review snapshots are removed once
their findings are implemented or tracked in beads; use `br show <id>` and
`jj log` when historical implementation context is needed.

## Using InfoMunge

- [Supported formats](FORMATS.md) lists MIME types, extensions, direction, and
  fidelity constraints.
- [Date formats](DATE_FORMATS.md) documents the supported date-formatting token
  subset.
- [DataWeave cookbook equivalents](DATAWEAVE_COOKBOOK_EQUIVALENTS.md) collects
  runnable transformation examples.
- [While loops](WHILE_LOOPS.md) documents loop syntax, control flow, results,
  and timeout behavior.
- [Lazy evaluation](lazy-evaluation.md) describes the current lazy builtins and
  their limitations.

## Developing InfoMunge

- [Architecture](ARCHITECTURE.md) introduces the execution pipeline and package
  boundaries.
- [Extending InfoMunge](EXTENDING.md) explains how to add operators, builtins,
  and formats.
- [Testing](TESTING.md) is the authoritative guide for unit, Godog, coverage,
  property, fuzz, mutation, and differential testing.
- [Error handling](ERROR_HANDLING.md) defines user-facing error conventions.

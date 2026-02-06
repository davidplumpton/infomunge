# Quick DataWeave to InfoMunge Migration Guide

## Goal

InfoMunge aims to be fully DataWeave-compatible by default.

## TL;DR

In most cases, migration is just:
1. Change `%dw 2.0` to `%im 0.1`
2. Run the script

## Current Differences

| Area | DataWeave | InfoMunge |
|------|-----------|-----------|
| Header | `%dw 2.0` | `%im 0.1` |

## What Already Works

The following DataWeave-style patterns are supported:
- Dot access on arrays (`items.name`)
- Multi-value selectors (`.*`, `.*field`)
- Implicit lambda params (`$`, `$$`)
- `mapObject`, `filterObject`, `pluck` object-lambda behavior (DataWeave-first semantics)
- `reduce` with default/initial accumulator values
- Nested `mapObject` inside object literals
- Object-return and pair-return `mapObject` styles

## Example

DataWeave:
```dataweave
%dw 2.0
output application/json
---
payload.users map (u, i) -> {
  id: u.id,
  label: u.name ++ "-" ++ (i as String)
}
```

InfoMunge:
```im
%im 0.1
output application/json
---
payload.users map (u, i) -> {
  id: u.id,
  label: u.name ++ "-" ++ (i as String)
}
```

## Migration Checklist

- [ ] Change `%dw 2.0` to `%im 0.1`
- [ ] Run feature/script tests
- [ ] File an issue for any behavior mismatch (do not gate via compatibility mode)

## Related Docs

- `docs/DATAWEAVE_COMPATIBILITY_TASKS.md`
- `docs/DATAWEAVE_COOKBOOK_EQUIVALENTS.md`
- `docs/DATAWEAVE_CONVERSION_SUMMARY.md`

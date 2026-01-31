# DataWeave Compatibility Enhancement Tasks

## Overview

This document tracks enhancements to improve InfoMunge syntax compatibility with DataWeave. All tasks are organized under Epic **infomunge-vv1**.

## Epic: InfoMunge Syntax Improvements - DataWeave Compatibility

**ID:** infomunge-vv1  
**Priority:** P1  
**Status:** Open

Epic for improving InfoMunge syntax to be more compatible with DataWeave patterns and reduce friction for users migrating from DataWeave.

---

## High Priority Tasks (P2)

### 1. Allow direct field selection on arrays without recursive descent
**ID:** infomunge-99x  
**Priority:** P2  
**Status:** Open

**Problem:**
- DataWeave allows `.name` on arrays to extract field from all items
- Currently requires `..name` (recursive descent) or `map()`
- Users migrating from DataWeave expect the direct syntax to work

**Solution:**
- Support both syntaxes: `.name` on arrays AND `..name`
- Treat `.name` on arrays as equivalent to `map(x -> x.name)`

**Example:**
```im
// Should work like this:
[{"name": "Alice"}, {"name": "Bob"}].name
// Instead of requiring:
[{"name": "Alice"}, {"name": "Bob"}]..name
```

---

### 2. Support implicit parentheses in object literal expressions
**ID:** infomunge-46x  
**Priority:** P2  
**Status:** Open

**Problem:**
- Currently requires: `{key: (val1 ++ val2)}`
- DataWeave allows: `{key: val1 ++ val2}`
- Extra parentheses add visual clutter and confusion

**Solution:**
- Improve parser to handle complex expressions in object literals without explicit parentheses
- Make parentheses optional while maintaining clarity

**Example:**
```im
// Should work like this:
{name: firstName ++ " " ++ lastName}
// Instead of requiring:
{name: (firstName ++ " " ++ lastName)}
```

---

### 3. Support both mapObject parameter orders for DataWeave compatibility
**ID:** infomunge-7sh  
**Priority:** P2  
**Status:** Open

**Problem:**
- DataWeave uses `mapObject(value, key)` order
- InfoMunge uses `(key, value)` (reversed)
- Requires significant mental overhead when switching between languages

**Solution:**
- Option A: Support both parameter orders (check function arity)
- Option B: Add configuration option to choose default order
- Option C: Accept both and infer from usage

**Example:**
```im
// DataWeave style (should work):
obj mapObject (value, key) -> [upper(key), value]
// InfoMunge style (currently required):
obj mapObject (key, value) -> [upper(key), value]
```

---

### 4. Support DataWeave multi-value selector syntax (.*)
**ID:** infomunge-sg5  
**Priority:** P2  
**Status:** Open

**Problem:**
- DataWeave uses `.*` to get all values at a level
- InfoMunge uses `..` (recursive descent)
- These have different semantics but `.*` is more intuitive

**Solution:**
- Implement `.*` selector for extracting all values at current level
- Keep `..` for recursive descent
- Document the difference clearly

**Example:**
```im
// DataWeave style (get all values at current level):
payload.users.*name
// Current InfoMunge workaround:
payload.users map (x) -> x.name
```

---

## Medium Priority Tasks (P3)

### 5. Support default lambda parameters ($ and $$)
**ID:** infomunge-10i  
**Priority:** P3  
**Status:** Open

**Problem:**
- DataWeave supports `$` for value and `$$` for index
- InfoMunge requires named parameters
- Reduces verbosity and familiarity for DataWeave users

**Solution:**
- Implement default parameter names: `$` and `$$`
- Optional feature that coexists with named parameters

**Example:**
```im
// DataWeave style:
[1, 2, 3] map $ * 2
// Current InfoMunge requirement:
[1, 2, 3] map (x) -> x * 2
```

---

### 6. Improve reduce() initial value handling
**ID:** infomunge-b3w  
**Priority:** P3  
**Status:** Open

**Problem:**
- DataWeave supports optional initial value: `reduce(acc, val = 0)`
- InfoMunge uses first element as initial value
- Makes simple operations less concise

**Solution:**
- Support optional initial value parameter
- Maintain backward compatibility (default: use first element)

**Example:**
```im
// DataWeave style with initial value:
payload reduce (acc, val = 0) -> acc + val
// Current InfoMunge requirement:
(payload map (x) -> x.value) reduce (acc, v) -> acc + v
```

---

### 7. Improve error messages for syntax differences
**ID:** infomunge-33v  
**Priority:** P3  
**Status:** Open

**Problem:**
- When encountering DataWeave syntax that doesn't work, errors are cryptic
- Users don't know what alternative syntax to use
- Migration from DataWeave is frustrating

**Solution:**
- Detect common DataWeave patterns and suggest alternatives
- Add context-aware error messages
- Include examples in error output

**Example:**
```
Error: array index must be an integer, got string
Hint: To extract fields from array items, use one of:
  - Recursive descent: items..name
  - Map transformation: items map (x) -> x.name
  - Multi-value selector: items.*name (when implemented)
```

---

### 8. Support nested mapObject in object literals
**ID:** infomunge-wwt  
**Priority:** P3  
**Status:** Open

**Problem:**
- Parsing fails with: `{book: item mapObject(k,v) -> [upper(k),v]}`
- Requires workarounds or refactoring
- Limits expressiveness and conciseness

**Solution:**
- Improve parser to handle nested operations in object constructors
- May require lookahead or restructuring parser rules

**Example:**
```im
// Should work like this:
payload map (item) -> {
  book: item mapObject (key, value) -> [upper(key), value]
}
```

---

### 9. Support optional mapObject return format (object vs array)
**ID:** infomunge-kbn  
**Priority:** P3  
**Status:** Open

**Problem:**
- DataWeave can create objects directly from mapObject
- InfoMunge requires `[key, value]` array pairs
- Less intuitive when you want object output

**Solution:**
- Detect if `[key, value]` pairs should be converted to object
- Support object literal syntax in mapObject results
- Consider: `mapObject((k, v) -> {(k): v})`

**Example:**
```im
// DataWeave style (creates object):
mapObject (value, key) -> {(upper(key)): value}
// Current InfoMunge (returns pairs):
mapObject (key, value) -> [upper(key), value]
// After conversion to object output
```

---

## Low Priority Tasks (P4)

### 10. Add DataWeave compatibility mode flag
**ID:** infomunge-1j2  
**Priority:** P4  
**Status:** Open

**Problem:**
- No unified way to enable DataWeave-compatible behavior
- Users must learn InfoMunge syntax even if familiar with DataWeave
- Could help adoption from DataWeave community

**Solution:**
- Add `--dataweave-compat` flag or similar
- Enables lenient parsing of DataWeave patterns
- Provides migration suggestions and helpful messages
- Could emit warnings for deprecated patterns

**Example:**
```bash
./infomunge --dataweave-compat -f script.dw
# Parses DataWeave syntax more leniently
# Suggests InfoMunge alternatives where needed
```

---

## Summary by Impact

### Highest Impact (Most Common Issues)
1. **Direct array field selection** (infomunge-99x) - Very frequent in transformations
2. **Implicit parentheses** (infomunge-46x) - Constant pain point
3. **mapObject parameter order** (infomunge-7sh) - Confusing for every mapObject usage
4. **Multi-value selectors** (infomunge-sg5) - Alternative to current recursive descent

### Medium Impact (Quality of Life)
5. **Error messages** (infomunge-33v) - Improves user experience significantly
6. **Default parameters** (infomunge-10i) - Reduces verbosity in common patterns
7. **Nested mapObject** (infomunge-wwt) - Enables more complex transformations

### Lower Impact (Nice to Have)
8. **Reduce initial value** (infomunge-b3w) - Useful but workarounds exist
9. **mapObject return format** (infomunge-kbn) - Can be worked around with conversion
10. **Compatibility mode** (infomunge-1j2) - Meta-feature, depends on other improvements

---

## Related Documentation

- **DATAWEAVE_COOKBOOK_EQUIVALENTS.md** - 20 cookbook examples with current workarounds
- **DATAWEAVE_CONVERSION_SUMMARY.md** - Complete syntax comparison and migration guide
- **CLAUDE.md** - Project conventions and workflow

---

## Tracking

All issues are linked to Epic **infomunge-vv1** via "blocks" relationship.

To view all related tasks:
```bash
bd show infomunge-vv1
```

To view specific task:
```bash
bd show infomunge-99x  # for example
```

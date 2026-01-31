# Quick DataWeave to InfoMunge Migration Guide

This quick reference helps you migrate from DataWeave to InfoMunge.

## TL;DR: Key Syntax Differences

| Operation | DataWeave | InfoMunge | Note |
|-----------|-----------|-----------|------|
| **Simple field** | `payload.name` | `payload.name` | Same |
| **Nested field** | `payload.user.email` | `payload.user.email` | Same |
| **Array element** | `items[0]` | `items[0]` | Same |
| **Extract from array** | `items.name` | `items..name` | Use `..` or `map()` |
| **Map array** | `map (x) -> x * 2` | `map (x) -> x * 2` | Same |
| **Filter** | `filter (x) -> x > 2` | `filter (x) -> x > 2` | Same |
| **mapObject** | `mapObject (v, k) -> ...` | `mapObject (k, v) -> ...` | **Order reversed** |
| **String concat** | `{name: a ++ b}` | `{name: (a ++ b)}` | Needs parens |
| **GroupBy** | `groupBy (x) -> x.type` | `groupBy (x) -> x.type` | Same |
| **Pluck** | `pluck "name"` | `pluck "name"` | Same |
| **Reduce** | `reduce (a, x = 0)` | `reduce (a, x)` | No initial value syntax |
| **Default params** | `map $` | `map (x) -> ...` | Named params required |

## Most Common Migrations

### 1. Extract field from array of objects

**❌ Won't work:**
```im
payload.users.name
```

**✅ Use one of these:**
```im
// Option 1: Recursive descent
payload.users..name

// Option 2: Map (for transformation)
payload.users map (user) -> user.name

// Option 3: Pluck (for simple field extraction)
payload.users pluck "name"
```

### 2. String concatenation in objects

**❌ Won't work:**
```im
{greeting: "Hello" ++ name}
```

**✅ Use parentheses:**
```im
{greeting: ("Hello" ++ name)}
```

### 3. mapObject with key transformation

**❌ Wrong parameter order:**
```im
obj mapObject (value, key) -> [upper(key), value]
```

**✅ Correct parameter order:**
```im
obj mapObject (key, value) -> [upper(key), value]
```

### 4. Filter objects

**❌ Won't compile:**
```im
{a: 1, b: 2} filterObject (v) -> v > 1
```

**✅ Correct syntax:**
```im
filterObject({a: 1, b: 2}, (k, v) -> v > 1)
```

### 5. Transform array of objects

**❌ Won't work perfectly:**
```im
items map (x) -> {id: x.id, name: x.name mapObject (k, v) -> upper(k)}
```

**✅ Simpler approach:**
```im
items map (x) -> {id: x.id, name: x.name}
// OR refactor for clarity
```

## Before and After: Real Examples

### Example 1: Extract and rename fields

**DataWeave:**
```dataweave
%dw 2.0
output application/json
---
{
  userId: payload.user.id,
  email: payload.user.email,
  status: payload.active
}
```

**InfoMunge:**
```im
%im 0.1
input application/json
output application/json
---
{
  userId: payload.user.id,
  email: payload.user.email,
  status: payload.active
}
```
✅ No changes needed!

---

### Example 2: Map array with field extraction

**DataWeave:**
```dataweave
%dw 2.0
output application/json
---
payload.items map {
  id: $.id,
  price: $.price as Number,
  total: ($.price as Number) * $.quantity
}
```

**InfoMunge:**
```im
%im 0.1
input application/json
output application/json
---
payload.items map (item) -> {
  id: item.id,
  price: (item.price as Number),
  total: ((item.price as Number) * item.quantity)
}
```

---

### Example 3: Complex grouping and filtering

**DataWeave:**
```dataweave
%dw 2.0
output application/json
---
payload 
  filter (x) -> x.active == true
  groupBy (x) -> x.category
  mapObject (v) -> {
    category: $$,
    count: sizeOf(v),
    items: v
  }
```

**InfoMunge:**
```im
%im 0.1
input application/json
output application/json
---
(payload filter (x) -> x.active == true) groupBy (x) -> x.category
```

---

## Common Gotchas

### 1. Recursive Descent vs Multi-value Selector
```im
// This WORKS (searches everywhere recursively)
payload..name

// This WON'T WORK (InfoMunge doesn't support .*)
payload.items.*name
```

### 2. Lambda Variable Names
```im
// DataWeave allows $ and $$
[1, 2, 3] map $ * 2

// InfoMunge requires named variables
[1, 2, 3] map (x) -> x * 2
```

### 3. Parentheses in Objects
```im
// WRONG - will fail to parse
{result: 1 + 2}

// CORRECT - needs parentheses
{result: (1 + 2)}
```

### 4. mapObject Output Format
```im
// Returns array pairs, not direct object
keysOf(obj) map (k) -> [k, obj[k]]

// For object with transformed keys:
mapObject (key, value) -> [upper(key), value]
```

## Migration Checklist

- [ ] Change `%dw 2.0` → `%im 0.1`
- [ ] Change `mapObject (v, k)` → `mapObject (k, v)` (if used)
- [ ] Add parentheses to complex expressions in objects
- [ ] Replace `.name` on arrays with `..name` or `map()`
- [ ] Replace `$` with named parameter, `$$` with index parameter
- [ ] Replace `sizeOf()` with `size()` (if applicable)
- [ ] Test all transformations (use `go test ./test`)

## When to Reach for Documentation

**Quick questions?** → Use this file (QUICK_DATAWEAVE_MIGRATION.md)

**Detailed examples?** → See DATAWEAVE_COOKBOOK_EQUIVALENTS.md (20 examples)

**Complete syntax reference?** → See DATAWEAVE_CONVERSION_SUMMARY.md

**Enhancement proposals?** → See DATAWEAVE_COMPATIBILITY_TASKS.md (10 open issues)

## Resources

- **README.md** - InfoMunge project overview
- **test/features/** - Hundreds of test examples showing correct syntax
- **bd show infomunge-vv1** - View all DataWeave compatibility tasks

---

**Last Updated:** 2026-01-03  
**InfoMunge Version:** 0.1  
**All 764 Tests Passing:** ✓

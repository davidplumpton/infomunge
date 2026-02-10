# Intensive Expression Testing

I want to develop a plan for intensive testing of complex expressions. I'm thinking of something like starting with a curated expression and data, (or possibly randomly created if it can be done well), and then randomly mutating it while validating invariant conditions.

Possibilities:

- Creating a grammar for the infomunge expression language
- Shrinking of expressions when an issue is found
- Comparison of output to dataweave cli

The goal is to greatly improve the robustness of the solution without having to manually provide a large number of expressions to evaluate in advance.

Possible solutions should make use of any relevant golang libraries.

---

## Proposed Approach

### Core Library: `rapid` (pgregory.net/rapid)

The recommended foundation is **rapid**, a mature property-based testing library for Go (v1.2.0, actively maintained). It provides:

- **Custom recursive generators** via `Deferred` + `OneOf` + `Custom` — exactly what's needed to generate expression trees from a grammar
- **Automatic shrinking** — when a failure is found, rapid automatically minimizes the failing expression to the smallest reproduction case, with no user code required
- **Native fuzz bridge** — `rapid.MakeFuzz` connects structured generators to Go's coverage-guided fuzzer (`go test -fuzz`), combining grammar-aware generation with coverage-guided mutation

Other options considered and rejected:
- **gopter** — unmaintained since 2020, verbose API, manual shrinking
- **Go native fuzzing alone** — only mutates raw bytes, will spend nearly all time on syntactically invalid input and never reach deep evaluator paths

### Architecture: Three Testing Layers

#### Layer 1: Grammar-Based Expression Generator

Build a recursive expression generator in Go that produces syntactically valid infomunge expressions. This is the foundation everything else builds on.

```
ExprGen = OneOf(
    LiteralGen        →  numbers, strings, booleans, null
    IdentGen          →  payload, x, y, item, $$
    BinaryOpGen       →  ExprGen op ExprGen   (op ∈ {+, -, *, /, **, %, ++, ==, !=, <, >, <=, >=, &&, ||, ~=})
    UnaryGen          →  !ExprGen, -ExprGen
    ArrayGen          →  [ExprGen, ...]
    ObjectGen         →  {key: ExprGen, ...}
    IndexGen          →  ExprGen[ExprGen], ExprGen.field
    CallGen           →  builtin(ExprGen, ...)
    CollectionOpGen   →  ExprGen map/filter/reduce/flatMap LambdaGen
    LambdaGen         →  (params) -> ExprGen
    ConditionalGen    →  if (ExprGen) ExprGen else ExprGen
    DefaultGen        →  ExprGen default ExprGen
    StringInterpGen   →  "text $(ExprGen) text"
    CaseGen           →  case ExprGen { pattern => ExprGen, ... }
    RangeIndexGen     →  ExprGen[ExprGen to ExprGen]
)
```

The generator would be parameterized by:
- **Max depth** — prevents unbounded recursion (rapid's `Deferred` naturally biases toward smaller trees, but an explicit depth limit provides a hard cap)
- **Feature subset** — start with arithmetic/comparison/arrays, incrementally add operators as confidence grows
- **Context shape** — what variables are in scope (payload structure, lambda parameters)

A key design choice: the generator operates at the **infomunge source level** (pre-preprocessing), not at the Go AST level. This means generated expressions exercise the full pipeline: preprocessor → parser → evaluator → formatter. This is where the most bugs live — in the interactions between these stages.

#### Layer 2: Property-Based Invariant Tests

With the generator in hand, define properties that must hold for all valid expressions:

**Universal properties (no oracle needed):**
- **No panics** — any generated expression should either produce a result or return an error, never panic. This alone will find many bugs.
- **Preprocessor idempotence** — `preprocess(preprocess(expr))` should equal `preprocess(expr)` for any expression that doesn't error
- **Determinism** — `eval(expr, ctx)` called twice with identical inputs produces identical output
- **Type consistency** — `typeOf(eval(expr))` should return a valid type string

**Algebraic properties:**
- **Arithmetic identities** — `x + 0 == x`, `x * 1 == x`, `x * 0 == 0`, `x ** 1 == x`
- **Commutativity** — `a + b == b + a`, `a * b == b * a` (for numeric a, b)
- **Associativity** — `(a + b) + c == a + (b + c)` (modulo floating-point)
- **Default operator** — `x default x == x`, `x default y == x` when x is non-null, `null default y == y`
- **Boolean logic** — `!(!x) == x`, `x && true == x`, `x || false == x`
- **Collection identity** — `arr map (x) -> x == arr`, `arr filter (x) -> true == arr`
- **Concatenation** — `[] ++ arr == arr`, `arr ++ [] == arr`
- **String operations** — `upper(lower(s))` preserves length, `sizeOf(s) >= 0`
- **Round-trip** — `read(write(val, "json"), "application/json") == val` for JSON-safe values

**Structural properties:**
- **sizeOf consistency** — `sizeOf(arr map f) == sizeOf(arr)` for any non-erroring f
- **filter subset** — `sizeOf(arr filter f) <= sizeOf(arr)`
- **flatten depth** — `sizeOf(flatten(arr)) >= sizeOf(arr)` for arrays of arrays
- **keys/values** — `sizeOf(keys(obj)) == sizeOf(values(obj))`

#### Layer 3: Differential Testing Against DataWeave CLI

For expressions within the overlap of infomunge and DataWeave syntax, run both and compare:

```
infomunge eval(expr, ctx)  ==  dw eval(expr, ctx)
```

Implementation approach:
1. The expression generator tags each expression with a compatibility flag (infomunge-only features like `while` loops excluded)
2. A test harness wraps the DataWeave CLI (`dw run -i payload ...`) and captures output
3. Both outputs are parsed and compared structurally (not string-equal, since formatting may differ)
4. Differences are logged with the generated expression for manual triage

This is the most powerful oracle but also the most constrained — it only covers the syntax subset both implementations share. It's most useful for core operators, collection functions, and type coercion behavior.

### Implementation Plan

**Phase 1: Foundation (no new dependencies yet)**

Write the expression generator as a standalone Go package `internal/testing/exprgen/` using rapid. Start with a minimal grammar subset:
- Literals: integers, floats, strings, booleans, null
- Binary operators: `+`, `-`, `*`, `/`, `==`, `!=`, `<`, `>`, `&&`, `||`
- Unary: `!`, `-`
- Arrays: `[expr, ...]`
- Parentheses: `(expr)`
- Variable references: `payload`, literal field access

Write one property test: "no panics when evaluating any generated expression with a sample payload". This single test, run with `rapid.Check` for thousands of iterations, will likely find issues immediately.

**Phase 2: Expand Grammar + Add Algebraic Properties**

Incrementally add expression forms to the generator:
- Object literals, dot access, indexing
- Collection operators: `map`, `filter`, `reduce`, `flatMap`
- Lambda expressions (both arrow and implicit `$`)
- `default`, `if/else`, string interpolation
- Builtin function calls (subset: `sizeOf`, `typeOf`, `upper`, `lower`, `flatten`, `keys`, `values`, etc.)

Add algebraic property tests for each operator/builtin as it's added to the generator.

**Phase 3: Coverage-Guided Fuzzing**

Bridge the generator to Go's native fuzzer via `rapid.MakeFuzz`:

```go
func FuzzExprEval(f *testing.F) {
    f.Fuzz(rapid.MakeFuzz(func(t *rapid.T) {
        expr := exprgen.Expression(3).Draw(t, "expr")  // depth 3
        ctx := exprgen.SampleContext().Draw(t, "ctx")
        result, err := runner.Eval(expr, ctx)
        // no panics is implicit; check other properties
        if err == nil {
            // result should be serializable
            _, serErr := json.Marshal(result)
            if serErr != nil {
                t.Fatalf("non-serializable result from %q: %v", expr, serErr)
            }
        }
    }))
}
```

Run with `go test -fuzz=FuzzExprEval -fuzztime=10m` to let coverage guidance discover interesting expression shapes that pure random generation misses.

**Phase 4: DataWeave Differential Testing**

Build a `dw` harness that shells out to the DataWeave CLI. Generate compatible expressions (exclude infomunge-only features), run both, compare. This phase is optional and depends on having `dw` available in the test environment.

**Phase 5: Mutation-Based Testing (Stretch Goal)**

Take expressions from the existing cucumber test suite (1600+ scenarios) and mutate them:
- Swap operators (`+` → `-`, `map` → `filter`)
- Change literals (`1` → `0`, `"a"` → `""`, `true` → `false`)
- Remove/add arguments to function calls
- Nest expressions deeper
- Swap subexpressions

This leverages the existing corpus of known-good expressions to explore the neighborhood of tested behavior. Mutations of passing tests should either still pass (equivalent mutation) or fail gracefully (error, not panic).

### Shrinking Strategy

Rapid handles shrinking automatically. When a property violation is found:
1. Rapid records the random byte stream that produced the failing expression
2. It systematically tries smaller byte streams that still trigger the failure
3. The output is a minimal expression like `payload[null]` instead of a 200-character monster

For the DataWeave differential tests, when a discrepancy is found, the expression can be further minimized by:
1. Using rapid's built-in shrinking for generated expressions
2. For mutated expressions, bisecting which mutation caused the discrepancy

### File Organization

```
internal/testing/
  exprgen/
    generator.go          — core expression generator (recursive, parameterized)
    generator_test.go     — self-tests for the generator (valid syntax, depth bounds)
    literals.go           — literal value generators
    operators.go          — operator expression generators
    collections.go        — collection operator generators (map, filter, etc.)
    context.go            — test context/payload generators
  properties/
    properties_test.go    — algebraic property tests
    nopanic_test.go       — universal no-panic property
    roundtrip_test.go     — serialization round-trip properties
  differential/
    dw_harness.go         — DataWeave CLI wrapper
    differential_test.go  — comparison tests
  mutation/
    mutator.go            — expression mutation strategies
    mutation_test.go      — mutation-based tests using cucumber corpus
```

### Expected Impact

- **Phase 1 alone** (no-panic property with minimal grammar) will likely surface preprocessor/evaluator edge cases within minutes of running
- **Phase 2** (algebraic properties) catches semantic bugs: wrong operator precedence, incorrect type coercion, broken identity laws
- **Phase 3** (coverage-guided) finds deep corner cases that pure random generation misses — the fuzzer learns which byte patterns lead to new code paths
- **Phase 4** (differential) catches behavioral divergences from DataWeave specification

The approach is incremental — each phase provides standalone value and the generator grows organically as more language features are added to it.

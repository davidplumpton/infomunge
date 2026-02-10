# Intensive Expression Testing

## Goal

Greatly improve the robustness of infomunge expression evaluation without manually authoring a large number of test expressions. Use grammar-aware generation, property-based invariants, mutation of existing test corpus, differential comparison against DataWeave, and automatic shrinking to find bugs and produce minimal reproducible cases that feed back into the cucumber regression suite.

## Core Library: `rapid` (pgregory.net/rapid)

The foundation is **rapid** (v1.2.0+, actively maintained). It provides:

- **Custom recursive generators** via `Deferred` + `OneOf` + `Custom` — exactly what's needed to generate expression trees from a grammar
- **Automatic shrinking** — when a failure is found, rapid minimizes the failing expression to the smallest reproduction case with no user code required
- **Native fuzz bridge** — `rapid.MakeFuzz` connects structured generators to Go's native fuzzer (`go test -fuzz`), combining grammar-aware generation with coverage-guided exploration

Other options considered and rejected:
- **gopter** — unmaintained since 2020, verbose API, requires manual shrinker implementations per AST node type
- **Go native fuzzing alone** — only mutates raw bytes; will spend nearly all time on syntactically invalid input and never reach deep evaluator paths

---

## Architecture: Four Testing Layers

### Layer 1: Grammar-Based Expression Generator

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

The generator is parameterized by:
- **Max depth** — prevents unbounded recursion (rapid's `Deferred` naturally biases toward smaller trees, but an explicit depth limit provides a hard cap)
- **Feature subset** — start with arithmetic/comparison/arrays, incrementally add operators as confidence grows
- **Context shape** — what variables are in scope (payload structure, lambda parameters)

A key design choice: the generator operates at the **infomunge source level** (pre-preprocessing), not at the Go AST level. This means generated expressions exercise the full pipeline: preprocessor → parser → evaluator → formatter. This is where the most bugs live — in the interactions between these stages.

To avoid wasting test budget on uninteresting parse failures, the generator should apply lightweight validity guards:
- Balance brackets and parentheses by construction
- Only reference variables that are in scope
- Only use operators with compatible arity
- Skip the generated case early (don't count it toward the test budget) if the preprocessor returns a syntax error

### Layer 2: Property-Based Invariant Tests

With the generator in hand, define properties that must hold for all valid expressions:

**Universal properties (no oracle needed):**
- **No panics** — any generated expression should either produce a result or return an error, never panic. This alone will find many bugs.
- **Determinism** — `eval(expr, ctx)` called twice with identical inputs produces identical output
- **Type consistency** — `typeOf(eval(expr))` should return a valid type string
- **Stability** — same input + expression should produce byte-identical output across repeated runs and across serialization

**Algebraic properties:**
- **Arithmetic identities** — `x + 0 == x`, `x * 1 == x`, `x * 0 == 0`, `x ** 1 == x`
- **Commutativity** — `a + b == b + a`, `a * b == b * a` (for numeric a, b)
- **Associativity** — `(a + b) + c == a + (b + c)` (modulo floating-point tolerance)
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

### Layer 3: Mutation-Based Testing from Cucumber Corpus

Take expressions from the existing cucumber test suite and systematically mutate them. This leverages the large body of known-good expressions (1600+ scenarios) to explore the neighborhood of tested behavior.

**Mutation operators:**
- **Operator replacement** — `+` → `-`, `and` → `or`, `map` → `filter`, `==` → `!=`, etc.
- **Literal perturbation** — `1` → `0`, `"a"` → `""`, `true` → `false`, `null` → `0`
- **Subtree swap/clone/delete** — exchange subexpressions between two corpus entries
- **Argument manipulation** — remove/add/reorder arguments to function calls
- **Nesting** — wrap a subexpression in an extra layer (`x` → `[x][0]`, `x` → `(x)`)
- **Parenthesis insertion/removal** — stress precedence handling
- **Function-name substitution** — swap among compatible builtins (same arity)

**Invariant for mutated expressions:** mutations of passing tests should either still pass (equivalent mutation) or fail gracefully (return an error, never panic).

**Corpus extraction:** parse `.feature` files under `test/features/` to extract expression/input/expected-output triples. This is a one-time setup that can be re-run as the test suite grows.

### Layer 4: Differential Testing Against DataWeave CLI

For expressions within the overlap of infomunge and DataWeave syntax, run both and compare:

```
infomunge eval(expr, ctx)  ==  dw eval(expr, ctx)
```

Implementation approach:
1. The expression generator tags each expression with a compatibility flag (infomunge-only features excluded)
2. A test harness wraps the DataWeave CLI (`dw run -i payload ...`) and captures output
3. Both outputs are parsed and compared structurally (not string-equal, since formatting may differ)
4. Normalization rules handle known divergences: numeric precision, null/missing conventions, key ordering
5. Differences are logged with the generated expression for manual triage

This is the most powerful oracle but also the most constrained — it only covers the syntax subset both implementations share. It is optional and depends on `dw` being available in the test environment.

---

## Failure Workflow: From Discovery to Regression Test

When a property violation or differential mismatch is found:

### 1. Automatic Shrinking

Rapid handles shrinking automatically:
1. Rapid records the random byte stream that produced the failing expression
2. It systematically tries smaller byte streams that still trigger the failure
3. The output is a minimal expression (e.g., `payload[null]` instead of a 200-character monster)

For mutated expressions, shrinking bisects which mutation caused the failure.

### 2. Failure Artifact Capture

Save each unique failure as a structured artifact under `tmp/intensive-testing/failures/`:

```
tmp/intensive-testing/failures/
  001_nopanic_2026-02-10T14:30:00.json
  002_algebraic_2026-02-10T14:31:12.json
  ...
```

Each artifact records:
- The minimized expression (post-shrinking)
- The input/payload that triggered it
- The property that was violated
- The actual vs expected result (or panic stack trace)
- The original (pre-shrink) expression for context
- The random seed for reproducibility

### 3. Cucumber Scenario Auto-Generation

Auto-generate a candidate `.feature` scenario skeleton from each failure artifact:

```gherkin
# Auto-generated from intensive testing failure 001
# Property: no-panic
# Shrunk from: (payload.items map (x) -> x.name)[null + 1]
# Original seed: 0xABCD1234
Scenario: Expression does not panic on null index arithmetic
  Given input payload is
    """
    {"items": [{"name": "a"}]}
    """
  When infomunge processes
    """
    %im 0.1
    output application/json
    ---
    payload[null]
    """
  Then the error should contain "..."
```

Generated scenarios are written to `tmp/intensive-testing/candidates/` for human review. Approved scenarios are promoted to `test/features/` as permanent regression tests.

### 4. Deduplication

Track a fingerprint (hash of minimized expression + violated property) to avoid re-reporting the same underlying bug across multiple generated inputs.

---

## Implementation Plan

### Phase 1: Foundation

Write the expression generator as a standalone Go package `internal/testing/exprgen/` using rapid. Start with a minimal grammar subset:
- Literals: integers, floats, strings, booleans, null
- Binary operators: `+`, `-`, `*`, `/`, `==`, `!=`, `<`, `>`, `&&`, `||`
- Unary: `!`, `-`
- Arrays: `[expr, ...]`
- Parentheses: `(expr)`
- Variable references: `payload`, literal field access

Write one property test: "no panics when evaluating any generated expression with a sample payload". This single test, run with `rapid.Check` for thousands of iterations, will likely find issues immediately.

Add the failure artifact capture and candidate scenario generator so that any bug found from day one produces a promotable cucumber skeleton.

**Deliverables:** generator package, no-panic property test, failure artifact writer, scenario skeleton generator.

### Phase 2: Expand Grammar + Add Algebraic Properties

Incrementally add expression forms to the generator:
- Object literals, dot access, indexing
- Collection operators: `map`, `filter`, `reduce`, `flatMap`
- Lambda expressions (both arrow and implicit `$`)
- `default`, `if/else`, string interpolation
- Builtin function calls (subset: `sizeOf`, `typeOf`, `upper`, `lower`, `flatten`, `keys`, `values`, etc.)

Add algebraic property tests for each operator/builtin as it's added to the generator.

**Deliverables:** expanded generator, algebraic + structural property tests.

### Phase 3: Coverage-Guided Fuzzing

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

Note: `rapid.MakeFuzz` provides structured generation within Go's fuzz infrastructure. The coverage guidance comes from Go's fuzzer driving the byte stream that rapid interprets; the structured generation ensures most of that exploration hits syntactically valid expressions rather than random bytes.

**Deliverables:** fuzz targets, corpus persistence, CI integration for short fuzz runs.

### Phase 4: Mutation-Based Testing from Cucumber Corpus

Build the corpus extractor and mutation engine:
1. Parse `.feature` files to extract expression/input pairs
2. Implement mutation operators (operator swap, literal perturbation, subtree manipulation, etc.)
3. Run mutated expressions through the no-panic and determinism oracles
4. Feed failures into the same artifact/scenario pipeline

This phase leverages the existing 1600+ scenarios as seeds — it's high value because these expressions already cover known-interesting corners of the language.

**Deliverables:** corpus extractor, mutation engine, mutation-based property tests.

### Phase 5: DataWeave Differential Testing

Build a `dw` harness that shells out to the DataWeave CLI. Generate compatible expressions (exclude infomunge-only features), run both, compare. This phase is optional and depends on having `dw` available in the test environment.

**Deliverables:** DataWeave harness, normalization rules, differential comparison tests.

### Phase 6: CI and Scaling

- Add a short-budget property run in CI (fixed seed + N cases, completes in under 60 seconds)
- Add a longer soak run for local/nightly use with corpus persistence
- Add a mutation run that processes a random subset of cucumber corpus per CI run
- Persist interesting fuzz corpus entries across runs

**Deliverables:** CI configuration, soak-run scripts, corpus management.

---

## File Organization

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
  mutation/
    extractor.go          — cucumber corpus extraction
    mutator.go            — expression mutation strategies
    mutation_test.go      — mutation-based tests using cucumber corpus
  differential/
    dw_harness.go         — DataWeave CLI wrapper
    differential_test.go  — comparison tests
  failures/
    artifact.go           — failure artifact capture and deduplication
    scenario_gen.go       — cucumber scenario skeleton generator
tmp/intensive-testing/
  failures/               — failure artifacts (JSON, not committed)
  candidates/             — generated .feature skeletons for human review
```

---

## Success Metrics

Track these to evaluate whether the investment is paying off:

| Metric | Target | Measured How |
|--------|--------|-------------|
| **Unique failures per 10k generated cases** | Decreasing over time as bugs are fixed | Count distinct fingerprints per run |
| **Shrink ratio** | Minimized expression ≤ 20% of original size on average | Compare pre/post shrink lengths |
| **Time-to-minimal-repro** | Under 30 seconds for any single failure | Wall-clock from detection to shrunk artifact |
| **Promoted cucumber regressions** | At least 1 per phase of development | Count scenarios moved from `candidates/` to `test/features/` |
| **Coverage deltas** | Measurable increase in preprocessor + evaluator line coverage | `go test -coverprofile` before/after |
| **Panic rate** | Zero panics on any generated expression | Count panics per run; target is 0 |
| **Mutation kill rate** | Mutations that change behavior are caught by existing tests | Ratio of detected vs undetected non-equivalent mutations |

Review metrics after each phase to decide whether to continue investing or redirect effort.

---

## Expected Impact

- **Phase 1 alone** (no-panic property with minimal grammar) will likely surface preprocessor/evaluator edge cases within minutes of running
- **Phase 2** (algebraic properties) catches semantic bugs: wrong operator precedence, incorrect type coercion, broken identity laws
- **Phase 3** (coverage-guided) finds deep corner cases that pure random generation misses — the fuzzer learns which byte patterns lead to new code paths
- **Phase 4** (mutation) explores the neighborhood of already-tested behavior using the cucumber corpus as seeds — high ROI because the seeds are already known to be interesting
- **Phase 5** (differential) catches behavioral divergences from DataWeave specification

The approach is incremental — each phase provides standalone value and the generator grows organically as more language features are added to it. The failure-to-regression pipeline ensures every discovered bug becomes a permanent cucumber test, creating a virtuous cycle between random exploration and deterministic regression coverage.

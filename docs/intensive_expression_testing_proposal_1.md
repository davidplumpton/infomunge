# Intensive Expression Testing

I want to develop a plan for intensive testing of complex expressions. I'm thinking of something like starting with a curated expression and data, (or possibly randomly created if it can be done well), and then randomly mutating it while validating invariant conditions.

Possibilities:

- Creating a grammar for the infomunge expression language
- Shrinking of expressions when an issue is found
- Comparison of output to dataweave cli

The goal is to greatly improve the robustness of the solution without having to manually provide a large number of expressions to evaluate in advance.

Possible solutions should make use of any relevant golang libraries.

## Proposed Approach

Use a hybrid strategy that combines:

1. Property-based generation of valid expression+input pairs
2. Mutation-based exploration seeded from curated "interesting" expressions
3. Differential and metamorphic oracles to detect incorrect behavior
4. Automatic shrinking/minimization for fast debugging

This gives broad state-space coverage while keeping failures explainable.

### 1) Build a Typed Expression Generator

Create a generator that emits:

- Script header + expression body (`%im 0.1 ... --- <expr>`)
- Inputs (`payload`, and optional named inputs)
- A lightweight type environment to keep generation mostly valid

Start with a bounded AST depth (for example, max depth 4-6) and weighted node selection so common operators/functions are exercised most, with periodic deep/complex cases.

Recommended libraries:

- `github.com/leanovate/gopter` for property-based testing
- `testing/quick` for simple deterministic fuzz loops when properties are straightforward

Implementation note:

- Generate AST first, then pretty-print to source. This avoids syntax errors dominating test time.

### 2) Add Mutation Engine on Top of Curated Seeds

Maintain a seed corpus of handpicked expressions that are historically tricky:

- Nested lambda/map/filter chains
- Mixed object/array access patterns
- Grouping/precedence-heavy expressions
- Null and missing-field paths
- Numeric edge cases (zero, negatives, large/small magnitudes)

Define mutation operators such as:

- Operator replacement (`+` -> `-`, `and` -> `or`, etc.)
- Subtree swap/clone/delete
- Constant perturbation (numbers, strings, booleans, null)
- Function-name substitution among compatible builtins
- Parenthesis insertion/removal

Run multiple mutation rounds with validity guards (skip clearly invalid forms early).

### 3) Use Multiple Oracles

Not every expression can be checked with a single expected value, so combine:

1. Differential oracle:
   - Execute expression in InfoMunge and DataWeave CLI when supported by both runtimes.
   - Compare normalized outputs (JSON canonicalization, numeric tolerance rules, null/missing conventions where mapped).
2. Metamorphic oracle:
   - Check invariant transformations even without external oracle.
   - Examples:
     - Idempotence for stable transforms: `f(f(x)) == f(x)` where applicable
     - Commutativity for safe operators: `a + b == b + a` (numeric-only cases)
     - Filter subset relation: `filter(p, xs)` is subset of `xs`
3. Stability oracle:
   - Same input + expression should produce byte-identical output and no panic across repeated runs.

### 4) Add Failure Shrinking/Minimization

When a failure appears, automatically reduce:

- AST size (remove subtrees, simplify literals)
- Input size/shape (drop fields/elements)
- Header/options (remove non-essential directives)

Goal: produce a minimal reproducible case that can be promoted directly into a cucumber scenario.

Recommended libraries:

- `github.com/leanovate/gopter` shrinkers (custom shrinkers for AST nodes)
- Optional custom delta-debugger for source-level minimization if AST shrinking stalls

### 5) Integrate with Existing Cucumber Workflow

Flow for every discovered bug:

1. Save failing case artifact under `tmp/intensive-testing/failures/`
2. Auto-generate a candidate `.feature` scenario skeleton
3. Human-review and promote to permanent regression test in `test/features/`

This keeps random/fuzz exploration and deterministic regression suites tightly connected.

### 6) Execution Plan (Phased)

Phase 1: Foundation

- Build AST generator + pretty-printer
- Add deterministic seed support and reproducible run logs
- Add panic/no-error/stability oracles

Phase 2: Differential checks

- Add DataWeave comparison harness for a supported subset
- Define normalization rules and mismatch reporting

Phase 3: Mutation + shrinking

- Add corpus-driven mutator
- Implement AST/input shrinkers and minimal repro output

Phase 4: CI and scaling

- Add short budget run in CI (for example, fixed seed + N cases)
- Add longer soak run locally/nightly with corpus persistence

### 7) Success Metrics

Track:

- Unique failures found per 10k generated cases
- Shrink ratio (original size vs minimized size)
- Time-to-minimal-repro
- Number of promoted cucumber regressions
- Coverage deltas in preprocessor/evaluator packages

The approach is successful when it continuously finds novel failures early, produces small reproducible test cases, and increases stable regression coverage without requiring large manual expression authoring.

# Mind Map - InfoMunge

> **For AI Agents:** Start with overview nodes [1-5], then follow inline references. Keep this file compact at roughly 20-50 numbered nodes. Update an existing node when possible; do not append a dated incident paragraph for a recurrence.
>
> **Memory review:** Before committing a MIND_MAP change, run `awk '/^\\[[0-9]+\\]/{n++} END {exit (n < 20) + (n > 50)}' MIND_MAP.md` and inspect `wc -l -c MIND_MAP.md`. If the check fails or the file grows materially, merge repeated guidance before adding more nodes.

[1] **Project Overview** - InfoMunge is a Go tool for transforming text data with DataWeave-inspired syntax. It parses inputs and headers, rewrites syntax, evaluates an AST, then formats output [3].

[2] **Repo Map** - Key entrypoints: CLI `cmd/infomunge/main.go`, CLI app `internal/cli/app.go`, WASM build `cmd/infomunge-wasm/main.go`, inputs `internal/io/input.go`, runner `internal/runner/runner.go`, preprocessor `internal/preprocessor/*`, evaluator `internal/evaluator/*`, formats `pkg/formats/*`, modules `modules/dw/core/*.im`, and standalone playground `docs/playground/index.html`.

[3] **Execution Pipeline** - Flow: CLI -> Inputs -> Header -> Preprocess -> Evaluate -> Format output. Header directives are parsed in the runner, preprocessing rewrites DataWeave-like syntax, evaluator execution preserves lazy values, and format dispatch serializes the result.

[4] **User-Facing Modes** - CLI accepts inline or file scripts and repeated named `-i` inputs. `./infomunge --server` is loopback-only without `--api-key`; the playground posts to `/run`, while standalone WASM evaluates locally. Server and WASM adapters resolve requested output MIME types for headerless scripts and retain raw text only when no MIME is available [31].

[5] **Formats** - Codec-local files under `pkg/formats` register MIME types/extensions through `pkg/formats/registry.go`; aliases and options use `options_dispatch.go`. `docs/FORMATS.md` is the user-facing registry/fidelity matrix and its consistency test must change with registrations. JSON/NDJSON preserve native integers and reject out-of-range integer tokens; Java properties are input-only; CSV output requires an array of objects and tracks first-seen headers. YAML and XML enforce single-document and resource/structure limits, with XML validating names and namespaces before rendering.

[6] **Preprocessor** - Syntax transforms live in `internal/preprocessor/*` and their ordered phase/loop/mapping contracts belong in `stages.go` and `transform_contract.go`. Preserve source mapping and be careful with bracket/brace scanning, grouped operands, and explicit versus implicit lambda boundaries. Collection operators, selectors, `update`, and infix `mod` have deliberate precedence/boundary rules; test both method-call and bare-infix forms when parity is intended. Nested collection lambdas rewrite inside-out so `$`/`$$` bindings remain distinct.

[7] **Evaluator** - AST evaluation and builtins live in `internal/evaluator/*`; new builtins need a handler plus metadata/dispatch spec. User functions share `invokeUserLambda` and exact ordinary-call arity; collection callbacks intentionally accept needed argument prefixes. Preserve cycle-safe structural equality, exact integer/rational behavior, lazy evaluation, `coerceToString`, null/absence semantics, and non-mutating updates. Missing fields from external input may defer as absence until `default`; script-created object misses remain ordinary null.

[8] **Runner And Modules** - Runner orchestration is in `internal/runner/runner.go`; module loading is in `internal/runner/module_loader.go`; core module scripts live under `modules/dw/core`.

[9] **Testing** - Cucumber tests run with `go test -v ./test -timeout 5m`; feature paths passed through `GODOG_PATHS` are relative to `./test`, and `TestFeatures` should be selected for focused runs. Add cucumber coverage for changed user-visible behavior and focused Go tests for internal behavior. Differential tests must classify both runtimes and keep temporary exceptions narrow and tied to an open beads issue. Preprocessor audits use `PrepareForParsing` with `Options.TraceTransforms`.

[10] **Quality Gates** - Prefer targeted tests, then `INFOMUNGE_SKIP_GODOG=1 go test ./... -timeout 5m` for the bounded package suite and cucumber separately when needed. In restricted workspaces, use an absolute repository-local cache such as `GOCACHE=/Users/david/Documents/code/repos/infomunge/tmp/go-build-cache`. Documentation/workflow-only edits need command, link, example, or formatting validation; cucumber is unnecessary unless runtime behavior changes. Intensive mutation and differential soaks are opt-in.

[11] **Version Control** - Use `jj` only; never use git. Run JJ operations sequentially, use `jj commit -m <description>`, and do not pass a message to `jj new`. `jj diff` has no `--check`; use focused text checks for whitespace.

[12] **Beads Workflow** - Issue tracking uses `br`: `bv --robot-triage`, `br show <id>`, `br update <id> --status in_progress`, `br close <id>`, and `br sync --flush-only`. `br sync --flush-only` exports `.beads/issues.jsonl`; it does not commit, so JJ must commit the exported file.

[13] **Landing The Plane** - Before handoff, file follow-up issues, run relevant quality gates, update/close the active issue, sync beads, remove temporary artifacts, verify `jj status`, and commit with a descriptive JJ message. Work on one beads task at a time before committing.

[14] **Agent Constraints** - `AGENTS.md` is authoritative; `GEMINI.md` imports it and is tool-specific. Prefer Go for runtime logic, do not add languages/dependencies/software without approval, stay inside the repository, put temporary files under `tmp`, and record durable mistakes and user preferences here compactly.

[15] **Docs** - `docs/README.md` indexes user and contributor documentation, including architecture, extension, and testing workflows. Remove completed plans and point-in-time audits once findings are implemented or tracked in beads.

[16] **Standalone Playground** - `make playground-wasm-serve` builds WASM and serves `docs/playground` on `127.0.0.1:8081`; `make playground` runs the Go backend on `127.0.0.1:8080`; `make playground-wasm` only rebuilds assets. Legacy gopherjs artifacts are removed.

[17] **User Preferences** - The user values explicit operating instructions in `AGENTS.md`, durable coordination in `MIND_MAP.md`, process improvements, and concrete beads follow-up tickets instead of implicit TODOs.

[18] **Ticket Selection Preference** - When asked for the next ticket, read `README.md`, `AGENTS.md`, and `MIND_MAP.md`, run `bv --robot-triage`, and prefer work already marked `in_progress` before claiming new work.

[19] **Review Preference** - For quality reviews, focus on the largest substantiated issues, use executable adversarial repros where practical, and create concrete beads tickets immediately for significant findings; apply the `quality` label when requested.

[20] **Memory Hygiene** - Keep roughly 20-50 compact nodes. Merge repeated pitfalls into category nodes, preserve only prevention steps that change future behavior, remove stale closed-ticket references, and review size/node count before committing. The review check is defined at the top of this file.

[21] **Patch Discipline** - Use `apply_patch` for manual edits, including temporary helper source; do not write files through shell redirection or shell patch commands. Inspect exact context before patching, split unrelated file edits, reopen changed regions, and remove imports when the last use disappears. Reflection-based AST helpers must account for cycles; declare differently typed Go slices separately.

[22] **Search Discipline** - Use `rg` first and `rg --files` before opening an unfamiliar path. Put flags before paths and `--` before patterns that may begin with `-`; use `rg -F` for punctuation-heavy literal searches. Keep discovery and inspection separate, and let expected no-match searches return normally rather than masking them with shell control operators.

[23] **Stateful Command Discipline** - Do not parallelize beads mutations with reads or any JJ operation. Use one logical shell command per call: no pipelines, `&&`, `||`, `;`, or multiple command lines. Run long commands directly and preserve their returned session/cell handle until completion. In orchestration code, call `tools.exec_command({...})` with raw JavaScript and surface every result; never bundle JJ with file reads, triage, or other stateful work.

[24] **Shell Quoting And Variables** - Avoid backticks and `$` forms in double-quoted shell search/probe strings; quote complete scripts deliberately or use a temporary file. Ordinary single quotes preserve literal `\\n`, and zsh `path`/`status` are special parameters, so use task-specific variable names. For stdout probes, declare the MIME in the script header and omit an output-path flag.

[25] **Cucumber Harness Pitfalls** - CLI scenarios use input-content steps plus the registered application-running step; in-process runner scenarios use a script step and `When I run the script`. Use docstrings for multiline fixtures, check `test/godog_test.go` for registered assertion wording, and select `TestFeatures` for focused `GODOG_PATHS` runs.

[26] **Script Formatting Pitfalls** - Keep complex function calls on one expression line unless multiline parser coverage exists. Parenthesize repeated unary minus. Runnable documentation snippets need a header and `---`; exact JSON expectations must account for compact serialization and Go escaping of `<`, `>`, and `&`.

[27] **Parser And Generator Safety** - Inline conditionals in dense object literals are parser-fragile; use simpler fixtures or parser-specific tests when unrelated. Generated expressions should avoid scientific notation and parser-active `$(` strings unless targeted. Normalize rounded negative zero in byte-stable JSON properties, use shared plain-ASCII differential strings, and assert runtime numeric types rather than assuming all literals are `float64`.

[28] **Error And Context Discipline** - Assert final rendered CLI errors, preserve filename context, and account for the direct `EvalError: ` prefix in focused checks. Keep generated positions unresolved until sourcemap formatting. Preserve public/internal call signatures where practical by adding context-aware variants first, then migrate callers incrementally. Resolve `ExecutionResult` before declaring lazy execution successful.

[29] **Server Exposure Safety** - HTTP server mode defaults to `127.0.0.1:8080`; unauthenticated non-loopback or wildcard binds are rejected. Network exposure requires an API key because `/run` evaluates caller-supplied scripts.

[30] **Object Order** - `Object` remains a map alias for API compatibility. Ordered construction uses `values.NewObject` and `values.SetObjectValue`; observable iteration uses `values.ObjectKeys`, which preserves tracked insertion order and sorts legacy untracked maps. Propagate metadata through readers and object-producing paths while keeping structural equality order-insensitive; do not add public metadata mutation without a concrete caller.

[31] **Assertion And AST Fixtures** - `must` returns a `MatcherResult`, while `assert`/`assertThat` return the asserted value. Builtin unit calls needing positions require a non-nil `ast.CallExpr.Fun`; raw evaluator tests use internal `__range` because preprocessing rewrites public `range`, and raw `Evaluate` does not parse InfoMunge object syntax.

[32] **Function Resolution And Arity** - Lexical functions resolve before public builtins, while generated double-underscore helpers resolve before lexical bindings; embedded wrappers use `__native`. `BuiltinSpec.Arity` is authoritative and registry tests enforce every declared bound before handlers execute.

[33] **Number Ordinal Indexing** - Integer ordinal selectors on numbers index the rune sequence from standard string coercion, including sign, decimal point, and exponent characters. Negative indexes count from the end; either out-of-range direction returns `null`, unlike strict direct string/array indexes.

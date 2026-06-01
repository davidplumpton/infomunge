# Mind Map - InfoMunge

> **For AI Agents:** Start with overview nodes [1-5], then follow inline references. Keep this file compact, roughly 20-50 nodes. When behavior, workflow, mistakes, or user preferences change, update the relevant compact node instead of appending duplicate one-off entries.

[1] **Project Overview** - InfoMunge is a Go tool for transforming text data with DataWeave-inspired syntax. It parses inputs and headers, rewrites syntax, evaluates an AST, then formats output [3].

[2] **Repo Map** - Key entrypoints: CLI `cmd/infomunge/main.go`, CLI app `internal/cli/app.go`, WASM build `cmd/infomunge-wasm/main.go`, inputs `internal/io/input.go`, runner `internal/runner/runner.go`, preprocessor `internal/preprocessor/*`, evaluator `internal/evaluator/*`, formats `pkg/formats/*`, modules `modules/dw/core/*.im`, standalone playground `docs/playground/index.html`.

[3] **Execution Pipeline** - Flow: CLI -> Inputs -> Header -> Preprocess -> Evaluate -> Format output. Header directives are parsed in the runner, preprocessing rewrites DataWeave-like syntax, evaluator executes builtins and lazy values, then output formatting serializes the result.

[4] **User-Facing Modes** - CLI accepts scripts directly or with `-f`, supports repeated named `-i` inputs, and can run server mode via `./infomunge --server --listen :8080`. The playground posts scripts and inputs to `/run`.

[5] **Formats** - Format readers/writers live in `pkg/formats/*` with the registry in `pkg/formats/registry.go`. JSON, XML, CSV, YAML, Java properties, and text behavior are user-visible. CSV output requires an array of objects.

[6] **Preprocessor** - Syntax transforms live in `internal/preprocessor/*`. Add operators in `internal/preprocessor/transformers_*`; be careful around bracket/brace scanning and grouped operands.

[7] **Evaluator** - AST evaluation and builtins live in `internal/evaluator/*`. New builtins usually belong in `internal/evaluator/builtins_*`; preserve lazy evaluation and nondeterministic builtin behavior such as `now()`.

[8] **Runner And Modules** - Runner orchestration is in `internal/runner/runner.go`; module loading behavior is in `internal/runner/module_loader.go`. Core module scripts live under `modules/dw/core`.

[9] **Testing** - Cucumber tests run with `go test -v ./test`. For feature work, add or update cucumber coverage. Feature files live under `test/features` from the repo root; use `GODOG_PATHS=features/...` because the test harness resolves feature paths relative to `./test`.

[10] **Quality Gates** - Prefer targeted package and cucumber tests for changed areas. Avoid blanket `go test ./...` when scratch helper `main` files exist under `tmp`; use scoped packages or remove helper conflicts first.

[11] **Version Control** - Use `jj` only; never use git commands. Commit with `jj commit -m <description>` and do not use `jj new` with a message. Run JJ commands sequentially because status/describe/commit can touch working-copy state and ref locks.

[12] **Beads Workflow** - Issue tracking uses `br` (beads rust): `bv --robot-triage`, `br show <id>`, `br update <id> --status in_progress`, `br close <id>`, `br sync --flush-only`. `br sync --flush-only` exports the beads database to `.beads/issues.jsonl`; `br` never runs VCS commands or commits. Use `jj` to commit the exported beads files.

[13] **Landing The Plane** - Before handoff: file follow-up issues, run quality gates when code changed, close/update beads, run `br sync --flush-only`, clean up temporary artifacts, verify `jj status`, then `jj commit -m <description>`.

[14] **Agent Constraints** - Use Go only, do not install new software without asking, stay inside the infomunge directory, and put temp files under `tmp`. Track durable mistakes and user preferences here, compactly.

[15] **Docs** - See `docs/ARCHITECTURE.md`, `docs/EXTENDING.md`, and `docs/TESTING.md` for deeper implementation and testing details.

[16] **Standalone Playground** - Build WASM assets with `make playground-wasm`, then open `docs/playground/index.html`. Legacy gopherjs artifacts and `cmd/infomunge-js` have been removed.

[17] **User Preferences** - The user values explicit operating instructions in `AGENTS.md`, durable coordination in `MIND_MAP.md`, process-level improvements, and creating beads tickets for meaningful follow-up work instead of leaving implicit TODOs.

[18] **Ticket Selection Preference** - When asked for the next ticket, start from `README.md`, `AGENTS.md`, and `MIND_MAP.md`, run `bv --robot-triage`, and prefer continuing work already marked `in_progress` before claiming new work.

[19] **Review Preference** - For codebase quality reviews, focus on the biggest substantiated issues and turn significant findings into concrete beads tickets immediately.

[20] **Memory Hygiene** - Keep this file compact. Merge repeated pitfalls into category nodes, preserve prevention steps that materially change future behavior, and avoid growing a dated entry for every recurrence unless the recurrence reveals a new prevention rule.

[21] **Patch Discipline** - Use the dedicated `apply_patch` tool for manual edits. Do not invoke patching through shell commands. Keep broad refactors in small hunks and reopen edited regions before testing.

[22] **Search Discipline** - Use `rg` first for searches. Put all flags before `--`, and place patterns that may begin with `-` after `--`. Use `rg -U` only when true multiline matching is needed.

[23] **Stateful Command Discipline** - Do not parallelize beads mutations with reads (`br update`/`br close` then `br show`), and do not parallelize any JJ operations, including read-only `jj diff`/`jj status`. Run state-changing and state-inspecting VCS/beads commands sequentially.

[24] **Shell Quoting Discipline** - Avoid backticks inside double-quoted shell commands; they still execute command substitution, including in `rg` patterns. Avoid complex inline shell quoting for beads descriptions. Prefer simple descriptions or temp files under `tmp` when multiline payloads are needed.

[25] **Cucumber Harness Pitfalls** - CLI failure scenarios usually need `Given the following input content`, while in-process runner scenarios use `Given the following script`. For multiline fixture files, use docstring-backed steps so real newlines are written.

[26] **Script Formatting Pitfalls** - Keep complex function calls in cucumber scripts on a single expression line unless multiline parser coverage exists. JSON output expectations are exact text comparisons, so match compact serialized output.

[27] **Parser Fragility Notes** - Inline `if` expressions inside object literals and dense object-value expressions have repeatedly hit brace/branch scanning bugs. For unrelated coverage, use simpler objects, arrays, defaults, or isolate conditionals in parser-specific tests.

[28] **Generated Test Data Pitfalls** - Property-generated source literals should avoid scientific notation and parser-active string sequences such as interpolation-looking `$(` unless explicitly testing those paths.

[29] **Error Context Discipline** - When changing wrapped error messages or unified error plumbing, verify final rendered CLI stderr/output and preserve essential context such as filenames.

[30] **Context Refactor Discipline** - Preserve existing public/internal call signatures where practical by adding context-aware variants first, then migrate call sites incrementally to reduce compile fallout.

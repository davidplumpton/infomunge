# Mind Map - InfoMunge

> **For AI Agents:** Start with overview nodes [1-5], then follow inline references [N]. Update nodes immediately when behavior or workflow changes. Keep it compact (20-50 nodes). Use node IDs when referencing.

[1] **Project Overview** - InfoMunge is a Go-based tool for transforming text data with a DataWeave-inspired syntax; it parses inputs, headers, rewrites syntax, evaluates an AST, and formats output [3][2]. It supports multiple inputs and multiple formats (JSON, XML, CSV, YAML, Java properties) [6][10].

[2] **Repo Map** - Key entrypoints and modules: CLI `cmd/infomunge/main.go`, JS build `cmd/infomunge-js/main.go`, CLI app `internal/cli/app.go`, inputs `internal/io/input.go`, runner `internal/runner/runner.go`, preprocessor `internal/preprocessor/*`, evaluator `internal/evaluator/*`, formats `pkg/formats/*`, core module scripts `modules/dw/core/*.im` (Arrays/Binaries/Strings/Objects/Numbers), standalone playground `docs/playground/index.html` [9][7][8][6][10][21].

[3] **Execution Pipeline** - Flow: CLI -> Inputs -> Header -> Preprocess -> Evaluate -> Format output [2]. Header directives are parsed in the runner, preprocessing rewrites DataWeave-like syntax, evaluation executes the AST with builtins, then output formatting occurs [7][8][9][6].

[4] **CLI Usage** - Example: `./infomunge -i payload <payload.json> "%im 0.1 output text/csv --- payload"`; multiple inputs supported with repeated `-i` flags [10][6].

[5] **Server Mode** - `./infomunge --server --listen :8080` starts HTTP server with playground UI and `/run` endpoint for scripts and inputs [10].

[6] **Formats** - Implemented in `pkg/formats/*` with registry in `pkg/formats/registry.go`; supports JSON, XML, CSV, YAML, Java properties [2]. CSV output expects an array of objects; otherwise validation error [18].

[7] **Preprocessor** - Syntax transforms live in `internal/preprocessor/*`; add operators in `internal/preprocessor/transformers_*` [2].

[8] **Evaluator** - AST evaluation and builtins in `internal/evaluator/*`; new builtins go in `internal/evaluator/builtins_*` [2].

[9] **Runner** - `internal/runner/runner.go` orchestrates header parsing and execution pipeline; module loading behavior in `internal/runner/module_loader.go` [3][11].

[10] **Inputs** - Inputs are named, optional, and can include a `format`; format defaults to `text/plain` when omitted [4][5]. Inputs are parsed in `internal/io/input.go` [2].

[11] **Module Loading** - Adjust module loading logic in `internal/runner/module_loader.go`; keep in sync with runner behavior [9].

[12] **Testing** - Uses Cucumber for Go; run `go test -v ./test` [14]. When adding features, add a cucumber test [15].

[13] **Version Control** - Use `jj` only; never use git commands. Commit with `jj commit -m <description>` and do not use `jj new` with a message [19].

[14] **Beads Workflow** - Issue tracking via `bd` (beads): `bv --robot-triage` to find prioritized work, `bd show <id>`, `bd update <id> --status in_progress`, `bd close <id>`, `bd sync` [19].

[15] **Agent Constraints** - Use Go only; do not install new software; put temp files in `tmp`; run cucumber tests with a 5 minute timeout [12][14][19].

[16] **Docs** - See `docs/ARCHITECTURE.md`, `docs/EXTENDING.md`, and `docs/TESTING.md` for details [2][12].

[17] **Example Script** - DataWeave-like example uses `%im 0.1`, `output application/json`, function definition, and expression block after `---` [4].

[18] **CSV Output Constraint** - CSV output requires an array of objects; non-array results or non-object rows yield validation errors [6].

[19] **Landing the Plane** - Session completion: file issues, run quality gates, update issue status, run `bd sync`, clean up, verify, then `jj commit -m <description>` [12][13][14].

[20] **Multiple Inputs** - CLI supports multiple named inputs; server `/run` accepts `inputs` array with `name`, `content`, and optional `format` [4][5][10].

[21] **Standalone Playground (JS)** - Build with gopherjs: `gopherjs build ./cmd/infomunge-js -o docs/playground/infomunge.js`, then open `docs/playground/index.html` for a local JS runner (no server) [4][5][2].

[22] **Agent Memory Policy** - Keep a small running log in this file for (a) agent mistakes and prevention steps and (b) user preferences. Update it when new evidence appears.

[23] **Agent Mistakes Log** - 2026-02-08: During ScanState refactor, exponent rewrite briefly regressed for grouped operands (`(1 + 1) ** (2 + 1)` stayed untransformed). Why: stop-condition check ran before handling opening brackets. Prevention: in right-operand scanners, process open/close bracket transitions before operator-stop checks and verify grouped-operand cucumber scenarios before handoff [22].

[24] **User Preferences Log** - Confirmed preference: keep persistent memory in `MIND_MAP.md` for both agent mistakes and things the user likes [22].

[25] **User Likes (Known So Far)** - Likes explicit operating instructions in `AGENTS.md`; likes using `MIND_MAP.md` as a durable coordination artifact; likes capturing process-level improvements, not just code changes [24].

[26] **Ideas Backlog (Current)** - Add dated mini-entries for each future mistake/preference; add a short "Do/Don't" list for fast agent onboarding; keep this section concise to avoid noise while preserving learning value [22][23][24].

[27] **Agent Mistakes Log** - 2026-02-08: Attempted file edits by invoking `apply_patch` through `exec_command` instead of using the dedicated `apply_patch` tool. Why: tool-routing oversight while parallelizing edits. Prevention: run patches only via the `apply_patch` tool and reserve `exec_command` for shell commands [22].

[28] **Agent Mistakes Log** - 2026-02-08: Ran targeted cucumber test with `GODOG_PATHS=test/features/...` from repo root, but the test runner resolves paths relative to `./test` so it failed with "feature path ... is not available". Why: forgot path base in `test/godog_test.go` workflow. Prevention: use `GODOG_PATHS=features/...` when invoking `go test -run TestFeatures ./test` [22].

[29] **Agent Mistakes Log** - 2026-02-12: Property tests used scientific-notation numeric literals (for example `6.103515625e-05`) directly in generated expressions; these conflicted with parser/operator handling around `**`/`default` and caused parse errors. Why: reused generic literal generators without constraining numeric source formatting for script embedding. Prevention: for algebraic property scripts, generate fixed-point numeric literals (for example via `FormatFloat(..., 'f', ...)`) and avoid exponent notation in source strings [22].

[30] **Agent Mistakes Log** - 2026-02-12: Property tests reused unconstrained string literal generation and produced values like `"$(..."`, which collided with interpolation parsing when embedded in generated expressions. Why: string generators were not scoped to parser-safe source contexts. Prevention: for property-generated source literals, constrain to a safe ASCII subset unless a test is explicitly targeting interpolation/escape behavior [22].

[31] **Agent Mistakes Log** - 2026-02-12: Mutation determinism checks initially treated all expressions as deterministic and failed on legitimate time/random outputs (for example `now()`). Why: deterministic invariant lacked a nondeterministic-builtin exclusion list. Prevention: keep no-panic checks universal, but gate deterministic-result assertions behind an explicit allow/deny list for known nondeterministic builtins [22].

[32] **User Preferences Log** - 2026-02-13: User asked for codebase quality review focused on biggest issues and expects significant findings to be turned into detailed beads tickets immediately.

[33] **Agent Mistakes Log** - 2026-02-13: Attempted to invoke `apply_patch` through `exec_command` again while editing multiple files. Why: defaulted to a parallel shell-edit pattern and skipped the dedicated patch tool. Prevention: use the `apply_patch` tool directly for all patch hunks, then parallelize only read/test commands [22].

[34] **Agent Mistakes Log** - 2026-02-13: Changed widely used runner/evaluator function signatures directly (`Evaluate`, `parseHeader`, var-decl parsers), which caused broad compile fallout across tests and helpers. Why: optimized for strict API purity before checking compatibility surface. Prevention: preserve existing signatures and add context-aware variants first, then migrate call sites incrementally [22].

[35] **Agent Mistakes Log** - 2026-02-13: Updated cucumber harness to use per-scenario work directories but kept the compiled CLI binary path relative (`../tmp/...`), causing parallel scenarios to fail with `fork/exec ... no such file or directory` when `cmd.Dir` changed. Why: overlooked path resolution behavior after setting per-scenario working directories. Prevention: use absolute paths for shared executables before running commands from scenario-specific directories [22].

[36] **Agent Mistakes Log** - 2026-02-13: Created a beads issue with backticks inside a double-quoted shell command, which triggered command substitution (`test/features/...` attempted execution) and stripped key text from the issue description. Why: unsafe shell quoting while passing multiline descriptions to `bd create`. Prevention: write descriptions to a file in `tmp` and pass with `--description \"$(cat tmp/<file>)\"`; avoid backticks in shell-quoted payloads [22].

[37] **Agent Mistakes Log** - 2026-02-13: Ran `bd update` and `bd show` in parallel, which produced a stale `show` result (`open`) due command timing. Why: parallelized a state-mutating command with a state-reading command. Prevention: for beads state changes, run mutation (`bd update`, `bd close`) first, then run verification (`bd show`) sequentially [22].

[38] **Agent Mistakes Log** - 2026-02-14: Repeated the `bd update`/`bd show` parallelization mistake and observed conflicting ticket status reads in the same turn. Why: defaulted to parallel command batching without excluding stateful beads operations. Prevention: treat beads status mutations as strictly sequential workflows and avoid parallel wrappers around `bd update` + `bd show` [22].

[39] **Agent Mistakes Log** - 2026-02-14: Added new cucumber script expressions using multi-line function argument formatting with commas; parser rejected them (`expected operand, found ','`). Why: assumed multi-line argument layout behaves like Go/JSON formatting. Prevention: keep complex function calls in feature scripts on a single expression line unless existing parser coverage confirms multiline argument separators [22].

[40] **Agent Mistakes Log** - 2026-02-14: Ran `go test ./...` without excluding `tmp`, causing expected build collisions from multiple helper `main` files in `tmp/`. Why: used blanket quality gate in a repo that intentionally keeps multiple executable scratch files in `tmp`. Prevention: run targeted package/feature tests for changed areas (or exclude `tmp`) instead of full-recursive `./...` when scratch binaries are present [22].

[41] **User Preferences Log** - 2026-02-14: User explicitly asked to create additional beads tickets whenever meaningful follow-up work is discovered during implementation, instead of leaving implicit TODOs.

[42] **Agent Mistakes Log** - 2026-02-14: Invoked `apply_patch` through `exec_command` while editing `pkg/formats/reader.go`; tool warning confirmed the routing error. Why: slipped back into shell-based patch habit under time pressure. Prevention: invoke the dedicated `apply_patch` tool directly for every patch and reserve `exec_command` for non-edit shell operations [22].

[43] **Agent Mistakes Log** - 2026-02-14: Added a cucumber script using multiline function arguments with commas again, and parser failed (`expected operand, found ','`). Why: ignored existing parser formatting constraint while drafting a new scenario. Prevention: keep feature-script function calls on one expression line unless there is proven multiline parser coverage for that pattern [22][39].

[44] **Agent Mistakes Log** - 2026-02-13: Applied an oversized patch hunk to `internal/preprocessor/rewriter_handlers.go` that briefly corrupted the `handleCloseBrace` function body. Why: attempted a very broad single-shot patch edit instead of smaller targeted hunks. Prevention: split large refactors into scoped patches and immediately validate edited regions with focused `sed` checks before running tests [22].

[45] **Agent Mistakes Log** - 2026-02-14: Ran `rg` with a search pattern beginning with `-` without using `--`, causing option parsing errors and a noisy search loop. Why: rushed CLI composition for ripgrep flags/pattern ordering. Prevention: when a pattern may begin with `-`, use `rg ... -- \"<pattern>\" <path>` and place `-g` options before `--` [22].

[46] **Agent Mistakes Log** - 2026-02-14: Batched `bd update` and `bd show` in parallel again, producing contradictory status reads (`in_progress` and stale `open`) for the same ticket in one run. Why: reused parallel wrapper around a state mutation plus immediate read. Prevention: run beads mutations and status verification strictly sequentially (`bd update` first, then `bd show`) [22][37][38].

[47] **User Preferences Log** - 2026-02-14: User asked to pick the next ticket from `README.md`, `AGENTS.md`, and `MIND_MAP.md` context with preference for work already marked `in_progress`; prioritize continuing active in-flight tickets first.

[48] **Agent Mistakes Log** - 2026-02-14: Ran `rg` with a pattern beginning with `--lazy` without placing `--` before the search pattern, causing option parsing failure. Why: reused a quick search template without guarding flag-like patterns. Prevention: always insert `--` before any pattern that may start with `-` (for example `rg -n -- \"--lazy\" ...`) [22][45].

[49] **Agent Mistakes Log** - 2026-02-14: While replacing duplicate determinism checks, introduced unrelated placeholder type assertions in `internal/testing/properties/nopanic_test.go`, creating noisy compile risk before correction. Why: rushed a broad patch instead of minimal targeted edits. Prevention: keep refactor patches narrowly scoped to intended lines and immediately re-open edited files for sanity checks before running tests [22].

[50] **User Preferences Log** - 2026-02-14: User asked to create a dedicated beads ticket for measuring cucumber-driven coverage and increasing coverage where it adds practical confidence.

[51] **Agent Mistakes Log** - 2026-02-14: Ran `bd close` and `bd show` in parallel, which again produced a stale `show` status (`in_progress`) even though the close succeeded. Why: reused parallel batching with a state mutation plus immediate read. Prevention: run all beads mutations (`bd update`, `bd close`) and verification (`bd show`) strictly sequentially [22][37][38][46].

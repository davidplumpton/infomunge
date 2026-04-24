# Mind Map - InfoMunge

> **For AI Agents:** Start with overview nodes [1-5], then follow inline references [N]. Update nodes immediately when behavior or workflow changes. Keep it compact (20-50 nodes). Use node IDs when referencing.

[1] **Project Overview** - InfoMunge is a Go-based tool for transforming text data with a DataWeave-inspired syntax; it parses inputs, headers, rewrites syntax, evaluates an AST, and formats output [3][2]. It supports multiple inputs and multiple formats (JSON, XML, CSV, YAML, Java properties) [6][10].

[2] **Repo Map** - Key entrypoints and modules: CLI `cmd/infomunge/main.go`, WASM build `cmd/infomunge-wasm/main.go`, CLI app `internal/cli/app.go`, inputs `internal/io/input.go`, runner `internal/runner/runner.go`, preprocessor `internal/preprocessor/*`, evaluator `internal/evaluator/*`, formats `pkg/formats/*`, core module scripts `modules/dw/core/*.im` (Arrays/Binaries/Strings/Objects/Numbers), standalone playground `docs/playground/index.html` [9][7][8][6][10][21].

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

[14] **Beads Workflow** - Issue tracking via `br` (beads rust): `bv --robot-triage` to find prioritized work, `br show <id>`, `br update <id> --status in_progress`, `br close <id>`, `br sync` [19].

[15] **Agent Constraints** - Use Go only; do not install new software; put temp files in `tmp`; run cucumber tests with a 5 minute timeout [12][14][19].

[16] **Docs** - See `docs/ARCHITECTURE.md`, `docs/EXTENDING.md`, and `docs/TESTING.md` for details [2][12].

[17] **Example Script** - DataWeave-like example uses `%im 0.1`, `output application/json`, function definition, and expression block after `---` [4].

[18] **CSV Output Constraint** - CSV output requires an array of objects; non-array results or non-object rows yield validation errors [6].

[19] **Landing the Plane** - Session completion: file issues, run quality gates, update issue status, run `br sync`, clean up, verify, then `jj commit -m <description>` [12][13][14].

[20] **Multiple Inputs** - CLI supports multiple named inputs; server `/run` accepts `inputs` array with `name`, `content`, and optional `format` [4][5][10].

[21] **Standalone Playground (WASM)** - Build with `make playground-wasm` (runs `GOOS=js GOARCH=wasm go build -o docs/playground/infomunge.wasm ./cmd/infomunge-wasm`), then open `docs/playground/index.html` for a local WASM runner (no server). Legacy gopherjs artifacts and `cmd/infomunge-js` have been removed [4][5][2].

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

[36] **Agent Mistakes Log** - 2026-02-13: Created a beads issue with backticks inside a double-quoted shell command, which triggered command substitution (`test/features/...` attempted execution) and stripped key text from the issue description. Why: unsafe shell quoting while passing multiline descriptions to `br create`. Prevention: write descriptions to a file in `tmp` and pass with `--description \"$(cat tmp/<file>)\"`; avoid backticks in shell-quoted payloads [22].

[37] **Agent Mistakes Log** - 2026-02-13: Ran `br update` and `br show` in parallel, which produced a stale `show` result (`open`) due command timing. Why: parallelized a state-mutating command with a state-reading command. Prevention: for beads state changes, run mutation (`br update`, `br close`) first, then run verification (`br show`) sequentially [22].

[38] **Agent Mistakes Log** - 2026-02-14: Repeated the `br update`/`br show` parallelization mistake and observed conflicting ticket status reads in the same turn. Why: defaulted to parallel command batching without excluding stateful beads operations. Prevention: treat beads status mutations as strictly sequential workflows and avoid parallel wrappers around `br update` + `br show` [22].

[39] **Agent Mistakes Log** - 2026-02-14: Added new cucumber script expressions using multi-line function argument formatting with commas; parser rejected them (`expected operand, found ','`). Why: assumed multi-line argument layout behaves like Go/JSON formatting. Prevention: keep complex function calls in feature scripts on a single expression line unless existing parser coverage confirms multiline argument separators [22].

[40] **Agent Mistakes Log** - 2026-02-14: Ran `go test ./...` without excluding `tmp`, causing expected build collisions from multiple helper `main` files in `tmp/`. Why: used blanket quality gate in a repo that intentionally keeps multiple executable scratch files in `tmp`. Prevention: run targeted package/feature tests for changed areas (or exclude `tmp`) instead of full-recursive `./...` when scratch binaries are present [22].

[41] **User Preferences Log** - 2026-02-14: User explicitly asked to create additional beads tickets whenever meaningful follow-up work is discovered during implementation, instead of leaving implicit TODOs.

[42] **Agent Mistakes Log** - 2026-02-14: Invoked `apply_patch` through `exec_command` while editing `pkg/formats/reader.go`; tool warning confirmed the routing error. Why: slipped back into shell-based patch habit under time pressure. Prevention: invoke the dedicated `apply_patch` tool directly for every patch and reserve `exec_command` for non-edit shell operations [22].

[43] **Agent Mistakes Log** - 2026-02-14: Added a cucumber script using multiline function arguments with commas again, and parser failed (`expected operand, found ','`). Why: ignored existing parser formatting constraint while drafting a new scenario. Prevention: keep feature-script function calls on one expression line unless there is proven multiline parser coverage for that pattern [22][39].

[44] **Agent Mistakes Log** - 2026-02-13: Applied an oversized patch hunk to `internal/preprocessor/rewriter_handlers.go` that briefly corrupted the `handleCloseBrace` function body. Why: attempted a very broad single-shot patch edit instead of smaller targeted hunks. Prevention: split large refactors into scoped patches and immediately validate edited regions with focused `sed` checks before running tests [22].

[45] **Agent Mistakes Log** - 2026-02-14: Ran `rg` with a search pattern beginning with `-` without using `--`, causing option parsing errors and a noisy search loop. Why: rushed CLI composition for ripgrep flags/pattern ordering. Prevention: when a pattern may begin with `-`, use `rg ... -- \"<pattern>\" <path>` and place `-g` options before `--` [22].

[46] **Agent Mistakes Log** - 2026-02-14: Batched `br update` and `br show` in parallel again, producing contradictory status reads (`in_progress` and stale `open`) for the same ticket in one run. Why: reused parallel wrapper around a state mutation plus immediate read. Prevention: run beads mutations and status verification strictly sequentially (`br update` first, then `br show`) [22][37][38].

[47] **User Preferences Log** - 2026-02-14: User asked to pick the next ticket from `README.md`, `AGENTS.md`, and `MIND_MAP.md` context with preference for work already marked `in_progress`; prioritize continuing active in-flight tickets first.

[48] **Agent Mistakes Log** - 2026-02-14: Ran `rg` with a pattern beginning with `--lazy` without placing `--` before the search pattern, causing option parsing failure. Why: reused a quick search template without guarding flag-like patterns. Prevention: always insert `--` before any pattern that may start with `-` (for example `rg -n -- \"--lazy\" ...`) [22][45].

[49] **Agent Mistakes Log** - 2026-02-14: While replacing duplicate determinism checks, introduced unrelated placeholder type assertions in `internal/testing/properties/nopanic_test.go`, creating noisy compile risk before correction. Why: rushed a broad patch instead of minimal targeted edits. Prevention: keep refactor patches narrowly scoped to intended lines and immediately re-open edited files for sanity checks before running tests [22].

[50] **User Preferences Log** - 2026-02-14: User asked to create a dedicated beads ticket for measuring cucumber-driven coverage and increasing coverage where it adds practical confidence.

[51] **Agent Mistakes Log** - 2026-02-14: Ran `br close` and `br show` in parallel, which again produced a stale `show` status (`in_progress`) even though the close succeeded. Why: reused parallel batching with a state mutation plus immediate read. Prevention: run all beads mutations (`br update`, `br close`) and verification (`br show`) strictly sequentially [22][37][38][46].

[52] **Agent Mistakes Log** - 2026-02-14: Added cucumber failure scenarios using `Given the following script` together with `When I run the application and it fails`; this mismatched harness fields and produced a misleading parse error (`script must have a header with '---' separator`). Why: mixed runner-level and CLI-level step contracts. Prevention: use `Given the following input content` for CLI failure paths and reserve `Given the following script` for in-process runner steps [22].

[53] **Agent Mistakes Log** - 2026-04-02: Added a cucumber scenario that used the quoted `a file named ... with content "..."` step for multiline script/module fixtures, so literal `\n` sequences were written and the parser failed with `script must have a header with '---' separator`. Why: reused the single-line fixture step without checking whether it unescapes newline sequences. Prevention: use a docstring-backed file fixture step for multiline files and prefer it for scripts/modules that require real line breaks [22].

[54] **Agent Mistakes Log** - 2026-04-19: Ran `rg` with `--` before later `-g` flags while searching for `exprToString`, so ripgrep treated `-g` and `*.go` as paths and returned noisy errors. Why: inserted the end-of-options marker too early when guarding a pattern search. Prevention: keep all ripgrep flags before `--`, and only then pass the search pattern and paths [22].

[55] **User Preferences Log** - 2026-04-19: User asked for a fresh-eyes audit that starts from `README.md`, `AGENTS.md`, and `MIND_MAP.md` context and turns substantiated quality findings into concrete beads tickets.

[56] **Agent Mistakes Log** - 2026-04-19: Repeated the ripgrep flag-order mistake while searching for `dw::core` references, placing `--` before `-g` filters and getting noisy `No such file or directory` errors. Why: reused the guarded-pattern template without keeping glob flags ahead of the end-of-options marker. Prevention: for ripgrep, place every option such as `-g` before `--`, then pass the pattern and paths [22][54].

[57] **Agent Mistakes Log** - 2026-04-19: Ran `jj describe` and `jj commit` in parallel, which raced on JJ's ref lock and caused the description update to fail while the commit succeeded. Why: treated stateful JJ mutations as parallel-safe. Prevention: run JJ metadata and commit commands sequentially, especially `jj describe`, `jj commit`, and other operations that write refs [22].

[58] **Agent Mistakes Log** - 2026-04-19: Ran `br update` and `br show` in parallel again while picking up `bd-xpnk`, which produced a stale `OPEN` ticket view immediately after the successful status change to `in_progress`. Why: reused parallel batching on a state mutation plus immediate verification despite repeated prior incidents. Prevention: treat all beads mutations and follow-up reads as strictly sequential workflows; never batch `br update`/`br close` with `br show` [22][37][38][46][51].

[59] **Agent Mistakes Log** - 2026-04-19: Added new cucumber expectations with pretty-spaced JSON arrays, but this harness compares output text exactly and the runner emits compact JSON, so the focused feature failed despite correct behavior. Why: assumed semantic JSON comparison instead of exact string comparison in `Then the output should be`. Prevention: when adding cucumber expectations for JSON output, match the repo's compact serialized form (or verify against nearby scenarios) before rerunning tests [22].

[60] **Agent Mistakes Log** - 2026-04-20: Applied a ticket's suggested `WrapIOf` simplification verbatim in `internal/cli/app.go` before checking how `unifiederrors.Error` is rendered, which removed the missing filename from the CLI error. Why: trusted the issue text without validating the concrete user-visible error path for this error type. Prevention: when changing wrapped-error messages, verify the final rendered stderr/output path first and keep essential context such as filenames even if the low-level cause text is removed [22].

[61] **Agent Mistakes Log** - 2026-04-20: Repeated the ripgrep flag-pattern mistake while searching for `--lazy`: first omitted the `--` separator, then placed `--` before later `-g` filters so ripgrep treated them as paths. Why: rushed a flag-like pattern search without following the established option ordering rule. Prevention: for ripgrep searches with patterns that may start with `-`, keep every option (including `-g`) before `--`, then pass the pattern and paths [22][45][48][54][56].

[62] **Agent Mistakes Log** - 2026-04-25: Ran `jj status`, `jj diff --stat`, and `jj file list .beads` in parallel during verification; `jj file list` failed to obtain a keep-ref lock. Why: assumed read-only JJ commands were parallel-safe, but JJ may snapshot/touch refs during status-style reads. Prevention: run JJ commands sequentially whenever they may inspect or update working-copy state; parallelize only non-JJ file reads [22][57].

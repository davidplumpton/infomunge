# Mind Map - InfoMunge

> **For AI Agents:** Start with overview nodes [1-5], then follow inline references [N]. Update nodes immediately when behavior or workflow changes. Keep it compact (20-50 nodes). Use node IDs when referencing.

[1] **Project Overview** - InfoMunge is a Go-based tool for transforming text data with a DataWeave-inspired syntax; it parses inputs, headers, rewrites syntax, evaluates an AST, and formats output [3][2]. It supports multiple inputs and multiple formats (JSON, XML, CSV, YAML, Java properties) [6][10].

[2] **Repo Map** - Key entrypoints and modules: CLI `cmd/infomunge/main.go`, JS build `cmd/infomunge-js/main.go`, CLI app `internal/cli/app.go`, inputs `internal/io/input.go`, runner `internal/runner/runner.go`, preprocessor `internal/preprocessor/*`, evaluator `internal/evaluator/*`, formats `pkg/formats/*`, core module scripts `modules/dw/core/*.im` (Arrays/Strings/Objects/Numbers), standalone playground `docs/playground/index.html` [9][7][8][6][10][21].

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

[14] **Beads Workflow** - Issue tracking via `bd` (beads): `bd ready`, `bd show <id>`, `bd update <id> --status in_progress`, `bd close <id>`, `bd sync` [19].

[15] **Agent Constraints** - Use Go only; do not install new software; put temp files in `tmp`; run cucumber tests with a 5 minute timeout [12][14][19].

[16] **Docs** - See `docs/ARCHITECTURE.md`, `docs/EXTENDING.md`, and `docs/TESTING.md` for details [2][12].

[17] **Example Script** - DataWeave-like example uses `%im 0.1`, `output application/json`, function definition, and expression block after `---` [4].

[18] **CSV Output Constraint** - CSV output requires an array of objects; non-array results or non-object rows yield validation errors [6].

[19] **Landing the Plane** - Session completion: file issues, run quality gates, update issue status, run `bd sync`, clean up, verify, then `jj commit -m <description>` [12][13][14].

[20] **Multiple Inputs** - CLI supports multiple named inputs; server `/run` accepts `inputs` array with `name`, `content`, and optional `format` [4][5][10].

[21] **Standalone Playground (JS)** - Build with gopherjs: `gopherjs build ./cmd/infomunge-js -o docs/playground/infomunge.js`, then open `docs/playground/index.html` for a local JS runner (no server) [4][5][2].

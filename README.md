# InfoMunge

`InfoMunge` is a Go-based experiment that implements a subset of DataWeave-style data transformations. [DataWeave](https://docs.mulesoft.com/mule-runtime/latest/dataweave) is MuleSoft's language for accessing and transforming data in Mule applications; InfoMunge is an independent project with its own syntax and runtime.

## Requirements

- The project will be structured so the domain logic can be extracted as a library
- Data can be received and emitted in several formats
- JSON, XML, CSV, and YAML support both input and output; Java properties are input-only
- Multiple inputs are possible, each with a different name
- Features will be developed incrementally

## Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (version 1.25.5 or later)

### Installation

```bash
go build -o infomunge ./cmd/infomunge
```

### Usage

```bash
./infomunge -i payload=payload.json "%im 0.1 output text/csv --- payload"
./infomunge -i payload=payload.json -i input2=input2.json -f infomungefile.im
```

Notes:
- CSV output expects an array of objects; non-array results or non-object rows return a validation error.
- `text/plain` output writes string values as-is. Non-string values are rendered as compact JSON, with `null` rendered as `null`.
- See the [supported-format and fidelity matrix](docs/FORMATS.md) for MIME aliases, file extensions, input/output direction, structured options, and passthrough limitations.
- Date formatting uses a documented subset of Java SimpleDateFormat tokens; see `docs/DATE_FORMATS.md`.
- For null literals, use `null` for DataWeave compatibility. InfoMunge also accepts `nil` as an alias.
- InfoMunge supports `%` as a modulo operator. DataWeave-compatible scripts should use the left-associative `mod` operator instead. Like DataWeave, infix `mod` binds less tightly than `+`, `-`, `*`, `/`, and `%`; use parentheses to override that precedence.
- Range selectors such as `items[1 to 3]` are inclusive. If either bound is outside the collection after resolving negative indexes, the selector returns `null`; direct array and string indexes remain strict and return an out-of-bounds error.
- Header `input` directives are accepted for DataWeave compatibility and documentation only. Input data is parsed before execution by the CLI `-i` flags, server `/run` inputs, or embedding test harness; header `input` lines do not reparse, rename, validate, or create inputs.

### Comparison with DataWeave

To compare InfoMunge with DataWeave using the [DataWeave CLI](https://github.com/mulesoft/data-weave-cli), provide each input as `Name=File`:
```bash
dw run -i=payload=payload.json "%dw 2.0 output application/xml --- payload"
```
The DataWeave CLI requires the `run` subcommand; InfoMunge accepts `run` as an optional compatibility subcommand.

### Server Mode

Run the HTTP server:

```bash
./infomunge --server
```

The server defaults to `127.0.0.1:8080`, so it is reachable only from the local
machine. The `/run` endpoint evaluates caller-supplied transformation scripts;
only expose it to trusted users. To listen on a network interface, configure a
shared API key:

```bash
./infomunge --server --listen 0.0.0.0:8080 --api-key your-secret
```

InfoMunge rejects non-loopback server addresses when `--api-key` is omitted.
Clients must send the key in `X-API-Key` or as a bearer token.

Open the playground at `http://localhost:8080/` to use the interactive UI. It offers input panels on the left (add, name, and format each input), a script editor in the center, and a live result panel on the right that runs through the `/run` endpoint.

POST a script with optional inputs:

```bash
curl -X POST http://localhost:8080/run \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: your-secret' \
  -d '{"script":"%im 0.1\noutput application/json\n---\nsizeOf(payload)","output":"json","inputs":[{"name":"payload","format":"json","content":"[1,2,3]"}]}'
```

Notes:
- `inputs` are optional. Each input has `name`, `content`, and optional `format`.
- If `format` is omitted, inputs are treated as `text/plain`.
- `output` can be a format like `json` or a MIME type like `application/json`.

### Size Limits

- CLI script files and imported module files are limited to 1 MiB each.
- File-backed CLI inputs and stdin-backed CLI inputs are limited to 10 MiB each.
- Server `/run` request bodies are limited to 1 MiB; `readUrl` responses are limited to 10 MiB.

### Standalone Playground (WASM)

Build and serve the standalone browser playground:

```bash
make playground-wasm-serve
```

Then open `http://127.0.0.1:8081/` in a browser. This target builds the WASM
artifacts and serves `docs/playground` over HTTP using the repository's Go-based
static server; no additional software is required. Use `make playground-wasm`
when you only need to rebuild the assets.

Standalone mode evaluates scripts locally in WebAssembly and does not expose the
`/run` endpoint. To use the Go server backend and `/run` API instead, run
`make playground`; it binds to `127.0.0.1:8080`, so open
`http://127.0.0.1:8080/`.

## How It Works

InfoMunge runs a small pipeline: parse CLI inputs, parse header directives, rewrite
DataWeave-like syntax into Go-parseable expressions, evaluate an AST, then format output.

```
CLI -> Inputs -> Header -> Preprocess -> Evaluate -> Format output
```

Key entrypoint: `internal/runner/runner.go`.

### Module Map

- CLI: `cmd/infomunge/main.go`, `internal/cli/app.go`
- Inputs: `internal/io/input.go`
- Runner + header parsing: `internal/runner/runner.go`
- Preprocessor (syntax transforms): `internal/preprocessor/*`
- Evaluator (AST + builtins + lazy): `internal/evaluator/*`
- Formats (readers/writers): `pkg/formats/*`
- Core module scripts: `modules/dw/core/*.im` (for example `Arrays.im`, `Binaries.im`, `Strings.im`, `Objects.im`, `Numbers.im`)

### Where To Add Things

- New operator: implement in `internal/preprocessor/transformers_*.go` and add
  its ordered transform contract in `internal/preprocessor/stages.go`
- New builtin: implement in `internal/evaluator/builtins_*.go` and add its
  metadata/dispatch spec in `internal/evaluator/builtin_specs_*.go`
- New format: implement and register the codec in its local
  `pkg/formats/*.go` `init`; option handlers use
  `pkg/formats/options_dispatch.go`
- Module loading behavior: `internal/runner/module_loader.go`

See the [documentation index](docs/README.md) for current user and contributor
guides.

### Example: Define a function to flatten a list

Based on the DataWeave cookbook example:
https://docs.mulesoft.com/dataweave/latest/dataweave-cookbook-define-function-to-flatten-list

```im
%im 0.1
output application/json
fun flattenList(items) =
  items flatMap (item) -> if (typeOf(item) == "Array") flattenList(item) else [item]
---
flattenList([[1, [2, 3]], [4], 5])
```

**Output:**
```json
[1,2,3,4,5]
```

## Development

### Running Tests

This project uses Cucumber for Go.

```bash
go test -v ./test -timeout 5m
```

Repo-wide package tests should also pass. This skips Godog because feature tests have their own command:

```bash
INFOMUNGE_SKIP_GODOG=1 go test ./... -timeout 5m
```

Scratch helper programs under `tmp/` are isolated in a nested Go module so they do not participate in the repo-root `go test ./...` package walk. Run helpers from inside `tmp/` by passing the helper file you created:

```bash
cd tmp && go run ./your-helper.go
```

The bounded repo-wide command skips the mutation corpus soak by default. Run the mutation soak explicitly when needed:

```bash
INTENSIVE_TEST_SOAK=1 go test -v ./internal/testing/mutation -run TestMutatedCorpusExpressions_NoPanics_AndDeterministic -timeout 30m
```

When the external DataWeave CLI is installed, the bounded suite runs 50
differential comparisons. Run the larger 500-comparison budget explicitly:

```bash
make test-differential-soak
```

Generate coverage from cucumber tests (runtime-focused packages):

```bash
go test -v ./test -run TestFeatures -count=1 -timeout 5m \
  -coverprofile=tmp/cucumber.cover \
  -coverpkg=./internal/runner,./internal/preprocessor,./internal/evaluator,./pkg/formats
go tool cover -func=tmp/cucumber.cover | tail -n 1
```

See [docs/TESTING.md](docs/TESTING.md) for the authoritative test,
intensive-testing, and coverage workflows, including the package-level coverage
summary.

### Version Control

This project uses `jj` for version control.

### Tasks and Issues

This project uses [beads_rust](https://github.com/Dicklesworthstone/beads_rust)
and its `br` executable for issue tracking. Follow [AGENTS.md](AGENTS.md) for the
project's complete `br` and `jj` workflow.

`bv --robot-triage` may emit legacy suggestions that invoke `bd`. Translate those
commands to the equivalent `br` commands instead of copying them verbatim. In
particular, export issue updates with `br sync --flush-only`; `br` does not run
version-control commands, so commit the exported files separately with `jj` as
described in [AGENTS.md](AGENTS.md).

## Third-Party Software

The standalone playground and compiled binaries include third-party software.
See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for the applicable license
and attribution notices.

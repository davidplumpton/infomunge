# InfoMunge

`InfoMunge` is a Go-based tool for processing and transforming text data. It is intended to be an experiment in vibe-coding a subset of functionality of DataWeave 2, an extremely powerful data transformation language from Mulesoft and Salesforce. It was announced that DataWeave would be open-sourced, although that seems to be taking a while. DataWeave is a trademarked name, so we will call this InfoMunge.

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
- Date formatting uses a documented subset of Java SimpleDateFormat tokens; see `docs/DATE_FORMATS.md`.
- For null literals, use `null` for DataWeave compatibility. InfoMunge also accepts `nil` as an alias.

### Comparison to Datawave

To compare the operating of infomunge to dataweave we can run the dataweave cli, for example:
```bash
dw run -i payload <payload.json> "%dw 2.0 output application/xml --- payload"
```
Note that dw uses an extra "run" command.

### Server Mode

Run the HTTP server:

```bash
./infomunge --server --listen :8080
```

Protect the `/run` endpoint with a shared API key:

```bash
./infomunge --server --listen :8080 --api-key your-secret
```

Open the playground at `http://localhost:8080/` to use the interactive UI. It offers input panels on the left (add, name, and format each input), a script editor in the center, and a live result panel on the right that runs through the `/run` endpoint.

POST a script with optional inputs:

```bash
curl -X POST http://localhost:8080/run \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: your-secret' \
  -d '{"script":"%im 0.1\\noutput application/json\\n---\\nsizeOf(payload)","output":"json","inputs":[{"name":"payload","format":"json","content":"[1,2,3]"}]}'
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

Build standalone browser assets:

```bash
make playground-wasm
```

Then open `docs/playground/index.html` in a browser. The page uses a local JS runner and does not require the server.

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

- New operator: `internal/preprocessor/transformers_*`
- New builtin: `internal/evaluator/builtins_*`
- New format: `pkg/formats/*` + `pkg/formats/registry.go`
- Module loading behavior: `internal/runner/module_loader.go`

See `docs/ARCHITECTURE.md`, `docs/EXTENDING.md`, and `docs/TESTING.md` for details.

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
go test -v ./test
```

Repo-wide package tests should also pass:

```bash
go test ./...
```

Scratch helper programs under `tmp/` are isolated in a nested Go module so they do not participate in the repo-root `go test ./...` package walk. Run helpers from inside `tmp/` by passing the helper file you created:

```bash
cd tmp && go run ./your-helper.go
```

Generate coverage from cucumber tests (runtime-focused packages):

```bash
timeout 5m go test -v ./test -run TestFeatures -count=1 \
  -coverprofile=tmp/cucumber.cover \
  -coverpkg=./internal/runner,./internal/preprocessor,./internal/evaluator,./pkg/formats
go tool cover -func=tmp/cucumber.cover | tail -n 1
```

### Version Control

This project uses `jj` for version control.

### Tasks and Issues

This project uses Beads https://github.com/steveyegge/beads

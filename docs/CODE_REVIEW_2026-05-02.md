# Source Code Review - 2026-05-02

## Scope

Reviewed production source files under `cmd`, `internal`, `pkg`, and `modules/dw/core/sources.go`.
Excluded `_test.go` files, `tmp`, generated/runtime assets, and `internal/testing` support code.

Scores are 0 to 10. For `Complexity`, 10 means low, well-contained complexity; 0 means difficult to change safely.

## Executive Summary

The project has a solid shape: CLI/server boundaries are understandable, the runner now has an `ExecutionResult` path that separates evaluation from formatting, evaluator runtime capabilities are separated from user variables, and the format registry has a clean production/test split. The main maintainability risk remains the preprocessor, where several scanner-heavy files are large, hand-written, and hard to modify without broad regression coverage.

Six follow-up beads were filed for the review findings that should become implementation work:

| Bead | Priority | Type | Summary |
| --- | --- | --- | --- |
| `bd-1vpd` | P1 | Bug | Escape XML output text and attributes |
| `bd-18bu` | P1 | Bug | Harden `readUrl` private-network validation against DNS rebinding |
| `bd-3ejl` | P2 | Bug | Add size limits for script and file-backed CLI inputs |
| `bd-2wlj` | P2 | Bug | Support `text/plain` output formatting |
| `bd-3tw4` | P3 | Task | Reduce legacy parser and error adapter surface |
| `bd-22rt` | P3 | Task | Split high-complexity preprocessor transforms and standardize mappings |

## Key Findings

### P1: XML output is not escaped

`pkg/formats/xml.go` writes element text and attribute values using formatted raw values. That can produce malformed XML and can let data break out of expected XML structure. The fix should use XML escaping for both text nodes and attributes, and should add cucumber coverage that exercises special characters in XML output.

Tracked by `bd-1vpd`.

### P1: `readUrl` validation can be bypassed by DNS rebinding

`internal/runtimeio/services.go` validates host resolution before performing the HTTP request, but the actual transport resolves again later. That leaves a time-of-check/time-of-use gap, and lookup failures currently fail open. The safer design is to enforce validation on the actual dial target and fail closed for ambiguous private-network checks.

Tracked by `bd-18bu`.

### P2: Some file reads are unbounded

`internal/io/input.go`, `internal/cli/app.go`, and file path reads in runner paths use `os.ReadFile` without the read limits applied elsewhere. Stdin, server bodies, URL reads, and XLSX paths have better limits. Large script or input files can still cause excessive memory use.

Tracked by `bd-3ejl`.

### P2: `text/plain` output is selected but not implemented

`pkg/formats/text.go` registers a text reader only. Server content-type normalization can select `text/plain`, which then fails when formatting output. Either add a text writer or stop accepting `text/plain` as an output format.

Tracked by `bd-2wlj`.

### P3: Runner and error packages have legacy surface area

`internal/runner/declarations.go` still contains older declaration parsing and handler paths that overlap with `declaration_ir.go`, and `internal/errors/adapters.go` exposes adapters that are not used by production code outside their own file. This increases the mental model without adding much active behavior.

Tracked by `bd-3tw4`.

### P3: Preprocessor transform files are too large

The preprocessor contract and staged pipeline are good foundations, but several transform files mix scanning, rewriting, mapping, and syntax-specific exceptions in the same long functions. This raises regression risk for language changes.

Tracked by `bd-22rt`.

## Area Scores

| Area | Overall | Notes |
| --- | ---: | --- |
| CLI, HTTP, IO, runtime services | 7.8 | Clear user boundary, but URL/file-input hardening needs work. |
| Runner | 7.0 | Good result model and module split; declaration legacy paths reduce clarity. |
| Preprocessor | 6.1 | Good staged architecture; highest complexity and regression risk. |
| Evaluator | 7.0 | Registry model is strong; some builtin files are broad and dense. |
| Formats | 7.4 | Cohesive format boundaries; XML is the major correctness outlier. |
| Shared infrastructure | 8.2 | Small, focused packages with good naming and low coupling. |

## File Ratings

### CLI, HTTP, IO, Runtime, Shared

| File | Clarity | Complexity | Naming | Cohesion | Coupling | Testability | Overall | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `cmd/infomunge/main.go` | 9 | 10 | 9 | 10 | 9 | 8 | 9 | Minimal entrypoint. |
| `cmd/infomunge-wasm/main.go` | 8 | 8 | 8 | 8 | 7 | 7 | 8 | Clean adapter with expected wasm boundary concerns. |
| `internal/cli/app.go` | 7 | 6 | 8 | 7 | 6 | 7 | 7 | Understandable CLI orchestration; mixes flag handling, input loading, and execution; script reads need limits. |
| `internal/cli/server.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Good server setup and body limits. |
| `internal/cli/webapp.go` | 9 | 10 | 9 | 10 | 9 | 8 | 9 | Focused embed wrapper. |
| `internal/handlers/run.go` | 8 | 8 | 8 | 8 | 7 | 8 | 8 | Good request boundary; exposes the `text/plain` output gap. |
| `internal/io/input.go` | 7 | 7 | 8 | 8 | 7 | 7 | 7 | Clear input parser; file-backed reads need limits. |
| `internal/io/input_name.go` | 9 | 9 | 9 | 9 | 8 | 9 | 9 | Focused validation helper. |
| `internal/output/output.go` | 8 | 8 | 8 | 8 | 7 | 7 | 8 | Good output metadata boundary; some XML-specific behavior leaks in. |
| `internal/runtimeio/services.go` | 7 | 6 | 8 | 8 | 6 | 7 | 7 | Runtime caps are useful; URL validation has security-sensitive coupling to networking. |
| `internal/readlimit/readlimit.go` | 9 | 10 | 9 | 10 | 9 | 9 | 9 | Small, clear limit helper. |
| `internal/version/version.go` | 8 | 10 | 8 | 9 | 9 | 7 | 8 | Simple version holder. |
| `internal/sourcemap/map.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Cohesive source mapping model. |
| `internal/stringutils/escape.go` | 9 | 10 | 9 | 10 | 9 | 9 | 9 | Tiny, focused helper. |
| `internal/stringutils/scan_state.go` | 8 | 8 | 8 | 9 | 8 | 8 | 8 | Useful shared scan state. |
| `internal/stringutils/scanner.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Helpful scanner helpers; state mutation requires care. |
| `internal/stringutils/stringutils.go` | 7 | 6 | 7 | 7 | 6 | 7 | 7 | Mixed operator/string helpers. |
| `modules/dw/core/sources.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Focused embedded-module source map. |
| `pkg/values/types.go` | 9 | 10 | 9 | 10 | 9 | 9 | 9 | Clean shared aliases. |

### Runner

| File | Clarity | Complexity | Naming | Cohesion | Coupling | Testability | Overall | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `internal/runner/execution.go` | 8 | 8 | 8 | 8 | 7 | 8 | 8 | `ExecutionResult` is a useful boundary. |
| `internal/runner/runner.go` | 7 | 6 | 8 | 6 | 6 | 7 | 7 | Core path is readable, but formatting, headers, and compatibility helpers share space. |
| `internal/runner/header_directives.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Good split for directive policy. |
| `internal/runner/declaration_ir.go` | 7 | 6 | 8 | 7 | 6 | 7 | 7 | Useful IR, but parsing remains dense. |
| `internal/runner/declarations.go` | 6 | 4 | 7 | 5 | 5 | 6 | 5 | Duplicates newer declaration paths and keeps legacy handlers alive. |
| `internal/runner/module_loader.go` | 8 | 8 | 8 | 8 | 7 | 7 | 8 | Clear module loading and path constraints. |
| `internal/runner/module_loader_js.go` | 8 | 8 | 8 | 8 | 7 | 7 | 8 | Clear JS variant. |
| `internal/runner/module_parser.go` | 8 | 8 | 8 | 8 | 6 | 7 | 8 | Clear parser wrapper with expected evaluator coupling. |
| `internal/runner/module_types.go` | 8 | 8 | 8 | 8 | 7 | 8 | 8 | Focused namespace model. |
| `internal/runner/expression_compiler.go` | 8 | 9 | 8 | 8 | 8 | 8 | 8 | Small adapter. |
| `internal/runner/error_helpers.go` | 8 | 9 | 8 | 9 | 8 | 8 | 8 | Focused error formatting helper. |
| `internal/runner/standard_modules.go` | 9 | 10 | 9 | 9 | 8 | 8 | 9 | Focused standard-module wrapper. |

### Preprocessor

| File | Clarity | Complexity | Naming | Cohesion | Coupling | Testability | Overall | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `internal/preprocessor/preprocessor.go` | 7 | 5 | 7 | 7 | 6 | 7 | 6 | Recursive rewrite hot path with high blast radius. |
| `internal/preprocessor/stages.go` | 8 | 6 | 8 | 8 | 6 | 8 | 7 | Good pipeline contract; many transforms share one path. |
| `internal/preprocessor/transform_contract.go` | 8 | 7 | 8 | 9 | 7 | 8 | 8 | Strong metadata and mapping contract. |
| `internal/preprocessor/modular_pipeline.go` | 9 | 9 | 9 | 9 | 8 | 9 | 9 | Small and clear. |
| `internal/preprocessor/mapping_pipeline.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Approximate mapping behavior is documented. |
| `internal/preprocessor/scan_state.go` | 8 | 9 | 8 | 8 | 8 | 8 | 8 | Focused aliases and scan constants. |
| `internal/preprocessor/scanner.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Tiny scanner helpers. |
| `internal/preprocessor/collection_operators.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Canonical operator list. |
| `internal/preprocessor/operator_scanners.go` | 7 | 6 | 7 | 8 | 7 | 7 | 7 | Good operator configuration; scanner logic is dense. |
| `internal/preprocessor/exact_mapping_helpers.go` | 8 | 7 | 7 | 8 | 7 | 8 | 8 | Good local buffer abstraction. |
| `internal/preprocessor/lambda_body_scan.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Focused scanner. |
| `internal/preprocessor/top_level_object.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Focused top-level object handling. |
| `internal/preprocessor/utils.go` | 6 | 5 | 6 | 5 | 5 | 6 | 6 | Utility grab bag. |
| `internal/preprocessor/rewriter_handlers.go` | 6 | 3 | 6 | 5 | 4 | 5 | 5 | Very large, exception-heavy rewrite file. |
| `internal/preprocessor/rewriter_handlers_objects.go` | 7 | 5 | 7 | 7 | 5 | 6 | 6 | Object special cases are understandable but coupled. |
| `internal/preprocessor/transformers_operators.go` | 6 | 4 | 6 | 6 | 5 | 6 | 5 | Many operator exceptions and scanning rules. |
| `internal/preprocessor/transformers_syntax.go` | 5 | 3 | 6 | 4 | 4 | 5 | 4 | Broadest, hardest transform file to safely change. |
| `internal/preprocessor/transformers_lambda.go` | 7 | 5 | 7 | 7 | 6 | 7 | 6 | Good scope, but scan-heavy. |
| `internal/preprocessor/transformers_collection.go` | 8 | 7 | 8 | 8 | 7 | 7 | 8 | Focused collection-operator adapter. |
| `internal/preprocessor/transformers_regex.go` | 7 | 5 | 7 | 7 | 6 | 7 | 6 | Better documented than most transform files, but large. |

### Evaluator

| File | Clarity | Complexity | Naming | Cohesion | Coupling | Testability | Overall | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `internal/evaluator/errors.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Clear error helpers. |
| `internal/evaluator/types.go` | 6 | 5 | 7 | 5 | 5 | 7 | 6 | Mixes coercion, number handling, dates, and formatting helpers. |
| `internal/evaluator/types_value.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Clear value conversion helpers. |
| `internal/evaluator/constants.go` | 9 | 10 | 9 | 10 | 9 | 8 | 9 | Simple constants. |
| `internal/evaluator/scope.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Runtime capability split is clean. |
| `internal/evaluator/visitor.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Clear AST visitor dispatch. |
| `internal/evaluator/evaluator.go` | 6 | 5 | 7 | 6 | 6 | 7 | 6 | Broad AST and operator core. |
| `internal/evaluator/eval_index.go` | 8 | 8 | 8 | 8 | 7 | 8 | 8 | Focused index behavior. |
| `internal/evaluator/lazy.go` | 8 | 6 | 8 | 8 | 6 | 7 | 7 | Lazy stream model is coherent but inherently complex. |
| `internal/evaluator/expression_compiler.go` | 8 | 9 | 8 | 8 | 8 | 8 | 8 | Focused expression compile bridge. |
| `internal/evaluator/modular_registry.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Strong builtin registration model. |
| `internal/evaluator/io_capabilities.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Clean capability abstraction. |
| `internal/evaluator/object_lambda_order.go` | 9 | 10 | 9 | 9 | 9 | 9 | 9 | Small, focused order tracking. |
| `internal/evaluator/assertions.go` | 7 | 6 | 7 | 8 | 7 | 7 | 7 | Useful but moderately dense assertion helpers. |
| `internal/evaluator/builtins_utils.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Practical shared builtin utilities. |
| `internal/evaluator/builtins_core.go` | 5 | 4 | 6 | 4 | 4 | 6 | 5 | Too broad: lazy, lambdas, defaults, module calls, coercion, and control helpers. |
| `internal/evaluator/builtins_control_flow.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Good extraction from core behavior. |
| `internal/evaluator/builtins_collection.go` | 6 | 5 | 7 | 6 | 6 | 7 | 6 | Long mixed collection functions. |
| `internal/evaluator/builtins_collection_ops.go` | 6 | 5 | 7 | 6 | 6 | 7 | 6 | Long lambda collection operations. |
| `internal/evaluator/builtins_collection_object_ops.go` | 7 | 6 | 7 | 7 | 6 | 7 | 7 | Object ops are more cohesive. |
| `internal/evaluator/builtins_arrays_collections.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Cohesive array helpers. |
| `internal/evaluator/builtins_arrays_lambda.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Cohesive lambda array helpers. |
| `internal/evaluator/builtins_arrays_reduce.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Focused reduction helpers. |
| `internal/evaluator/builtins_arrays_lazy.go` | 8 | 7 | 8 | 8 | 7 | 7 | 8 | Clear lazy array helpers. |
| `internal/evaluator/builtins_arrays_mapobject.go` | 7 | 6 | 7 | 7 | 6 | 7 | 7 | Useful but moderately coupled to object semantics. |
| `internal/evaluator/builtins_arrays_do.go` | 7 | 6 | 7 | 7 | 6 | 7 | 7 | Clear side-effect sequencing; runtime coupling expected. |
| `internal/evaluator/builtins_assert.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Focused assertion builtin. |
| `internal/evaluator/builtins_assertion.go` | 7 | 6 | 7 | 7 | 7 | 7 | 7 | Coherent assertion library surface. |
| `internal/evaluator/builtins_date.go` | 6 | 5 | 7 | 6 | 6 | 7 | 6 | Broad date parsing and formatting rules. |
| `internal/evaluator/builtins_io.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Good capability-based IO. |
| `internal/evaluator/builtins_math.go` | 7 | 6 | 7 | 7 | 7 | 7 | 7 | Clear enough; mixed numeric helpers. |
| `internal/evaluator/builtins_numbers_radix.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Focused radix functions. |
| `internal/evaluator/builtins_runtime.go` | 7 | 6 | 7 | 7 | 6 | 7 | 7 | Environment, logging, and try-style behavior share one file. |
| `internal/evaluator/builtins_string_basic.go` | 7 | 6 | 7 | 7 | 7 | 7 | 7 | Cohesive enough; broad string surface. |
| `internal/evaluator/builtins_string_encoding.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Focused encoding helpers. |
| `internal/evaluator/builtins_string_inflect.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Focused inflection helpers. |
| `internal/evaluator/builtins_string_regex.go` | 7 | 6 | 7 | 7 | 7 | 7 | 7 | Regex behavior is clear but edge-case heavy. |
| `internal/evaluator/builtins_string_text.go` | 7 | 6 | 7 | 7 | 7 | 7 | 7 | Text helpers are readable. |
| `internal/evaluator/builtins_update.go` | 7 | 5 | 7 | 7 | 6 | 7 | 6 | Update expressions are inherently complex. |
| `internal/evaluator/builtins_url.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Focused URL helpers. |
| `internal/evaluator/builtin_specs_assertion.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_collections.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_core.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_date.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_io.go` | 9 | 10 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_numbers.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_runtime.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_strings.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |
| `internal/evaluator/builtin_specs_url.go` | 9 | 10 | 9 | 9 | 8 | 8 | 9 | Declarative builtin specs. |

### Formats

| File | Clarity | Complexity | Naming | Cohesion | Coupling | Testability | Overall | Notes |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `pkg/formats/types.go` | 9 | 10 | 9 | 10 | 9 | 9 | 9 | Clean interfaces. |
| `pkg/formats/registry.go` | 8 | 9 | 8 | 8 | 8 | 8 | 8 | Simple compatibility registry. |
| `pkg/formats/core/types.go` | 9 | 9 | 9 | 9 | 9 | 9 | 9 | Clear registry abstractions. |
| `pkg/formats/core/registry.go` | 9 | 8 | 9 | 9 | 8 | 8 | 9 | Synchronized registry is clean. |
| `pkg/formats/reader.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Good read dispatch. |
| `pkg/formats/writer.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Good write dispatch. |
| `pkg/formats/options_dispatch.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Clear option routing. |
| `pkg/formats/serialize.go` | 9 | 9 | 9 | 9 | 8 | 8 | 9 | Small shared serializer helper. |
| `pkg/formats/json.go` | 9 | 9 | 9 | 9 | 8 | 9 | 9 | Straightforward JSON support. |
| `pkg/formats/csv.go` | 8 | 8 | 8 | 9 | 8 | 8 | 8 | Cohesive CSV behavior. |
| `pkg/formats/ndjson.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Clear line-delimited JSON support. |
| `pkg/formats/yaml.go` | 8 | 9 | 8 | 8 | 8 | 8 | 8 | Thin YAML support. |
| `pkg/formats/text.go` | 7 | 10 | 8 | 8 | 8 | 6 | 7 | Reader only; output support gap. |
| `pkg/formats/binary.go` | 8 | 9 | 8 | 8 | 8 | 8 | 8 | Small binary support. |
| `pkg/formats/dw.go` | 8 | 10 | 8 | 8 | 8 | 8 | 8 | Intentional unsupported-format boundary. |
| `pkg/formats/avro.go` | 7 | 10 | 8 | 7 | 8 | 7 | 8 | Placeholder-style binary boundary. |
| `pkg/formats/excel.go` | 8 | 9 | 8 | 8 | 8 | 7 | 8 | Thin Excel entry. |
| `pkg/formats/excel_structured.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Good structured workbook conversion. |
| `pkg/formats/xlsx_workbook.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Cohesive workbook parsing. |
| `pkg/formats/xlsx_worksheet.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Cohesive worksheet parsing. |
| `pkg/formats/xlsx_cell.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Focused cell conversion. |
| `pkg/formats/xml.go` | 6 | 4 | 7 | 5 | 5 | 6 | 5 | Large and correctness-sensitive; output escaping bug. |
| `pkg/formats/xml_options.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Clear option parsing. |
| `pkg/formats/xml_state_machine.go` | 7 | 6 | 8 | 8 | 7 | 7 | 7 | Useful validation, but overlaps with XML decoder concerns. |
| `pkg/formats/flatfile.go` | 7 | 6 | 7 | 8 | 7 | 7 | 7 | Clear goal, format-specific complexity. |
| `pkg/formats/java.go` | 7 | 6 | 7 | 7 | 7 | 7 | 7 | Structured surrogate for Java serialization. |
| `pkg/formats/multipart.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Good multipart boundary. |
| `pkg/formats/properties.go` | 8 | 7 | 8 | 8 | 8 | 8 | 8 | Cohesive properties support. |
| `pkg/formats/urlencoded.go` | 8 | 8 | 8 | 8 | 8 | 8 | 8 | Focused URL-encoded support. |
| `pkg/formats/protobuf.go` | 8 | 10 | 8 | 8 | 8 | 8 | 8 | Clear public entry. |
| `pkg/formats/protobuf_options.go` | 8 | 7 | 8 | 8 | 7 | 8 | 8 | Good option extraction. |
| `pkg/formats/protobuf_schema.go` | 8 | 6 | 8 | 8 | 7 | 8 | 7 | Schema parsing is dense but contained. |
| `pkg/formats/protobuf_wire.go` | 8 | 6 | 8 | 8 | 7 | 8 | 7 | Wire rules are complex but isolated. |
| `pkg/formats/protobuf_decode.go` | 8 | 6 | 8 | 8 | 7 | 8 | 7 | Decoder complexity is local. |
| `pkg/formats/protobuf_encode.go` | 8 | 6 | 8 | 8 | 7 | 8 | 7 | Encoder complexity is local. |

## Highest-Leverage Refactors

1. Fix the two P1 correctness/security findings before broad refactors.
2. Split preprocessor transforms by syntax family and migrate more transformations to exact mappings.
3. Consolidate declaration parsing around `declaration_ir.go` and remove unused compatibility handlers.
4. Split `builtins_core.go` into smaller files by behavior: module calls, coercion/defaults, lambdas/lazy, and control helpers.
5. Split `pkg/formats/xml.go` into reader, writer, option application, and tree conversion units after fixing escaping.


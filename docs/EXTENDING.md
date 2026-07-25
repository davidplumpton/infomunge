# Extending InfoMunge

This guide lists the registration points and tests that keep an extension wired
into the same execution paths as the built-in features.

## Add an Operator or Syntax Transform

1. Implement the rewrite in the appropriate
   `internal/preprocessor/transformers_*.go` file. Shared binary-operator
   scanning and configuration live in
   `internal/preprocessor/collection_operators.go` and
   `internal/preprocessor/operator_scanners.go`.
2. Add a transform contract to the appropriate contract list in
   `internal/preprocessor/stages.go`. The full stage order is declared by
   `CreateFullPreprocessingPipelineWithOptions`; post-rewriter stages are
   declared by `postProcessingStages`.
3. Choose every contract property deliberately:
   - `Phase` selects the stage. Stage order is significant because an earlier
     rewrite can change what later scanners see.
   - `Order` controls execution within a phase; lower values run first.
   - `Loop` is `TransformLoopOnce` or `TransformLoopFixpoint`. Use a fixpoint
     only when nested or repeated occurrences require another pass.
   - `Mapping` is `TransformMappingExact` or `TransformMappingInferred`. Exact
     handlers must return one source offset for every output byte. Prefer an
     exact mapping for operators and other position-sensitive rewrites;
     inferred mapping is suitable only when `inferStageMapping` can describe
     the rewrite without losing useful error locations.
   - Binary operators must also declare precedence and associativity. Reuse
     `configuredBinaryOperatorTransform` when its scanner matches the syntax.
4. Keep the lifecycle and contract definitions in
   `internal/preprocessor/transform_contract.go` in sync if a new phase or
   contract mode is required. A new phase must also be placed explicitly in
   the full and, when applicable, post-processing stage lists.
5. Add focused unit coverage in
   `internal/preprocessor/transformers_test.go` or a nearby transformer test,
   and contract/order/mapping coverage in
   `internal/preprocessor/transform_contract_test.go`. Add end-to-end Cucumber
   coverage under `test/features/*operator*.feature` or the closest syntax
   feature.

Mapping is a correctness requirement, not only diagnostic metadata. Each
handler's local mapping is composed with earlier mappings, and the pipeline
rejects a result whose byte length differs from its mapping length. Tests for a
new transform should therefore exercise transformed output and an error
position after the transform.

## Add a Builtin Function

1. Implement the handler in the appropriate
   `internal/evaluator/builtins_*.go` file. Regular handlers receive evaluated
   arguments; special handlers receive the AST and are responsible for
   evaluation semantics such as laziness or control flow.
2. Add a `BuiltinSpec` to the matching
   `internal/evaluator/builtin_specs_*.go` group using
   `regularBuiltinSpec` or `specialBuiltinSpec`. The spec is the source of
   dispatch and metadata, so declare the name, category/module, arity,
   evaluation mode, handler, and summary there rather than editing registry
   maps directly. Registry construction enforces the declared arity before
   invoking the handler, so optional arguments and variadic bounds must be
   represented accurately in the spec.
3. If introducing a new spec group, add that group to `defaultBuiltinSpecs` in
   `internal/evaluator/modular_registry.go`. Existing groups need no additional
   central registration.
4. Add handler behavior tests in
   `internal/evaluator/builtins_test.go` or a focused neighboring test. Add
   registry metadata and validation tests in
   `internal/evaluator/modular_registry_test.go`, then add user-facing Cucumber
   coverage in the relevant `test/features/*.feature` file.

If the builtin also needs a DataWeave-compatible module wrapper or composition,
update the relevant `modules/dw/core/*.im` script and cover both direct and
module-qualified calls.

## Add a Format

1. Add the codec implementation and its `init` function in a format-local file
   under `pkg/formats/*.go`.
2. In that local `init`, register each supported MIME type with
   `RegisterReader`, `RegisterWriter`, `RegisterObjectReader`, or
   `RegisterArrayReader` as appropriate, and register filename extensions with
   `RegisterExtension`. These APIs are exposed by
   `pkg/formats/registry.go`; built-in codecs do not add entries to a central
   format list.
3. If the codec accepts header options, implement option-aware handlers and
   register them with `RegisterReadOptionsHandler` and/or
   `RegisterWriteOptionsHandler`. Register equivalent MIME types with
   `RegisterOptionsAlias`. The option dispatch API is in
   `pkg/formats/options_dispatch.go`.
4. Add parser tests in `pkg/formats/reader_test.go`, serializer tests in
   `pkg/formats/writer_test.go`, and focused codec tests where useful. Cover
   registry or option-alias behavior in `pkg/formats/registry_test.go`. Add
   input/output Cucumber coverage in the relevant
   `test/features/*.feature` file.

Reader and writer support are independent: input-only formats should not
register a writer. Register aliases and option handlers for every MIME type the
runner can select, and verify both the canonical MIME type and aliases.

## Change Module Loading Behavior

Update `internal/runner/module_loader.go` to change module search paths or
import resolution. Cover the behavior in runner unit tests and
`test/features/import_directive.feature`.

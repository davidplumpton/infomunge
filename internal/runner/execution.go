package runner

import (
	"context"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
	"infomunge/internal/output"
	"infomunge/internal/preprocessor"
	"infomunge/internal/sourcemap"
	"os"
	"strings"
)

// ExecutionResult is the structured result of evaluating a script.
// It intentionally carries the unformatted value plus header metadata so callers
// can decide whether to format, serialize, or inspect the value directly.
type ExecutionResult struct {
	Value          evaluator.Value
	HasHeader      bool
	OutputMimeType string
	Context        evaluator.Context
	OutputMetadata output.Metadata
}

// Resolved returns a copy of the execution result with lazy and stream values
// materialized.
func (r ExecutionResult) Resolved() (ExecutionResult, error) {
	value, err := ResolveResult(r.Value)
	if err != nil {
		return ExecutionResult{}, err
	}
	r.Value = value
	return r, nil
}

// ExecuteString evaluates a script and returns the structured result without
// formatting or printing it. Adapter layers own formatting and output side
// effects.
func ExecuteString(goCtx context.Context, script string, additionalContext evaluator.Context, opts RunnerOptions) (ExecutionResult, error) {
	if goCtx == nil {
		goCtx = context.Background()
	}
	if opts.BaseDir == "" {
		baseDir, err := os.Getwd()
		if err != nil {
			opts.BaseDir = "."
		} else {
			opts.BaseDir = baseDir
		}
	}
	return executeWithConfig(goCtx, script, additionalContext, opts)
}

func executeWithConfig(goCtx context.Context, raw string, additionalContext evaluator.Context, opts RunnerOptions) (ExecutionResult, error) {
	if opts.Lazy {
		return ExecutionResult{}, unifiederrors.ValidationError(lazyFlagUnsupportedMessage)
	}

	header, body, bodyOffset := preprocessor.ExtractHeaderAndBody(raw)
	hasHeader := bodyOffset != 0

	loader := NewModuleLoader(opts.BaseDir)
	loader.Options = opts
	evalScope, outputMimeType, outputMetadata, err := parseHeaderWithGoContextAndOptions(header, hasHeader, goCtx, raw, loader, opts)
	if err != nil {
		return ExecutionResult{HasHeader: hasHeader, OutputMimeType: outputMimeType}, err
	}

	for k, v := range additionalContext {
		if evaluator.IsReservedBindingName(k) {
			return ExecutionResult{HasHeader: hasHeader, OutputMimeType: outputMimeType, Context: evalScope.Vars, OutputMetadata: outputMetadata}, unifiederrors.ValidationErrorf("binding name %q is reserved for runtime metadata", k)
		}
		evalScope.Vars[k] = v
	}
	evalScope = installEvaluationCapabilities(evalScope, opts)

	prepOpts := preprocessor.Options{}
	if strings.ContainsAny(body, "\n\r") {
		prepOpts.AllowMultilineIfElse = true
	}
	parseableExpr, mapping, err := preprocessor.PrepareForParsing(body, prepOpts)
	if err != nil {
		return ExecutionResult{HasHeader: hasHeader, OutputMimeType: outputMimeType, Context: evalScope.Vars, OutputMetadata: outputMetadata}, err
	}
	bodyMap := sourcemap.Identity(raw).SliceSource(bodyOffset, bodyOffset+len(body)).Compose(parseableExpr, mapping)
	value, err := evaluator.EvaluateWithScopeAndContext(parseableExpr, evalScope, &evaluator.ErrorContext{SourceMap: bodyMap})
	if err != nil {
		return ExecutionResult{HasHeader: hasHeader, OutputMimeType: outputMimeType, Context: evalScope.Vars, OutputMetadata: outputMetadata}, err
	}

	return ExecutionResult{
		Value:          value,
		HasHeader:      hasHeader,
		OutputMimeType: outputMimeType,
		Context:        evalScope.Vars,
		OutputMetadata: outputMetadata,
	}, nil
}

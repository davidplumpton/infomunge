package runner

import (
	"context"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
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

// ExecuteString evaluates a script without formatting or printing the result.
func ExecuteString(script string, additionalContext evaluator.Context) (ExecutionResult, error) {
	return ExecuteStringWithGoContext(context.Background(), script, additionalContext)
}

// ExecuteStringWithGoContext evaluates a script with a Go context.
func ExecuteStringWithGoContext(goCtx context.Context, script string, additionalContext evaluator.Context) (ExecutionResult, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}
	return ExecuteStringWithGoContextAndOptions(goCtx, script, additionalContext, RunnerOptions{BaseDir: baseDir})
}

// ExecuteStringWithContextAndOptions evaluates a script with runner options.
func ExecuteStringWithContextAndOptions(script string, additionalContext evaluator.Context, opts RunnerOptions) (ExecutionResult, error) {
	return ExecuteStringWithGoContextAndOptions(context.Background(), script, additionalContext, opts)
}

// ExecuteStringWithGoContextAndOptions evaluates a script with runner options and a Go context.
func ExecuteStringWithGoContextAndOptions(goCtx context.Context, script string, additionalContext evaluator.Context, opts RunnerOptions) (ExecutionResult, error) {
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
	evalContext, outputMimeType, err := parseHeaderWithGoContext(header, hasHeader, goCtx, raw, loader)
	if err != nil {
		return ExecutionResult{HasHeader: hasHeader, OutputMimeType: outputMimeType}, err
	}

	for k, v := range additionalContext {
		evalContext[k] = v
	}

	prepOpts := preprocessor.Options{}
	if strings.ContainsAny(body, "\n\r") {
		prepOpts.AllowMultilineIfElse = true
	}
	parseableExpr, mapping, err := preprocessor.PrepareForParsing(body, prepOpts)
	if err != nil {
		return ExecutionResult{HasHeader: hasHeader, OutputMimeType: outputMimeType, Context: evalContext}, err
	}
	bodyMap := sourcemap.Identity(raw).SliceSource(bodyOffset, bodyOffset+len(body)).Compose(parseableExpr, mapping)
	value, err := evaluator.EvaluateWithGoContextAndSourceMap(parseableExpr, evalContext, goCtx, bodyMap)
	if err != nil {
		return ExecutionResult{HasHeader: hasHeader, OutputMimeType: outputMimeType, Context: evalContext}, err
	}

	return ExecutionResult{
		Value:          value,
		HasHeader:      hasHeader,
		OutputMimeType: outputMimeType,
		Context:        evalContext,
	}, nil
}

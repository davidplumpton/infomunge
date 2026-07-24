package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"infomunge/internal/evaluator"
)

// Run executes the infomunge process on the given file.
//
// Deprecated: use ReadScriptFile, ExecuteString, and FormatExecutionResult in
// the caller-owned adapter instead.
func Run(filePath string) error {
	return RunWithConfig(filePath, RunnerOptions{})
}

// RunWithConfig executes the infomunge process on the given file with options.
//
// Deprecated: use ReadScriptFile, ExecuteString, and FormatExecutionResult in
// the caller-owned adapter instead.
func RunWithConfig(filePath string, opts RunnerOptions) error {
	content, err := ReadScriptFile(filePath)
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	if opts.BaseDir == "" {
		opts.BaseDir = filepath.Dir(absPath)
	}

	return executeAndPrintString(context.Background(), content, nil, opts)
}

// RunFromStringWithContext executes an infomunge script with additional context variables.
//
// Deprecated: use ExecuteString and FormatExecutionResult in the caller-owned
// adapter instead.
func RunFromStringWithContext(raw string, additionalContext evaluator.Context) error {
	return RunFromStringWithContextAndOptionsAndGoContext(context.Background(), raw, additionalContext, RunnerOptions{})
}

// RunFromStringWithContextAndOptionsAndGoContext executes an infomunge script with additional context variables, options, and Go context.
//
// Deprecated: use ExecuteString and FormatExecutionResult in the caller-owned
// adapter instead.
func RunFromStringWithContextAndOptionsAndGoContext(goCtx context.Context, raw string, additionalContext evaluator.Context, opts RunnerOptions) error {
	return executeAndPrintString(goCtx, raw, additionalContext, opts)
}

func executeAndPrintString(goCtx context.Context, raw string, additionalContext evaluator.Context, opts RunnerOptions) error {
	if err := RequireScriptHeader(raw); err != nil {
		return err
	}

	result, err := ExecuteString(goCtx, raw, additionalContext, opts)
	if err != nil {
		return err
	}

	return PrintExecutionResult(result)
}

// PrintExecutionResult formats and writes an execution result to stdout.
//
// Deprecated: use FormatExecutionResult and write output in the caller-owned
// adapter instead.
func PrintExecutionResult(result ExecutionResult) error {
	formatted, err := FormatExecutionResult(result)
	if err != nil {
		return err
	}
	fmt.Print(formatted)
	return nil
}

// RunString executes an infomunge script from a string with optional additional context
// and returns only the resolved value.
//
// Deprecated: use ExecuteString and ExecutionResult.Resolved instead.
func RunString(script string, additionalContext evaluator.Context) (evaluator.Value, error) {
	return RunStringWithGoContext(context.Background(), script, additionalContext)
}

// RunStringWithGoContext executes an infomunge script with optional additional
// context and Go context, returning only the resolved value.
//
// Deprecated: use ExecuteString and ExecutionResult.Resolved instead.
func RunStringWithGoContext(goCtx context.Context, script string, additionalContext evaluator.Context) (evaluator.Value, error) {
	result, err := ExecuteString(goCtx, script, additionalContext, RunnerOptions{})
	if err != nil {
		return nil, err
	}
	resolved, err := result.Resolved()
	if err != nil {
		return nil, err
	}
	return resolved.Value, nil
}

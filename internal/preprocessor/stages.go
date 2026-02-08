package preprocessor

import (
	unifiederrors "infomunge/internal/errors"
)

// transformHandler defines the signature for transformation functions
type transformHandler func(string) string

// PipelineStage represents a group of related transformations
type PipelineStage interface {
	Name() string
	Execute(input string) (string, error)
}

// basicStage is a simple pipeline stage implementation
type basicStage struct {
	name     string
	handlers []transformHandler
}

func (bs *basicStage) Name() string {
	return bs.name
}

func (bs *basicStage) Execute(input string) (string, error) {
	result := input
	for _, handler := range bs.handlers {
		result = handler(result)
	}
	return result, nil
}

// errorAwareHandler is a transformation function that can return errors
type errorAwareHandler func(string) (string, error)

// errorAwareStage is a pipeline stage that supports error-returning handlers
type errorAwareStage struct {
	name    string
	handler errorAwareHandler
}

func (es *errorAwareStage) Name() string {
	return es.name
}

func (es *errorAwareStage) Execute(input string) (string, error) {
	return es.handler(input)
}

// Create stages for different transformation groups

// CreateRegexLiteralStage creates pipeline for regex literal transformations.
// This must run before other stages that might misinterpret slashes.
func CreateRegexLiteralStage() PipelineStage {
	return &basicStage{
		name: "Regex Literal Processing",
		handlers: []transformHandler{
			replaceRegexLiterals,
		},
	}
}

// CreateStringProcessingStage creates pipeline for string-related transformations
func CreateStringProcessingStage() PipelineStage {
	return &basicStage{
		name: "String Processing",
		handlers: []transformHandler{
			replaceStringInterpolation,
			replaceArrayRangeIndexing,
		},
	}
}

// CreateOperatorProcessingStage creates pipeline for operator transformations
func CreateOperatorProcessingStage() PipelineStage {
	return &errorAwareStage{
		name: "Operator Processing",
		handler: func(s string) (string, error) {
			// Apply looping operators first
			transforms := []struct {
				name string
				fn   errorAwareHandler
			}{
				{"replaceDefaultOperator", replaceDefaultOperatorErr},
				{"replaceOnNullOperator", replaceOnNullOperatorErr},
				{"replaceThenOperator", replaceThenOperatorErr},
				{"replaceToOperator", replaceToOperatorErr},
			}
			result := s
			for _, t := range transforms {
				prevResult := ""
				iterCount := 0
				for result != prevResult {
					if iterCount > MaxTransformIterations {
						return result, unifiederrors.ParseErrorf("infinite loop detected in %s", t.name)
					}
					prevResult = result
					nextResult, err := t.fn(result)
					if err != nil {
						return result, err
					}
					result = nextResult
					iterCount++
				}
			}
			return result, nil
		},
	}
}

// CreateFunctionalProcessingStage creates pipeline for functional transformations
func CreateFunctionalProcessingStage() PipelineStage {
	return &errorAwareStage{
		name: "Functional Processing",
		handler: func(s string) (string, error) {
			wrap := func(fn transformHandler) errorAwareHandler {
				return func(input string) (string, error) {
					return fn(input), nil
				}
			}

			// Apply looping functional transformations
			transforms := []struct {
				name string
				fn   errorAwareHandler
			}{
				{"replaceImplicitLambdas", wrap(replaceImplicitLambdas)},
				{"replaceModuleCall", wrap(replaceModuleCall)},
				{"replaceCaseStatements", wrap(replaceCaseStatements)},
				{"replaceArrowFunctions", wrap(replaceArrowFunctions)},
				{"replaceFilterOperator", wrap(replaceFilterOperator)},
				{"replaceMapOperator", wrap(replaceMapOperator)},
				{"replaceReduceOperator", wrap(replaceReduceOperator)},
				{"replaceGroupByOperator", wrap(replaceGroupByOperator)},
				{"replacePluckOperator", wrap(replacePluckOperator)},
				{"replaceFlatMapOperator", wrap(replaceFlatMapOperator)},
				{"replaceMaxByOperator", wrap(replaceMaxByOperator)},
				{"replaceMinByOperator", wrap(replaceMinByOperator)},
				{"replaceOrderByOperator", wrap(replaceOrderByOperator)},
				{"replaceSortOperator", wrap(replaceSortOperator)},
				{"replaceDistinctByOperator", wrap(replaceDistinctByOperator)},
				{"replaceFilterObjectOperator", wrap(replaceFilterObjectOperator)},
				{"replaceMapObjectOperator", wrap(replaceMapObjectOperator)},
				{"replaceUpdateOperator", replaceUpdateOperatorErr},
				{"replaceAsOperator", wrap(replaceAsOperator)},
				{"replaceIsOperator", wrap(replaceIsOperator)},
				{"replaceFindOperator", replaceFindOperatorErr},
				{"replaceContainsOperator", replaceContainsOperatorErr},
				{"replaceSplitByOperator", replaceSplitByOperatorErr},
				{"replaceJoinByOperator", replaceJoinByOperatorErr},
				{"replaceConcatenateOperator", replaceConcatenateOperatorErr},
				{"replaceRemoveOperator", replaceRemoveOperatorErr},
				{"replaceExponentOperator", wrap(replaceExponentOperator)},
				{"replaceMatchOperator", replaceMatchOperatorErr},
				{"replaceMatchesOperator", replaceMatchesOperatorErr},
				{"replaceModOperator", replaceModOperatorErr},
				{"replaceRepeatOperator", replaceRepeatOperatorErr},
				{"replaceSubstringOperator", wrap(replaceSubstringOperator)},
				{"replaceContainsMethodCall", wrap(replaceContainsMethodCall)},
				{"replaceFindMethodCall", wrap(replaceFindMethodCall)},
				{"replaceMatchMethodCall", wrap(replaceMatchMethodCall)},
				{"replaceMatchesMethodCall", wrap(replaceMatchesMethodCall)},
				{"replaceScanMethodCall", wrap(replaceScanMethodCall)},
				{"replaceSplitByMethodCall", wrap(replaceSplitByMethodCall)},
				{"replacePipeToFunctionOperator", wrap(replacePipeToFunctionOperator)},
				{"replaceReplaceOperator", wrap(replaceReplaceOperator)},
				{"replaceAssignmentExpressions", wrap(replaceAssignmentExpressions)},
			}
			result := s
			for _, t := range transforms {
				prevResult := ""
				iterCount := 0
				for result != prevResult {
					if iterCount > MaxTransformIterations {
						return result, unifiederrors.ParseErrorf("infinite loop detected in %s", t.name)
					}
					prevResult = result
					nextResult, err := t.fn(result)
					if err != nil {
						return result, err
					}
					result = nextResult
					iterCount++
				}
			}
			return result, nil
		},
	}
}

// CreateSelectorProcessingStage creates pipeline for selector transformations (.* and ..)
// This must run before operator processing so that selectors bind tightly.
func CreateSelectorProcessingStage() PipelineStage {
	return &basicStage{
		name: "Selector Processing",
		handlers: []transformHandler{
			replaceFilterSelectors,
			replaceMetadataSelectors,
			replaceRecursiveDescent,
		},
	}
}

// CreateSyntaxProcessingStage creates pipeline for syntax transformations
func CreateSyntaxProcessingStage() PipelineStage {
	return &basicStage{
		name: "Syntax Processing",
		handlers: []transformHandler{
			replaceDotNotation,
			replaceKeyAttributes,
			replaceMultiStatementSequences,
		},
	}
}

// CreateModularPostProcessingPipeline builds pipeline using stages
func CreateModularPostProcessingPipeline() *ModularPipeline {
	stages := []PipelineStage{
		// Note: Regex literal processing is done in PrepareForParsing before the rewriter,
		// so we don't include it here.
		CreateStringProcessingStage(),
		CreateOperatorProcessingStage(),
		CreateSelectorProcessingStage(), // .* and .. selectors must run before functional processing
		CreateFunctionalProcessingStage(),
		CreateSyntaxProcessingStage(),
	}

	return NewModularPipeline(stages)
}

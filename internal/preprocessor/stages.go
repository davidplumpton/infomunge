package preprocessor

import (
	unifiederrors "infomunge/internal/errors"
)

// transformHandler defines the signature for transformation functions
type transformHandler func(string) string

type mappedTransformHandler func(string) (string, []int)

// PipelineStage represents a group of related transformations
type PipelineStage interface {
	Name() string
	Execute(input string, mapping []int) (string, []int, error)
}

// basicStage is a simple pipeline stage implementation
type basicStage struct {
	name     string
	handlers []mappedTransformHandler
}

func (bs *basicStage) Name() string {
	return bs.name
}

func (bs *basicStage) Execute(input string, mapping []int) (string, []int, error) {
	result := input
	resultMapping := mapping
	for _, handler := range bs.handlers {
		var local []int
		result, local = handler(result)
		resultMapping = composeMappings(resultMapping, local)
	}
	return result, resultMapping, nil
}

// errorAwareHandler is a transformation function that can return errors
type errorAwareHandler func(string) (string, error)

type mappedErrorAwareHandler func(string) (string, []int, error)

// errorAwareStage is a pipeline stage that supports error-returning handlers
type errorAwareStage struct {
	name    string
	handler mappedErrorAwareHandler
}

func (es *errorAwareStage) Name() string {
	return es.name
}

func (es *errorAwareStage) Execute(input string, mapping []int) (string, []int, error) {
	result, local, err := es.handler(input)
	if err != nil {
		return result, mapping, err
	}
	return result, composeMappings(mapping, local), nil
}

func inferMappedTransform(fn transformHandler) mappedTransformHandler {
	return func(input string) (string, []int) {
		result := fn(input)
		return result, inferStageMapping(input, result)
	}
}

func exactMappedTransform(fn mappedTransformHandler) mappedTransformHandler {
	return fn
}

func inferMappedErrorTransform(fn errorAwareHandler) mappedErrorAwareHandler {
	return func(input string) (string, []int, error) {
		result, err := fn(input)
		if err != nil {
			return result, nil, err
		}
		return result, inferStageMapping(input, result), nil
	}
}

// Create stages for different transformation groups

// CreateRegexLiteralStage creates pipeline for regex literal transformations.
// This must run before other stages that might misinterpret slashes.
func CreateRegexLiteralStage() PipelineStage {
	return &basicStage{
		name: "Regex Literal Processing",
		handlers: []mappedTransformHandler{
			exactMappedTransform(replaceRegexLiteralsWithMapping),
		},
	}
}

// CreateStringProcessingStage creates pipeline for string-related transformations
func CreateStringProcessingStage() PipelineStage {
	return &basicStage{
		name: "String Processing",
		handlers: []mappedTransformHandler{
			inferMappedTransform(replaceStringInterpolation),
			inferMappedTransform(replaceArrayRangeIndexing),
		},
	}
}

// CreateOperatorProcessingStage creates pipeline for operator transformations
func CreateOperatorProcessingStage() PipelineStage {
	return &errorAwareStage{
		name: "Operator Processing",
		handler: func(s string) (string, []int, error) {
			// Apply looping operators first
			transforms := []struct {
				name string
				fn   mappedErrorAwareHandler
			}{
				{"replaceDefaultOperator", replaceDefaultOperatorErr},
				{"replaceOnNullOperator", replaceOnNullOperatorErr},
				{"replaceThenOperator", replaceThenOperatorErr},
				{"replaceToOperator", replaceToOperatorErr},
			}
			result := s
			resultMapping := identityMapping(len(s))
			for _, t := range transforms {
				prevResult := ""
				iterCount := 0
				for result != prevResult {
					if iterCount > MaxTransformIterations {
						return result, resultMapping, unifiederrors.ParseErrorf("infinite loop detected in %s", t.name)
					}
					prevResult = result
					nextResult, local, err := t.fn(result)
					if err != nil {
						return result, resultMapping, err
					}
					result = nextResult
					resultMapping = composeMappings(resultMapping, local)
					iterCount++
				}
			}
			return result, resultMapping, nil
		},
	}
}

// CreateFunctionalProcessingStage creates pipeline for functional transformations
func CreateFunctionalProcessingStage() PipelineStage {
	return &errorAwareStage{
		name: "Functional Processing",
		handler: func(s string) (string, []int, error) {
			wrap := func(fn transformHandler) mappedErrorAwareHandler {
				return func(input string) (string, []int, error) {
					result := fn(input)
					return result, inferStageMapping(input, result), nil
				}
			}

			// Apply looping functional transformations
			transforms := []struct {
				name string
				fn   mappedErrorAwareHandler
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
				{"replaceUpdateOperator", inferMappedErrorTransform(replaceUpdateOperatorErr)},
				{"replaceAsOperator", wrap(replaceAsOperator)},
				{"replaceIsOperator", wrap(replaceIsOperator)},
				{"replaceFindOperator", inferMappedErrorTransform(replaceFindOperatorErr)},
				{"replaceContainsOperator", inferMappedErrorTransform(replaceContainsOperatorErr)},
				{"replaceSplitByOperator", inferMappedErrorTransform(replaceSplitByOperatorErr)},
				{"replaceJoinByOperator", inferMappedErrorTransform(replaceJoinByOperatorErr)},
				{"replaceConcatenateOperator", inferMappedErrorTransform(replaceConcatenateOperatorErr)},
				{"replaceRemoveOperator", inferMappedErrorTransform(replaceRemoveOperatorErr)},
				{"replaceExponentOperator", wrap(replaceExponentOperator)},
				{"replaceMatchOperator", inferMappedErrorTransform(replaceMatchOperatorErr)},
				{"replaceMatchesOperator", inferMappedErrorTransform(replaceMatchesOperatorErr)},
				{"replaceModOperator", inferMappedErrorTransform(replaceModOperatorErr)},
				{"replaceRepeatOperator", inferMappedErrorTransform(replaceRepeatOperatorErr)},
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
			resultMapping := identityMapping(len(s))
			for _, t := range transforms {
				prevResult := ""
				iterCount := 0
				for result != prevResult {
					if iterCount > MaxTransformIterations {
						return result, resultMapping, unifiederrors.ParseErrorf("infinite loop detected in %s", t.name)
					}
					prevResult = result
					nextResult, local, err := t.fn(result)
					if err != nil {
						return result, resultMapping, err
					}
					result = nextResult
					resultMapping = composeMappings(resultMapping, local)
					iterCount++
				}
			}
			return result, resultMapping, nil
		},
	}
}

// CreateSelectorProcessingStage creates pipeline for selector transformations (.* and ..)
// This must run before operator processing so that selectors bind tightly.
func CreateSelectorProcessingStage() PipelineStage {
	return &basicStage{
		name: "Selector Processing",
		handlers: []mappedTransformHandler{
			inferMappedTransform(replaceFilterSelectors),
			inferMappedTransform(replaceMetadataSelectors),
			exactMappedTransform(replaceRecursiveDescentWithMapping),
		},
	}
}

// CreateSyntaxProcessingStage creates pipeline for syntax transformations
func CreateSyntaxProcessingStage() PipelineStage {
	return &basicStage{
		name: "Syntax Processing",
		handlers: []mappedTransformHandler{
			exactMappedTransform(replaceDotNotationWithMapping),
			inferMappedTransform(replaceKeyAttributes),
			inferMappedTransform(replaceMultiStatementSequences),
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

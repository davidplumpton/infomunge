package preprocessor

// transformHandler defines the signature for transformation functions.
type transformHandler func(string) string

type mappedTransformHandler func(string) (string, []int)

// PipelineStage represents a group of related transformations.
type PipelineStage interface {
	Name() string
	Execute(input string, mapping []int) (string, []int, error)
}

// errorAwareHandler is a transformation function that can return errors.
type errorAwareHandler func(string) (string, error)

type mappedErrorAwareHandler func(string) (string, []int, error)

func inferredTransform(name string, phase TransformPhase, order int, loop TransformLoopMode, fn transformHandler) TransformContract {
	return TransformContract{
		Name:          name,
		Phase:         phase,
		Order:         order,
		Associativity: TransformAssociativityNone,
		Mapping:       TransformMappingInferred,
		Loop:          loop,
		Handler: func(input string) (string, []int, error) {
			result := fn(input)
			return result, inferStageMapping(input, result), nil
		},
	}
}

func exactTransform(name string, phase TransformPhase, order int, loop TransformLoopMode, fn mappedTransformHandler) TransformContract {
	return TransformContract{
		Name:          name,
		Phase:         phase,
		Order:         order,
		Associativity: TransformAssociativityNone,
		Mapping:       TransformMappingExact,
		Loop:          loop,
		Handler: func(input string) (string, []int, error) {
			result, mapping := fn(input)
			return result, mapping, nil
		},
	}
}

func inferredErrorTransform(name string, phase TransformPhase, order int, loop TransformLoopMode, fn errorAwareHandler) TransformContract {
	return TransformContract{
		Name:          name,
		Phase:         phase,
		Order:         order,
		Associativity: TransformAssociativityNone,
		Mapping:       TransformMappingInferred,
		Loop:          loop,
		Handler: func(input string) (string, []int, error) {
			result, err := fn(input)
			if err != nil {
				return result, nil, err
			}
			return result, inferStageMapping(input, result), nil
		},
	}
}

func configuredBinaryOperatorTransform(name string, phase TransformPhase, order int, key string, precedence int, associativity TransformAssociativity) TransformContract {
	return TransformContract{
		Name:          name,
		Phase:         phase,
		Order:         order,
		Precedence:    precedence,
		Associativity: associativity,
		Mapping:       TransformMappingExact,
		Loop:          TransformLoopFixpoint,
		Handler: func(input string) (string, []int, error) {
			return replaceConfiguredBinaryOperatorWithMapping(input, key)
		},
	}
}

// CreateRegexLiteralStage creates pipeline for regex literal transformations.
// This must run before other stages that might misinterpret slashes.
func CreateRegexLiteralStage() PipelineStage {
	return createRegexLiteralStage(nil)
}

func createRegexLiteralStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Regex Literal Processing", TransformPhaseRegex, trace, []TransformContract{
		exactTransform("replaceRegexLiteralsWithMapping", TransformPhaseRegex, 10, TransformLoopOnce, replaceRegexLiteralsWithMapping),
	})
}

// CreateStringProcessingStage creates pipeline for string-related transformations.
func CreateStringProcessingStage() PipelineStage {
	return createStringProcessingStage(nil)
}

func createStringProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("String Processing", TransformPhaseString, trace, []TransformContract{
		inferredTransform("replaceStringInterpolation", TransformPhaseString, 10, TransformLoopOnce, replaceStringInterpolation),
		inferredTransform("replaceArrayRangeIndexing", TransformPhaseString, 20, TransformLoopOnce, replaceArrayRangeIndexing),
	})
}

// CreateOperatorProcessingStage creates pipeline for low-precedence operator transformations.
func CreateOperatorProcessingStage() PipelineStage {
	return createOperatorProcessingStage(nil)
}

func createOperatorProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Operator Processing", TransformPhaseOperator, trace, operatorProcessingContracts())
}

func operatorProcessingContracts() []TransformContract {
	return []TransformContract{
		configuredBinaryOperatorTransform("replaceDefaultOperator", TransformPhaseOperator, 10, binaryOpDefault, TransformPrecedenceDefault, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceOnNullOperator", TransformPhaseOperator, 20, binaryOpOnNull, TransformPrecedenceNullChain, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceThenOperator", TransformPhaseOperator, 30, binaryOpThen, TransformPrecedenceNullChain, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceToOperator", TransformPhaseOperator, 40, binaryOpTo, TransformPrecedenceRange, TransformAssociativityLeft),
	}
}

// CreateFunctionalProcessingStage creates pipeline for functional transformations.
func CreateFunctionalProcessingStage() PipelineStage {
	return createFunctionalProcessingStage(nil)
}

func createFunctionalProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Functional Processing", TransformPhaseFunctional, trace, functionalProcessingContracts())
}

func functionalProcessingContracts() []TransformContract {
	return []TransformContract{
		inferredTransform("replaceImplicitLambdas", TransformPhaseFunctional, 10, TransformLoopFixpoint, replaceImplicitLambdas),
		inferredTransform("replaceModuleCall", TransformPhaseFunctional, 20, TransformLoopFixpoint, replaceModuleCall),
		inferredTransform("replaceCaseStatements", TransformPhaseFunctional, 30, TransformLoopFixpoint, replaceCaseStatements),
		inferredTransform("replaceArrowFunctions", TransformPhaseFunctional, 40, TransformLoopFixpoint, replaceArrowFunctions),
		inferredTransform("replaceFilterOperator", TransformPhaseFunctional, 50, TransformLoopFixpoint, replaceFilterOperator),
		inferredTransform("replaceMapOperator", TransformPhaseFunctional, 60, TransformLoopFixpoint, replaceMapOperator),
		inferredTransform("replaceReduceOperator", TransformPhaseFunctional, 70, TransformLoopFixpoint, replaceReduceOperator),
		inferredTransform("replaceGroupByOperator", TransformPhaseFunctional, 80, TransformLoopFixpoint, replaceGroupByOperator),
		inferredTransform("replacePluckOperator", TransformPhaseFunctional, 90, TransformLoopFixpoint, replacePluckOperator),
		inferredTransform("replaceFlatMapOperator", TransformPhaseFunctional, 100, TransformLoopFixpoint, replaceFlatMapOperator),
		inferredTransform("replaceMaxByOperator", TransformPhaseFunctional, 110, TransformLoopFixpoint, replaceMaxByOperator),
		inferredTransform("replaceMinByOperator", TransformPhaseFunctional, 120, TransformLoopFixpoint, replaceMinByOperator),
		inferredTransform("replaceOrderByOperator", TransformPhaseFunctional, 130, TransformLoopFixpoint, replaceOrderByOperator),
		inferredTransform("replaceSortOperator", TransformPhaseFunctional, 140, TransformLoopFixpoint, replaceSortOperator),
		inferredTransform("replaceDistinctByOperator", TransformPhaseFunctional, 150, TransformLoopFixpoint, replaceDistinctByOperator),
		inferredTransform("replaceFilterObjectOperator", TransformPhaseFunctional, 160, TransformLoopFixpoint, replaceFilterObjectOperator),
		inferredTransform("replaceMapObjectOperator", TransformPhaseFunctional, 170, TransformLoopFixpoint, replaceMapObjectOperator),
		configuredBinaryOperatorTransform("replaceUpdateOperator", TransformPhaseFunctional, 180, binaryOpUpdate, TransformPrecedenceCollection, TransformAssociativityLeft),
		inferredTransform("replaceAsOperator", TransformPhaseFunctional, 190, TransformLoopFixpoint, replaceAsOperator),
		inferredTransform("replaceIsOperator", TransformPhaseFunctional, 200, TransformLoopFixpoint, replaceIsOperator),
		configuredBinaryOperatorTransform("replaceFindOperator", TransformPhaseFunctional, 210, binaryOpFind, TransformPrecedenceCollection, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceContainsOperator", TransformPhaseFunctional, 220, binaryOpContains, TransformPrecedenceComparison, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceSplitByOperator", TransformPhaseFunctional, 230, binaryOpSplitBy, TransformPrecedenceCollection, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceJoinByOperator", TransformPhaseFunctional, 240, binaryOpJoinBy, TransformPrecedenceCollection, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceConcatenateOperator", TransformPhaseFunctional, 250, binaryOpConcatenate, TransformPrecedenceAdditive, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceRemoveOperator", TransformPhaseFunctional, 260, binaryOpRemove, TransformPrecedenceAdditive, TransformAssociativityLeft),
		inferredTransform("replaceExponentOperator", TransformPhaseFunctional, 270, TransformLoopFixpoint, replaceExponentOperator),
		configuredBinaryOperatorTransform("replaceMatchOperator", TransformPhaseFunctional, 280, binaryOpMatch, TransformPrecedenceComparison, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceMatchesOperator", TransformPhaseFunctional, 290, binaryOpMatches, TransformPrecedenceComparison, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceModOperator", TransformPhaseFunctional, 300, binaryOpMod, TransformPrecedenceMultiplicative, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceRepeatOperator", TransformPhaseFunctional, 310, binaryOpRepeat, TransformPrecedenceMultiplicative, TransformAssociativityLeft),
		inferredTransform("replaceSubstringOperator", TransformPhaseFunctional, 320, TransformLoopFixpoint, replaceSubstringOperator),
		inferredTransform("replaceContainsMethodCall", TransformPhaseFunctional, 330, TransformLoopFixpoint, replaceContainsMethodCall),
		inferredTransform("replaceFindMethodCall", TransformPhaseFunctional, 340, TransformLoopFixpoint, replaceFindMethodCall),
		inferredTransform("replaceMatchMethodCall", TransformPhaseFunctional, 350, TransformLoopFixpoint, replaceMatchMethodCall),
		inferredTransform("replaceMatchesMethodCall", TransformPhaseFunctional, 360, TransformLoopFixpoint, replaceMatchesMethodCall),
		inferredTransform("replaceScanMethodCall", TransformPhaseFunctional, 370, TransformLoopFixpoint, replaceScanMethodCall),
		inferredTransform("replaceSplitByMethodCall", TransformPhaseFunctional, 380, TransformLoopFixpoint, replaceSplitByMethodCall),
		inferredTransform("replacePipeToFunctionOperator", TransformPhaseFunctional, 390, TransformLoopFixpoint, replacePipeToFunctionOperator),
		inferredTransform("replaceReplaceOperator", TransformPhaseFunctional, 400, TransformLoopFixpoint, replaceReplaceOperator),
		inferredTransform("replaceAssignmentExpressions", TransformPhaseFunctional, 410, TransformLoopFixpoint, replaceAssignmentExpressions),
	}
}

// CreateSelectorProcessingStage creates pipeline for selector transformations (.* and ..).
// This must run before functional processing so selectors bind tightly.
func CreateSelectorProcessingStage() PipelineStage {
	return createSelectorProcessingStage(nil)
}

func createSelectorProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Selector Processing", TransformPhaseSelector, trace, []TransformContract{
		inferredTransform("replaceFilterSelectors", TransformPhaseSelector, 10, TransformLoopOnce, replaceFilterSelectors),
		inferredTransform("replaceMetadataSelectors", TransformPhaseSelector, 20, TransformLoopOnce, replaceMetadataSelectors),
		exactTransform("replaceRecursiveDescentWithMapping", TransformPhaseSelector, 30, TransformLoopOnce, replaceRecursiveDescentWithMapping),
	})
}

// CreateSyntaxProcessingStage creates pipeline for syntax transformations.
func CreateSyntaxProcessingStage() PipelineStage {
	return createSyntaxProcessingStage(nil)
}

func createSyntaxProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Syntax Processing", TransformPhaseSyntax, trace, []TransformContract{
		exactTransform("replaceDotNotationWithMapping", TransformPhaseSyntax, 10, TransformLoopOnce, replaceDotNotationWithMapping),
		inferredTransform("replaceKeyAttributes", TransformPhaseSyntax, 20, TransformLoopOnce, replaceKeyAttributes),
		inferredTransform("replaceMultiStatementSequences", TransformPhaseSyntax, 30, TransformLoopOnce, replaceMultiStatementSequences),
	})
}

// CreateModularPostProcessingPipeline builds the default post-rewriter pipeline.
func CreateModularPostProcessingPipeline() *ModularPipeline {
	return CreateModularPostProcessingPipelineWithOptions(Options{})
}

// CreateModularPostProcessingPipelineWithOptions builds the post-rewriter pipeline
// using explicit phases. The phase order is part of the transform contract:
// string -> low-precedence operator -> selector -> functional -> syntax.
func CreateModularPostProcessingPipelineWithOptions(opts Options) *ModularPipeline {
	stages := []PipelineStage{
		// Note: Regex literal processing is done in PrepareForParsing before the rewriter,
		// so we don't include it here.
		createStringProcessingStage(opts.TraceTransforms),
		createOperatorProcessingStage(opts.TraceTransforms),
		createSelectorProcessingStage(opts.TraceTransforms),
		createFunctionalProcessingStage(opts.TraceTransforms),
		createSyntaxProcessingStage(opts.TraceTransforms),
	}

	return NewModularPipeline(stages)
}

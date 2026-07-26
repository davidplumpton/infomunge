package preprocessor

// transformHandler defines the signature for transformation functions.
type transformHandler func(string) string

type mappedTransformHandler func(string) (string, []int)

// PipelineStage represents a group of related transformations.
type PipelineStage interface {
	Name() string
	Execute(input string, mapping []int) (string, []int, error)
}

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

func exactErrorTransform(name string, phase TransformPhase, order int, loop TransformLoopMode, fn mappedErrorAwareHandler) TransformContract {
	return TransformContract{
		Name:          name,
		Phase:         phase,
		Order:         order,
		Associativity: TransformAssociativityNone,
		Mapping:       TransformMappingExact,
		Loop:          loop,
		Handler:       fn,
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
		binaryOpKey:   key,
		Handler: func(input string) (string, []int, error) {
			return replaceConfiguredBinaryOperatorWithMapping(input, key)
		},
	}
}

func createCommentProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Comment Processing", TransformPhaseComment, trace, []TransformContract{
		exactTransform("stripLineComments", TransformPhaseComment, 10, TransformLoopOnce, stripLineCommentsWithMapping),
	})
}

func stripLineCommentsWithMapping(input string) (string, []int) {
	result := StripLineComments(input)
	return result, identityMapping(len(result))
}

func createRegexLiteralStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Regex Literal Processing", TransformPhaseRegex, trace, []TransformContract{
		exactTransform("replaceRegexLiteralsWithMapping", TransformPhaseRegex, 10, TransformLoopOnce, replaceRegexLiteralsWithMapping),
	})
}

func createWrapperProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Wrapper Processing", TransformPhaseWrapper, trace, []TransformContract{
		inferredTransform("wrapImplicitObjectLiteralBodies", TransformPhaseWrapper, 10, TransformLoopOnce, wrapImplicitObjectLiteralBodies),
		exactTransform("wrapTopLevelObjectLiteral", TransformPhaseWrapper, 20, TransformLoopOnce, wrapTopLevelObjectLiteralWithMapping),
	})
}

func wrapTopLevelObjectLiteralWithMapping(input string) (string, []int) {
	result, wrapped := wrapTopLevelObjectLiteral(input)
	if !wrapped {
		return result, identityMapping(len(result))
	}

	wrapperPositions := make([]int, len(result))
	for i := range wrapperPositions {
		switch {
		case i == 0:
			wrapperPositions[i] = 0
		case i == len(result)-1:
			if len(input) == 0 {
				wrapperPositions[i] = 0
			} else {
				wrapperPositions[i] = len(input) - 1
			}
		default:
			wrapperPositions[i] = i - 1
		}
	}
	return result, wrapperPositions
}

func createCoreRewriterStage(opts Options) PipelineStage {
	return newContractStage("Core Rewriter", TransformPhaseRewrite, opts.TraceTransforms, []TransformContract{
		exactErrorTransform("rewriteCoreSyntax", TransformPhaseRewrite, 10, TransformLoopOnce, func(input string) (string, []int, error) {
			return newRewriter(input, opts).RewriteCoreWithDepth(0)
		}),
	})
}

func createStringProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("String Processing", TransformPhaseString, trace, []TransformContract{
		inferredTransform("replaceStringInterpolation", TransformPhaseString, 10, TransformLoopOnce, replaceStringInterpolation),
		inferredTransform("replaceArrayRangeIndexing", TransformPhaseString, 20, TransformLoopOnce, replaceArrayRangeIndexing),
	})
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
		exactTransform("replaceAsOperator", TransformPhaseFunctional, 190, TransformLoopFixpoint, replaceAsOperatorWithMapping),
		exactTransform("replaceIsOperator", TransformPhaseFunctional, 200, TransformLoopFixpoint, replaceIsOperatorWithMapping),
		configuredBinaryOperatorTransform("replaceFindOperator", TransformPhaseFunctional, 210, binaryOpFind, TransformPrecedenceCollection, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceContainsOperator", TransformPhaseFunctional, 220, binaryOpContains, TransformPrecedenceKeywordComparison, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceSplitByOperator", TransformPhaseFunctional, 230, binaryOpSplitBy, TransformPrecedenceCollection, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceJoinByOperator", TransformPhaseFunctional, 240, binaryOpJoinBy, TransformPrecedenceCollection, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceConcatenateOperator", TransformPhaseFunctional, 250, binaryOpConcatenate, TransformPrecedenceAdditive, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceRemoveOperator", TransformPhaseFunctional, 260, binaryOpRemove, TransformPrecedenceAdditive, TransformAssociativityLeft),
		inferredTransform("replaceExponentOperator", TransformPhaseFunctional, 270, TransformLoopFixpoint, replaceExponentOperator),
		configuredBinaryOperatorTransform("replaceMatchOperator", TransformPhaseFunctional, 280, binaryOpMatch, TransformPrecedenceKeywordComparison, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceMatchesOperator", TransformPhaseFunctional, 290, binaryOpMatches, TransformPrecedenceKeywordComparison, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceScanOperator", TransformPhaseFunctional, 300, binaryOpScan, TransformPrecedenceKeywordComparison, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceModOperator", TransformPhaseFunctional, 310, binaryOpMod, TransformPrecedenceModulo, TransformAssociativityLeft),
		configuredBinaryOperatorTransform("replaceRepeatOperator", TransformPhaseFunctional, 320, binaryOpRepeat, TransformPrecedenceMultiplicative, TransformAssociativityLeft),
		inferredTransform("replaceSubstringOperator", TransformPhaseFunctional, 330, TransformLoopFixpoint, replaceSubstringOperator),
		inferredTransform("replaceContainsMethodCall", TransformPhaseFunctional, 340, TransformLoopFixpoint, replaceContainsMethodCall),
		inferredTransform("replaceFindMethodCall", TransformPhaseFunctional, 350, TransformLoopFixpoint, replaceFindMethodCall),
		inferredTransform("replaceMatchMethodCall", TransformPhaseFunctional, 360, TransformLoopFixpoint, replaceMatchMethodCall),
		inferredTransform("replaceMatchesMethodCall", TransformPhaseFunctional, 370, TransformLoopFixpoint, replaceMatchesMethodCall),
		inferredTransform("replaceScanMethodCall", TransformPhaseFunctional, 380, TransformLoopFixpoint, replaceScanMethodCall),
		inferredTransform("replaceSplitByMethodCall", TransformPhaseFunctional, 390, TransformLoopFixpoint, replaceSplitByMethodCall),
		inferredTransform("replacePipeToFunctionOperator", TransformPhaseFunctional, 400, TransformLoopFixpoint, replacePipeToFunctionOperator),
		inferredTransform("replaceReplaceOperator", TransformPhaseFunctional, 410, TransformLoopFixpoint, replaceReplaceOperator),
		inferredTransform("replaceAssignmentExpressions", TransformPhaseFunctional, 420, TransformLoopFixpoint, replaceAssignmentExpressions),
	}
}

func createSelectorProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Selector Processing", TransformPhaseSelector, trace, []TransformContract{
		inferredTransform("replaceFilterSelectors", TransformPhaseSelector, 10, TransformLoopOnce, replaceFilterSelectors),
		inferredTransform("replaceMetadataSelectors", TransformPhaseSelector, 20, TransformLoopOnce, replaceMetadataSelectors),
		exactTransform("replaceRecursiveDescentWithMapping", TransformPhaseSelector, 30, TransformLoopOnce, replaceRecursiveDescentWithMapping),
	})
}

func createSyntaxProcessingStage(trace TransformTraceFunc) PipelineStage {
	return newContractStage("Syntax Processing", TransformPhaseSyntax, trace, []TransformContract{
		exactTransform("replaceDotNotationWithMapping", TransformPhaseSyntax, 10, TransformLoopOnce, replaceDotNotationWithMapping),
		inferredTransform("replaceKeyAttributes", TransformPhaseSyntax, 20, TransformLoopOnce, replaceKeyAttributes),
		inferredTransform("replaceMultiStatementSequences", TransformPhaseSyntax, 30, TransformLoopOnce, replaceMultiStatementSequences),
	})
}

// CreateFullPreprocessingPipelineWithOptions builds the complete preprocessing
// pipeline using explicit phases. The phase order is part of the transform
// contract: comment -> regex -> wrapper -> rewrite -> string -> low-precedence
// operator -> selector -> functional -> syntax.
func CreateFullPreprocessingPipelineWithOptions(opts Options) *ModularPipeline {
	stages := []PipelineStage{
		createCommentProcessingStage(opts.TraceTransforms),
		createRegexLiteralStage(opts.TraceTransforms),
		createWrapperProcessingStage(opts.TraceTransforms),
		createCoreRewriterStage(opts),
	}
	stages = append(stages, postProcessingStages(opts)...)

	return NewModularPipeline(stages)
}

// CreateModularPostProcessingPipelineWithOptions builds the post-rewriter
// subset of the preprocessing pipeline using explicit phases. The phase order
// is part of the transform contract: string -> low-precedence operator ->
// selector -> functional -> syntax.
func CreateModularPostProcessingPipelineWithOptions(opts Options) *ModularPipeline {
	return NewModularPipeline(postProcessingStages(opts))
}

func postProcessingStages(opts Options) []PipelineStage {
	return []PipelineStage{
		createStringProcessingStage(opts.TraceTransforms),
		createOperatorProcessingStage(opts.TraceTransforms),
		createSelectorProcessingStage(opts.TraceTransforms),
		createFunctionalProcessingStage(opts.TraceTransforms),
		createSyntaxProcessingStage(opts.TraceTransforms),
	}
}

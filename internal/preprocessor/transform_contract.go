package preprocessor

import (
	"sort"

	unifiederrors "infomunge/internal/errors"
)

// Transform lifecycle:
//  1. Comment processing preserves byte positions while removing line comments.
//  2. Regex literal processing runs before syntax that could misread slashes.
//  3. Wrapper processing normalizes implicit and top-level object literals.
//  4. The recursive byte rewriter handles syntax whose structure depends on
//     braces, strings, and branch bodies.
//  5. The modular post-processing pipeline runs ordered transform contracts for
//     strings, operators, selectors, functional syntax, and final syntax cleanup.
//
// Ordering rule: phases run in the order declared by
// CreateFullPreprocessingPipelineWithOptions. The post-rewriter subset is
// declared by CreateModularPostProcessingPipelineWithOptions. Within a phase,
// lower Order values run first. A contract must state whether it runs once or to
// a fixpoint and whether its source mapping is exact or inferred.
type TransformPhase string

const (
	TransformPhaseComment    TransformPhase = "comment"
	TransformPhaseRegex      TransformPhase = "regex"
	TransformPhaseWrapper    TransformPhase = "wrapper"
	TransformPhaseRewrite    TransformPhase = "rewrite"
	TransformPhaseString     TransformPhase = "string"
	TransformPhaseOperator   TransformPhase = "operator"
	TransformPhaseSelector   TransformPhase = "selector"
	TransformPhaseFunctional TransformPhase = "functional"
	TransformPhaseSyntax     TransformPhase = "syntax"
)

type TransformLoopMode string

const (
	TransformLoopOnce     TransformLoopMode = "once"
	TransformLoopFixpoint TransformLoopMode = "fixpoint"
)

type TransformMappingMode string

const (
	TransformMappingExact    TransformMappingMode = "exact"
	TransformMappingInferred TransformMappingMode = "inferred"
)

type TransformAssociativity string

const (
	TransformAssociativityNone  TransformAssociativity = "none"
	TransformAssociativityLeft  TransformAssociativity = "left"
	TransformAssociativityRight TransformAssociativity = "right"
)

const (
	TransformPrecedenceNone = iota
	TransformPrecedenceDefault
	TransformPrecedenceNullChain
	TransformPrecedenceRange
	TransformPrecedenceCollection
	TransformPrecedenceType
	TransformPrecedenceComparison
	// TransformPrecedenceModulo matches DataWeave's infix mod operator:
	// it binds less tightly than additive and multiplicative arithmetic.
	TransformPrecedenceModulo
	TransformPrecedenceAdditive
	TransformPrecedenceMultiplicative
	TransformPrecedencePower
)

type TransformTraceFunc func(TransformTraceEntry)

type TransformTraceEntry struct {
	Stage         string
	Phase         TransformPhase
	Transform     string
	Order         int
	Mapping       TransformMappingMode
	Loop          TransformLoopMode
	Before        string
	After         string
	Changed       bool
	ExactMapping  bool
	Precedence    int
	Associativity TransformAssociativity
}

type TransformContract struct {
	Name          string
	Phase         TransformPhase
	Order         int
	Precedence    int
	Associativity TransformAssociativity
	Mapping       TransformMappingMode
	Loop          TransformLoopMode
	Handler       mappedErrorAwareHandler
	binaryOpKey   string
}

type contractStage struct {
	name       string
	phase      TransformPhase
	trace      TransformTraceFunc
	transforms []TransformContract
}

func newContractStage(name string, phase TransformPhase, trace TransformTraceFunc, transforms []TransformContract) PipelineStage {
	ordered := append([]TransformContract(nil), transforms...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Order < ordered[j].Order
	})
	configureBinaryOperatorAssociativity(ordered)

	return &contractStage{
		name:       name,
		phase:      phase,
		trace:      trace,
		transforms: ordered,
	}
}

func configureBinaryOperatorAssociativity(contracts []TransformContract) {
	for i := range contracts {
		contract := &contracts[i]
		if contract.binaryOpKey == "" ||
			contract.Precedence == TransformPrecedenceNone ||
			contract.Associativity != TransformAssociativityLeft {
			continue
		}

		peerOps := make([]string, 0)
		for _, peer := range contracts {
			if peer.binaryOpKey == "" ||
				peer.Precedence != contract.Precedence ||
				peer.Associativity != contract.Associativity {
				continue
			}
			config, ok := binaryOperatorConfigs[peer.binaryOpKey]
			if ok {
				peerOps = append(peerOps, config.Operator)
			}
		}

		key := contract.binaryOpKey
		contract.Handler = func(input string) (string, []int, error) {
			return replaceConfiguredBinaryOperatorWithMappingAndPeers(input, key, peerOps)
		}
	}
}

func (cs *contractStage) Name() string {
	return cs.name
}

func (cs *contractStage) Execute(input string, mapping []int) (string, []int, error) {
	result := input
	resultMapping := mapping

	for _, transform := range cs.transforms {
		before := result
		next, local, err := executeTransformContract(transform, result)
		if err != nil {
			return result, resultMapping, err
		}
		if len(next) != len(local) {
			return next, resultMapping, unifiederrors.ParseErrorf(
				"mapping invariant violated in %s: len(result)=%d, len(mapping)=%d",
				transform.Name,
				len(next),
				len(local),
			)
		}

		result = next
		resultMapping = composeMappings(resultMapping, local)
		cs.emitTrace(transform, before, result)
	}

	return result, resultMapping, nil
}

func executeTransformContract(transform TransformContract, input string) (string, []int, error) {
	if transform.Handler == nil {
		return input, identityMapping(len(input)), unifiederrors.ParseErrorf("transform %s has no handler", transform.Name)
	}

	switch transform.Loop {
	case TransformLoopFixpoint:
		return executeFixpointTransform(transform, input)
	case TransformLoopOnce, "":
		result, mapping, err := transform.Handler(input)
		if err != nil {
			return result, mapping, err
		}
		return result, mapping, nil
	default:
		return input, identityMapping(len(input)), unifiederrors.ParseErrorf(
			"transform %s has unsupported loop mode %q",
			transform.Name,
			transform.Loop,
		)
	}
}

func executeFixpointTransform(transform TransformContract, input string) (string, []int, error) {
	result := input
	resultMapping := identityMapping(len(input))
	prevResult := ""
	iterCount := 0

	for result != prevResult {
		if iterCount > MaxTransformIterations {
			return result, resultMapping, unifiederrors.ParseErrorf("infinite loop detected in %s", transform.Name)
		}

		prevResult = result
		nextResult, local, err := transform.Handler(result)
		if err != nil {
			return result, resultMapping, err
		}
		if len(nextResult) != len(local) {
			return nextResult, resultMapping, unifiederrors.ParseErrorf(
				"mapping invariant violated in %s: len(result)=%d, len(mapping)=%d",
				transform.Name,
				len(nextResult),
				len(local),
			)
		}

		result = nextResult
		resultMapping = composeMappings(resultMapping, local)
		iterCount++
	}

	return result, resultMapping, nil
}

func (cs *contractStage) emitTrace(transform TransformContract, before, after string) {
	if cs.trace == nil {
		return
	}

	cs.trace(TransformTraceEntry{
		Stage:         cs.name,
		Phase:         transform.Phase,
		Transform:     transform.Name,
		Order:         transform.Order,
		Mapping:       transform.Mapping,
		Loop:          transform.Loop,
		Before:        before,
		After:         after,
		Changed:       before != after,
		ExactMapping:  transform.Mapping == TransformMappingExact,
		Precedence:    transform.Precedence,
		Associativity: transform.Associativity,
	})
}

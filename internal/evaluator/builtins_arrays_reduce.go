package evaluator

import (
	"go/ast"
	"go/token"
)

type reduceSetup struct {
	accParamIdx  int
	elemParamIdx int
	hasInitial   bool
	initialValue interface{}
}

func determineReduceSetup(lambda *Lambda) reduceSetup {
	setup := reduceSetup{
		accParamIdx:  0,
		elemParamIdx: 1,
	}

	if !lambda.HasDefaults() {
		return setup
	}

	if defaultVal, ok := lambda.GetDefault(lambda.ParamName(1)); ok {
		setup.hasInitial = true
		setup.initialValue = defaultVal
		setup.accParamIdx = 1
		setup.elemParamIdx = 0
		return setup
	}

	if defaultVal, ok := lambda.GetDefault(lambda.ParamName(0)); ok {
		setup.hasInitial = true
		setup.initialValue = defaultVal
		setup.accParamIdx = 0
		setup.elemParamIdx = 1
	}

	return setup
}

func handleReduceEmptyArray(array []interface{}, setup reduceSetup, pos token.Pos) (interface{}, bool, error) {
	if len(array) != 0 {
		return nil, false, nil
	}
	if setup.hasInitial {
		return setup.initialValue, true, nil
	}
	return nil, true, newPosError("reduce cannot be applied to an empty array without an initial value", pos)
}

func reduceInitialAccumulator(array []interface{}, setup reduceSetup) (interface{}, int) {
	if setup.hasInitial {
		return setup.initialValue, 0
	}
	return array[0], 1
}

func runReduce(array []interface{}, lambda *Lambda, context map[string]interface{}, depth int, setup reduceSetup) (interface{}, error) {
	accumulator, startIdx := reduceInitialAccumulator(array, setup)
	for i := startIdx; i < len(array); i++ {
		lambdaContext := copyContext(context)
		lambdaContext[lambda.ParamName(setup.accParamIdx)] = accumulator
		lambdaContext[lambda.ParamName(setup.elemParamIdx)] = array[i]
		if lambda.ParamCount() > 2 {
			lambdaContext[lambda.ParamName(2)] = i
		}

		result, err := evalASTWithDepth(lambda.BodyAST, lambdaContext, depth+1)
		if err != nil {
			return nil, err
		}
		accumulator = result
	}

	return accumulator, nil
}

// callBuiltinReduce implements the __reduce(array, lambda) function.
//
// Reduces an array to a single value by applying a lambda function cumulatively.
//
// Arguments:
//   - array: The array to reduce
//   - lambda: A function with 2-3 parameters
//
// Lambda parameters (2-3 params):
//   - accumulator: The accumulated result from previous iterations
//   - element: The current array element
//   - index (optional): The current element's index
//
// Initial value handling (DataWeave style):
//   - If the lambda has a default value on any parameter, that parameter becomes
//     the accumulator and its default becomes the initial value.
//   - Example: (item, acc = 0) -> acc + item  // acc starts at 0
//   - Example: (acc = [], item) -> acc ++ [item]  // acc starts as empty array
//   - Without defaults: first element is used as initial accumulator, iteration
//     starts from second element.
//
// Returns an error if the array is empty and no initial value is provided.
func callBuiltinReduce(e *ast.CallExpr, context map[string]interface{}, depth int) (interface{}, error) {
	array, lambda, err := evalArrayAndLambda("reduce", e, context, depth, 2, 3)
	if err != nil {
		return nil, err
	}

	setup := determineReduceSetup(lambda)
	if result, handled, err := handleReduceEmptyArray(array, setup, e.Args[0].Pos()); handled {
		return result, err
	}

	return runReduce(array, lambda, context, depth, setup)
}

package evaluator

import (
	"go/ast"
)

type reduceSetup struct {
	hasInitial   bool
	initialValue Value
}

func determineReduceSetup(lambda *Lambda) reduceSetup {
	setup := reduceSetup{}
	if defaultVal, ok := lambda.GetDefault(lambda.ParamName(1)); ok {
		setup.hasInitial = true
		setup.initialValue = defaultVal
	}

	return setup
}

func handleReduceEmptyArray(array Array, setup reduceSetup) (Value, bool, error) {
	if len(array) != 0 {
		return nil, false, nil
	}
	if setup.hasInitial {
		return setup.initialValue, true, nil
	}
	return nil, true, nil
}

func reduceInitialAccumulator(array Array, setup reduceSetup) (Value, int) {
	if setup.hasInitial {
		return setup.initialValue, 0
	}
	return array[0], 1
}

func runReduce(array Array, lambda *Lambda, scope *Scope, depth int, setup reduceSetup) (Value, error) {
	accumulator, startIdx := reduceInitialAccumulator(array, setup)
	for i := startIdx; i < len(array); i++ {
		result, err := evalLambdaWithBindingsAtDepth(lambda, scope, depth+1, func(lambdaContext Context) {
			lambdaContext[lambda.ParamName(0)] = array[i]
			lambdaContext[lambda.ParamName(1)] = accumulator
			if lambda.ParamCount() > 2 {
				lambdaContext[lambda.ParamName(2)] = i
			}
		})
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
//   - element: The current array element
//   - accumulator: The accumulated result from previous iterations
//   - index (optional InfoMunge extension): The current element's index
//
// Initial value handling (DataWeave style):
//   - A default value on the second (accumulator) parameter becomes the initial
//     value.
//   - Example: (item, acc = 0) -> acc + item  // acc starts at 0
//   - A default on the first (element) parameter does not initialize the
//     accumulator because every invocation supplies an element.
//   - Without an accumulator default: first element is used as the initial
//     accumulator, and iteration starts from the second element.
//
// Returns null if the array is empty and no initial value is provided.
func callBuiltinReduce(e *ast.CallExpr, scope *Scope, depth int) (Value, error) {
	array, lambda, err := evalArrayAndLambda("reduce", e, scope, depth, 2, 3)
	if err != nil {
		return nil, err
	}

	setup := determineReduceSetup(lambda)
	if result, handled, err := handleReduceEmptyArray(array, setup); handled {
		return result, err
	}

	return runReduce(array, lambda, scope, depth, setup)
}

package evaluator

// shouldUseLegacyObjectLambdaOrder returns true when the lambda explicitly
// declares legacy InfoMunge ordering (key, value) or (k, v). All other
// parameter names default to DataWeave ordering (value, key).
func shouldUseLegacyObjectLambdaOrder(lambda *Lambda) bool {
	if lambda == nil || lambda.ParamCount() < 2 {
		return false
	}

	param0 := lambda.ParamName(0)
	param1 := lambda.ParamName(1)

	return (param0 == "key" && param1 == "value") || (param0 == "k" && param1 == "v")
}


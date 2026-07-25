package evaluator

func numberBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		withArityMessages(
			regularBuiltinSpec("ceil", builtinCategoryNumbers, exactArity(1), callBuiltinCeil, "ceiling"),
			"ceil function requires exactly 1 argument",
			"ceil function requires exactly 1 argument",
		),
		regularBuiltinSpec("floor", builtinCategoryNumbers, exactArity(1), callBuiltinFloor, "floor"),
		regularBuiltinSpec("round", builtinCategoryNumbers, exactArity(1), callBuiltinRound, "round"),
		regularBuiltinSpec("sqrt", builtinCategoryNumbers, exactArity(1), callBuiltinSqrt, "square root"),
		regularBuiltinSpec("abs", builtinCategoryNumbers, exactArity(1), callBuiltinAbs, "absolute value"),
		regularBuiltinSpec("max", builtinCategoryNumbers, variadicArity(1), callBuiltinMax, "maximum"),
		regularBuiltinSpec("min", builtinCategoryNumbers, variadicArity(1), callBuiltinMin, "minimum"),
		regularBuiltinSpec("pow", builtinCategoryNumbers, exactArity(2), callBuiltinPow, "power"),
		regularBuiltinSpec("sum", builtinCategoryNumbers, exactArity(1), callBuiltinSum, "sum numbers"),
		regularBuiltinSpec("avg", builtinCategoryNumbers, exactArity(1), callBuiltinAvg, "average numbers"),
		regularBuiltinSpec("mod", builtinCategoryNumbers, exactArity(2), callBuiltinMod, "modulo"),
		regularBuiltinSpec("toRadix", builtinCategoryNumbers, exactArity(2), callBuiltinToRadix, "format number with radix"),
		regularBuiltinSpec("fromRadix", builtinCategoryNumbers, exactArity(2), callBuiltinFromRadix, "parse number with radix"),
		regularBuiltinSpec("toBinary", builtinCategoryNumbers, exactArity(1), callBuiltinToBinary, "format binary"),
		regularBuiltinSpec("fromBinary", builtinCategoryNumbers, exactArity(1), callBuiltinFromBinary, "parse binary"),
		regularBuiltinSpec("random", builtinCategoryNumbers, exactArity(0), callBuiltinRandom, "random number"),
		regularBuiltinSpec("randomInt", builtinCategoryNumbers, exactArity(1), callBuiltinRandomInt, "random integer"),
		regularBuiltinSpec("to", builtinCategoryNumbers, exactArity(2), callBuiltinTo, "inclusive range"),
		regularBuiltinSpec("isDecimal", builtinCategoryNumbers, exactArity(1), callBuiltinIsDecimal, "decimal check"),
		regularBuiltinSpec("isInteger", builtinCategoryNumbers, exactArity(1), callBuiltinIsInteger, "integer check"),
		regularBuiltinSpec("isEven", builtinCategoryNumbers, exactArity(1), callBuiltinIsEven, "even integer check"),
		regularBuiltinSpec("isOdd", builtinCategoryNumbers, exactArity(1), callBuiltinIsOdd, "odd integer check"),
	}
}

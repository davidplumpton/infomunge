package evaluator

func coreBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		specialBuiltinSpec("__default", builtinCategoryCore, exactArity(2), callBuiltinDefault, "default operator"),
		specialBuiltinSpec("__lambda", builtinCategoryCore, exactArity(2), callBuiltinLambdaAST, "lambda expression"),
		specialBuiltinSpec("__modcall", builtinCategoryCore, variadicArity(2), callBuiltinModCall, "module function call"),
		specialBuiltinSpec("__native", builtinCategoryCore, variadicArity(1), callBuiltinNative, "explicit native builtin call"),
		withArityMessages(
			specialBuiltinSpec("__coerce", builtinCategoryCore, rangeArity(2, 3), callBuiltinCoerce, "type coercion"),
			"as operator requires 2 or 3 arguments: value, type[, config]",
			"as operator requires 2 or 3 arguments: value, type[, config]",
		),
		specialBuiltinSpec("__case", builtinCategoryCore, exactArity(2), callBuiltinCase, "pattern matching case"),
		specialBuiltinSpec("lazy_eval", builtinCategoryCore, exactArity(1), callBuiltinLazyEval, "lazy expression wrapper"),
		specialBuiltinSpec("onNull", builtinCategoryCore, exactArity(2), callBuiltinOnNull, "null fallback"),
		specialBuiltinSpec("then", builtinCategoryCore, exactArity(2), callBuiltinThen, "chained expression"),
		regularBuiltinSpec("force_eval", builtinCategoryCore, exactArity(1), callBuiltinForceEval, "force lazy evaluation"),
		regularBuiltinSpec("typeOf", builtinCategoryCore, exactArity(1), callBuiltinTypeOf, "value type name"),
		regularBuiltinSpec("__isType", builtinCategoryCore, exactArity(2), callBuiltinIsType, "type check operator"),
	}
}

func controlFlowBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		specialBuiltinSpec("__ifelse", builtinCategoryCore, exactArity(3), callBuiltinIfElse, "conditional expression"),
		specialBuiltinSpec("__while", builtinCategoryCore, exactArity(2), callBuiltinWhile, "while loop"),
		specialBuiltinSpec("__assign", builtinCategoryCore, exactArity(2), callBuiltinAssign, "do-block assignment"),
		specialBuiltinSpec("__break", builtinCategoryCore, exactArity(0), callBuiltinBreak, "break loop"),
		specialBuiltinSpec("__continue", builtinCategoryCore, exactArity(0), callBuiltinContinue, "continue loop"),
		specialBuiltinSpec("__seq", builtinCategoryCore, variadicArity(1), callBuiltinSeq, "ordered expression sequence"),
		specialBuiltinSpec("__do", builtinCategoryCore, exactArity(2), callBuiltinDo, "do block"),
	}
}

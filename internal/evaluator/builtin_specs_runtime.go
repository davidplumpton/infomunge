package evaluator

func runtimeBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		specialBuiltinSpec("try", builtinCategoryRuntime, exactArity(1), callBuiltinTry, "try expression"),
		specialBuiltinSpec("orElse", builtinCategoryRuntime, exactArity(2), callBuiltinOrElse, "try fallback"),
		specialBuiltinSpec("orElseTry", builtinCategoryRuntime, exactArity(2), callBuiltinOrElseTry, "try fallback expression"),
		regularBuiltinSpec("uuid", builtinCategoryRuntime, exactArity(0), callBuiltinUUID, "random UUID"),
		regularBuiltinSpec("log", builtinCategoryRuntime, rangeArity(1, 2), callBuiltinLog, "log value"),
		regularBuiltinSpec("logDebug", builtinCategoryRuntime, rangeArity(1, 2), callBuiltinLogDebug, "debug log value"),
		regularBuiltinSpec("logInfo", builtinCategoryRuntime, rangeArity(1, 2), callBuiltinLogInfo, "info log value"),
		regularBuiltinSpec("logWarn", builtinCategoryRuntime, rangeArity(1, 2), callBuiltinLogWarn, "warning log value"),
		regularBuiltinSpec("logError", builtinCategoryRuntime, rangeArity(1, 2), callBuiltinLogError, "error log value"),
		regularBuiltinSpec("logWith", builtinCategoryRuntime, exactArity(2), callBuiltinLogWith, "structured log value"),
		regularBuiltinSpec("evaluateCompatibilityFlag", builtinCategoryRuntime, exactArity(1), callBuiltinEvaluateCompatibilityFlag, "compatibility flag"),
		regularBuiltinSpec("fail", builtinCategoryRuntime, rangeArity(0, 1), callBuiltinFail, "raise failure"),
		regularBuiltinSpec("assert", builtinCategoryRuntime, rangeArity(1, 2), callBuiltinAssert, "assert condition"),
		regularBuiltinSpec("assertThat", builtinCategoryRuntime, exactArity(2), callBuiltinAssertThat, "assert matcher"),
		regularBuiltinSpec("envVar", builtinCategoryRuntime, exactArity(1), callBuiltinEnvVar, "environment variable"),
		regularBuiltinSpec("envVars", builtinCategoryRuntime, exactArity(0), callBuiltinEnvVars, "environment variables"),
	}
}

package evaluator

func runtimeBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		specialBuiltinSpec("try", builtinCategoryRuntime, exactArity(1), callBuiltinTry, "try expression"),
		specialBuiltinSpec("orElse", builtinCategoryRuntime, exactArity(2), callBuiltinOrElse, "try fallback"),
		specialBuiltinSpec("orElseTry", builtinCategoryRuntime, exactArity(2), callBuiltinOrElseTry, "try fallback expression"),
		regularBuiltinSpec("uuid", builtinCategoryRuntime, exactArity(0), callBuiltinUUID, "random UUID"),
		regularBuiltinSpec("log", builtinCategoryRuntime, exactArity(1), callBuiltinLog, "log value"),
		regularBuiltinSpec("logDebug", builtinCategoryRuntime, exactArity(1), callBuiltinLogDebug, "debug log value"),
		regularBuiltinSpec("logInfo", builtinCategoryRuntime, exactArity(1), callBuiltinLogInfo, "info log value"),
		regularBuiltinSpec("logWarn", builtinCategoryRuntime, exactArity(1), callBuiltinLogWarn, "warning log value"),
		regularBuiltinSpec("logError", builtinCategoryRuntime, exactArity(1), callBuiltinLogError, "error log value"),
		regularBuiltinSpec("logWith", builtinCategoryRuntime, exactArity(2), callBuiltinLogWith, "structured log value"),
		regularBuiltinSpec("evaluateCompatibilityFlag", builtinCategoryRuntime, exactArity(1), callBuiltinEvaluateCompatibilityFlag, "compatibility flag"),
		regularBuiltinSpec("fail", builtinCategoryRuntime, rangeArity(0, 1), callBuiltinFail, "raise failure"),
		withArityMessages(
			regularBuiltinSpec("assert", builtinCategoryRuntime, rangeArity(2, 3), callBuiltinAssert, "assert condition"),
			"assert expects 2 or 3 arguments: value, matcher, [message]",
			"assert expects 2 or 3 arguments: value, matcher, [message]",
		),
		regularBuiltinSpec("assertThat", builtinCategoryRuntime, exactArity(2), callBuiltinAssertThat, "assert matcher"),
		regularBuiltinSpec("envVar", builtinCategoryRuntime, exactArity(1), callBuiltinEnvVar, "environment variable"),
		regularBuiltinSpec("envVars", builtinCategoryRuntime, exactArity(0), callBuiltinEnvVars, "environment variables"),
	}
}

package evaluator

func ioBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		withArityMessages(
			specialBuiltinSpec("read", builtinCategoryIO, rangeArity(2, 3), callBuiltinRead, "read content"),
			"read function requires at least 2 arguments: content and mimeType",
			"read function accepts at most 3 arguments: content, mimeType, and optional options object",
		),
		specialBuiltinSpec("readUrl", builtinCategoryIO, exactArity(2), callBuiltinReadUrl, "read URL"),
		withArityMessages(
			specialBuiltinSpec("write", builtinCategoryIO, rangeArity(2, 3), callBuiltinWrite, "write content"),
			"write requires exactly 2 arguments: value and mimeType",
			"write accepts at most 3 arguments: value, mimeType, and optional options object",
		),
	}
}

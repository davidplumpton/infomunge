package evaluator

func ioBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		specialBuiltinSpec("read", builtinCategoryIO, rangeArity(2, 3), callBuiltinRead, "read content"),
		specialBuiltinSpec("readUrl", builtinCategoryIO, exactArity(2), callBuiltinReadUrl, "read URL"),
		specialBuiltinSpec("write", builtinCategoryIO, rangeArity(2, 3), callBuiltinWrite, "write content"),
	}
}

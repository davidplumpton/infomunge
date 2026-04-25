package evaluator

func urlBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		regularBuiltinSpec("parseURI", builtinCategoryURLs, exactArity(1), callBuiltinParseURI, "parse URI"),
		regularBuiltinSpec("compose", builtinCategoryURLs, exactArity(1), callBuiltinCompose, "compose URI"),
		regularBuiltinSpec("encodeURI", builtinCategoryURLs, exactArity(1), callBuiltinEncodeURI, "encode URI"),
		regularBuiltinSpec("decodeURI", builtinCategoryURLs, exactArity(1), callBuiltinDecodeURI, "decode URI"),
		regularBuiltinSpec("encodeURIComponent", builtinCategoryURLs, exactArity(1), callBuiltinEncodeURIComponent, "encode URI component"),
		regularBuiltinSpec("decodeURIComponent", builtinCategoryURLs, exactArity(1), callBuiltinDecodeURIComponent, "decode URI component"),
	}
}

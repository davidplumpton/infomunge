package evaluator

func assertionBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		specialBuiltinSpec("must", builtinCategoryAssertions, exactArity(2), callMust, "assert value with matcher"),
		specialBuiltinSpec("eachItem", builtinCategoryAssertions, exactArity(1), callEachItemMatcher, "each item matcher"),
		specialBuiltinSpec("haveItem", builtinCategoryAssertions, exactArity(1), callHaveItemMatcher, "have item matcher"),
		specialBuiltinSpec("anyOf", builtinCategoryAssertions, exactArity(1), callAnyOfMatcher, "any matcher"),
		specialBuiltinSpec("notBe", builtinCategoryAssertions, exactArity(1), callNotBeMatcher, "negated matcher"),
		regularBuiltinSpec("beArray", builtinCategoryAssertions, exactArity(0), callBeArray, "array matcher"),
		regularBuiltinSpec("beObject", builtinCategoryAssertions, exactArity(0), callBeObject, "object matcher"),
		regularBuiltinSpec("beString", builtinCategoryAssertions, exactArity(0), callBeString, "string matcher"),
		regularBuiltinSpec("beNumber", builtinCategoryAssertions, exactArity(0), callBeNumber, "number matcher"),
		regularBuiltinSpec("beBoolean", builtinCategoryAssertions, exactArity(0), callBeBoolean, "boolean matcher"),
		regularBuiltinSpec("beNull", builtinCategoryAssertions, exactArity(0), callBeNull, "null matcher"),
		regularBuiltinSpec("beEmpty", builtinCategoryAssertions, exactArity(0), callBeEmpty, "empty matcher"),
		regularBuiltinSpec("beBlank", builtinCategoryAssertions, exactArity(0), callBeBlank, "blank matcher"),
		regularBuiltinSpec("equalTo", builtinCategoryAssertions, exactArity(1), callEqualTo, "equality matcher"),
		regularBuiltinSpec("beGreaterThan", builtinCategoryAssertions, rangeArity(1, 2), callBeGreaterThan, "greater-than matcher"),
		regularBuiltinSpec("beLowerThan", builtinCategoryAssertions, rangeArity(1, 2), callBeLowerThan, "lower-than matcher"),
		regularBuiltinSpec("beOneOf", builtinCategoryAssertions, exactArity(1), callBeOneOf, "one-of matcher"),
		regularBuiltinSpec("containStr", builtinCategoryAssertions, exactArity(1), callContainStr, "string containment matcher"),
		regularBuiltinSpec("containVal", builtinCategoryAssertions, exactArity(1), callContainVal, "value containment matcher"),
		regularBuiltinSpec("startWith", builtinCategoryAssertions, exactArity(1), callStartWith, "prefix matcher"),
		regularBuiltinSpec("endWith", builtinCategoryAssertions, exactArity(1), callEndWith, "suffix matcher"),
		regularBuiltinSpec("haveSize", builtinCategoryAssertions, exactArity(1), callHaveSize, "size matcher"),
		regularBuiltinSpec("haveKey", builtinCategoryAssertions, exactArity(1), callHaveKey, "object key matcher"),
		regularBuiltinSpec("haveValue", builtinCategoryAssertions, exactArity(1), callHaveValue, "object value matcher"),
		regularBuiltinSpec("notBeNull", builtinCategoryAssertions, exactArity(0), callNotBeNull, "not-null matcher"),
	}
}

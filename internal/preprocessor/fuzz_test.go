package preprocessor

import "testing"

func FuzzPrepareForParsing(f *testing.F) {
	seeds := []string{
		"",
		"payload map $ + 1",
		"%im 0.1\noutput application/json\n---\n{a: 1}",
		"if (a) b else c",
		"a and b or c",
		"obj.field",
		`"Hello $(name)"`,
		`/[a-z]+/`,
		"payload filter $.age > 21",
		"do { var x = 1 --- x }",
		"a default b",
		"a ++ b",
		"a[0]",
		"payload map ((item) -> item.name)",
		"payload pluck ((value, key) -> { (key): value })",
		`"escaped \\\" quote"`,
		"// line comment\n---\n{a: 1}",
	}

	for _, seed := range seeds {
		f.Add(seed, false)
		f.Add(seed, true)
	}

	f.Fuzz(func(t *testing.T, input string, allowMultiline bool) {
		opts := Options{AllowMultilineIfElse: allowMultiline}
		_, _, _ = PrepareForParsing(input, opts)
	})
}

func FuzzExtractHeaderAndBody(f *testing.F) {
	seeds := []string{
		"",
		"---",
		"output application/json\n---\n{a: 1}",
		"%im 0.1 var x = 10 output application/json --- x + 1",
		"fun addOne(x) = do {\n  var y = x + 1\n  ---\n  y\n}\noutput application/json\n---\naddOne(2)",
		`"--- inside string"`,
		"// --- line comment\npayload",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _, offset := ExtractHeaderAndBody(input)
		if offset < 0 || offset > len(input) {
			t.Fatalf("offset out of range: %d (len=%d)", offset, len(input))
		}
	})
}

package mutation

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	numberLiteralRe   = regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	stringLiteralRe   = regexp.MustCompile(`"([^"\\]|\\.)*"`)
	functionCallName  = regexp.MustCompile(`\b[A-Za-z_][A-Za-z0-9_]*\s*\(`)
	functionSwapNames = map[string][]string{
		"sizeOf":     {"typeOf", "upper", "lower", "flatten", "keys", "values", "trim", "isEmpty"},
		"typeOf":     {"sizeOf", "upper", "lower", "trim", "isEmpty"},
		"upper":      {"lower", "trim"},
		"lower":      {"upper", "trim"},
		"flatten":    {"keys", "values"},
		"keys":       {"values"},
		"values":     {"keys"},
		"trim":       {"upper", "lower"},
		"isEmpty":    {"sizeOf", "typeOf"},
		"contains":   {"startsWith", "endsWith"},
		"startsWith": {"contains", "endsWith"},
		"endsWith":   {"contains", "startsWith"},
	}
)

type mutationOperator func(string, *rand.Rand) string

type replacement struct {
	old string
	new string
}

var singleExprMutations = []mutationOperator{
	OperatorSwap,
	LiteralPerturb,
	SubtreeSwapWithin,
	ArgManipulate,
	NestWrap,
	ParenMutate,
	FunctionSwap,
}

// Mutate applies one random mutation to an expression.
func Mutate(expr string, rng *rand.Rand) string {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	}
	if strings.TrimSpace(expr) == "" {
		return expr
	}

	idxs := rng.Perm(len(singleExprMutations))
	for _, idx := range idxs {
		mutated := singleExprMutations[idx](expr, rng)
		if mutated != expr && strings.TrimSpace(mutated) != "" {
			return mutated
		}
	}

	return expr
}

// MutateN applies n random mutations in sequence.
func MutateN(expr string, n int, rng *rand.Rand) string {
	if n <= 0 {
		return expr
	}
	mutated := expr
	for i := 0; i < n; i++ {
		mutated = Mutate(mutated, rng)
	}
	return mutated
}

// OperatorSwap replaces one operator token with a compatible alternative.
func OperatorSwap(expr string, rng *rand.Rand) string {
	choices := []replacement{
		{old: " + ", new: " - "},
		{old: " - ", new: " + "},
		{old: " * ", new: " / "},
		{old: " / ", new: " * "},
		{old: " == ", new: " != "},
		{old: " != ", new: " == "},
		{old: " <= ", new: " >= "},
		{old: " >= ", new: " <= "},
		{old: " < ", new: " > "},
		{old: " > ", new: " < "},
		{old: " && ", new: " || "},
		{old: " || ", new: " && "},
		{old: " map ", new: " filter "},
		{old: " filter ", new: " map "},
		{old: " ++ ", new: " + "},
	}

	return applyReplacement(expr, choices, rng)
}

// LiteralPerturb changes a literal value while preserving source-level syntax.
func LiteralPerturb(expr string, rng *rand.Rand) string {
	stringMatches := stringLiteralRe.FindAllStringIndex(expr, -1)
	if len(stringMatches) > 0 {
		chosen := stringMatches[rng.Intn(len(stringMatches))]
		return expr[:chosen[0]] + `""` + expr[chosen[1]:]
	}

	boolOrNull := []struct {
		old string
		new string
	}{
		{old: "true", new: "false"},
		{old: "false", new: "true"},
		{old: "null", new: "0"},
		{old: "nil", new: "0"},
	}

	for _, c := range boolOrNull {
		if idx := wholeWordIndex(expr, c.old); idx >= 0 {
			return expr[:idx] + c.new + expr[idx+len(c.old):]
		}
	}

	num := numberLiteralRe.FindStringIndex(expr)
	if num != nil {
		return expr[:num[0]] + "0" + expr[num[1]:]
	}

	return expr
}

// SubtreeSwap replaces one bracketed subexpression in exprA with one from exprB.
func SubtreeSwap(exprA, exprB string, rng *rand.Rand) string {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	}
	spansA := findBracketSpans(exprA)
	spansB := findBracketSpans(exprB)
	if len(spansA) == 0 || len(spansB) == 0 {
		return exprA
	}

	a := spansA[rng.Intn(len(spansA))]
	b := spansB[rng.Intn(len(spansB))]
	return exprA[:a.start] + exprB[b.start:b.end] + exprA[a.end:]
}

// ArgManipulate mutates argument lists by remove/add/reorder.
func ArgManipulate(expr string, rng *rand.Rand) string {
	callStart, callEnd, argsStart, argsEnd, ok := findFunctionCall(expr)
	if !ok {
		return expr
	}

	args := splitArgs(expr[argsStart:argsEnd])
	if len(args) == 0 {
		return expr
	}

	action := rng.Intn(3)
	switch action {
	case 0: // remove
		if len(args) == 1 {
			return expr
		}
		removeIdx := rng.Intn(len(args))
		args = append(args[:removeIdx], args[removeIdx+1:]...)
	case 1: // add
		addIdx := rng.Intn(len(args))
		args = append(args, args[addIdx])
	case 2: // reorder
		if len(args) == 1 {
			return expr
		}
		i := rng.Intn(len(args))
		j := rng.Intn(len(args))
		for j == i {
			j = rng.Intn(len(args))
		}
		args[i], args[j] = args[j], args[i]
	}

	newCall := expr[callStart:argsStart] + strings.Join(args, ", ") + expr[argsEnd:callEnd]
	return expr[:callStart] + newCall + expr[callEnd:]
}

// NestWrap wraps an expression with an extra syntactic layer.
func NestWrap(expr string, rng *rand.Rand) string {
	if rng.Intn(2) == 0 {
		return "(" + expr + ")"
	}
	return "[" + expr + "][0]"
}

// ParenMutate inserts or removes one pair of parentheses.
func ParenMutate(expr string, rng *rand.Rand) string {
	open := strings.Index(expr, "(")
	close := strings.LastIndex(expr, ")")
	if open >= 0 && close > open {
		return expr[:open] + expr[open+1:close] + expr[close+1:]
	}
	return "(" + expr + ")"
}

// FunctionSwap replaces one builtin function name with another of compatible arity.
func FunctionSwap(expr string, rng *rand.Rand) string {
	matches := functionCallName.FindAllStringIndex(expr, -1)
	if len(matches) == 0 {
		return expr
	}

	type candidate struct {
		start int
		end   int
		name  string
	}
	candidates := make([]candidate, 0, len(matches))
	for _, m := range matches {
		callPrefix := strings.TrimSpace(expr[m[0]:m[1]])
		name := strings.TrimSuffix(callPrefix, "(")
		name = strings.TrimSpace(name)
		alts, ok := functionSwapNames[name]
		if !ok || len(alts) == 0 {
			continue
		}
		candidates = append(candidates, candidate{start: m[0], end: m[1], name: name})
	}
	if len(candidates) == 0 {
		return expr
	}

	chosen := candidates[rng.Intn(len(candidates))]
	alts := functionSwapNames[chosen.name]
	repl := alts[rng.Intn(len(alts))]
	return expr[:chosen.start] + repl + expr[chosen.end-1:]
}

func applyReplacement(expr string, choices []replacement, rng *rand.Rand) string {
	type match struct {
		idx int
		old string
		new string
	}
	var matches []match
	for _, c := range choices {
		if idx := strings.Index(expr, c.old); idx >= 0 {
			matches = append(matches, match{idx: idx, old: c.old, new: c.new})
		}
	}
	if len(matches) == 0 {
		return expr
	}
	chosen := matches[rng.Intn(len(matches))]
	return expr[:chosen.idx] + chosen.new + expr[chosen.idx+len(chosen.old):]
}

type span struct {
	start int
	end   int
}

func findBracketSpans(expr string) []span {
	var spans []span
	type stackItem struct {
		ch  rune
		pos int
	}
	stack := make([]stackItem, 0)
	inString := false
	escaped := false

	for i, r := range expr {
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == '"' {
				inString = false
			}
			continue
		}
		if r == '"' {
			inString = true
			continue
		}
		switch r {
		case '(', '[', '{':
			stack = append(stack, stackItem{ch: r, pos: i})
		case ')', ']', '}':
			if len(stack) == 0 {
				continue
			}
			open := stack[len(stack)-1]
			if matchingBracket(open.ch) != r {
				continue
			}
			stack = stack[:len(stack)-1]
			spans = append(spans, span{start: open.pos, end: i + 1})
		}
	}
	return spans
}

func matchingBracket(ch rune) rune {
	switch ch {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	default:
		return 0
	}
}

func findFunctionCall(expr string) (callStart, callEnd, argsStart, argsEnd int, ok bool) {
	for i := 0; i < len(expr); i++ {
		if !isIdentStart(rune(expr[i])) {
			continue
		}
		j := i + 1
		for j < len(expr) && isIdentPart(rune(expr[j])) {
			j++
		}
		k := j
		for k < len(expr) && unicode.IsSpace(rune(expr[k])) {
			k++
		}
		if k >= len(expr) || expr[k] != '(' {
			i = j
			continue
		}

		end := findMatchingParen(expr, k)
		if end < 0 {
			return 0, 0, 0, 0, false
		}
		return i, end + 1, k + 1, end, true
	}
	return 0, 0, 0, 0, false
}

func findMatchingParen(s string, open int) int {
	depth := 0
	inString := false
	escaped := false
	for i := open; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == '(' {
			depth++
		} else if ch == ')' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitArgs(argsText string) []string {
	argsText = strings.TrimSpace(argsText)
	if argsText == "" {
		return nil
	}

	var args []string
	start := 0
	depthParen, depthBracket, depthBrace := 0, 0, 0
	inString := false
	escaped := false

	for i := 0; i < len(argsText); i++ {
		ch := argsText[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '(':
			depthParen++
		case ')':
			depthParen--
		case '[':
			depthBracket++
		case ']':
			depthBracket--
		case '{':
			depthBrace++
		case '}':
			depthBrace--
		case ',':
			if depthParen == 0 && depthBracket == 0 && depthBrace == 0 {
				part := strings.TrimSpace(argsText[start:i])
				if part != "" {
					args = append(args, part)
				}
				start = i + 1
			}
		}
	}

	last := strings.TrimSpace(argsText[start:])
	if last != "" {
		args = append(args, last)
	}
	return args
}

func wholeWordIndex(s, word string) int {
	for i := 0; i <= len(s)-len(word); i++ {
		if !strings.HasPrefix(s[i:], word) {
			continue
		}
		beforeOK := i == 0 || !isIdentPart(rune(s[i-1]))
		afterIdx := i + len(word)
		afterOK := afterIdx >= len(s) || !isIdentPart(rune(s[afterIdx]))
		if beforeOK && afterOK {
			return i
		}
	}
	return -1
}

func isIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// SubtreeSwapWithin swaps two bracketed subexpressions within a single expression.
func SubtreeSwapWithin(expr string, rng *rand.Rand) string {
	spans := findBracketSpans(expr)
	if len(spans) < 2 {
		return expr
	}
	i := rng.Intn(len(spans))
	j := rng.Intn(len(spans))
	for j == i {
		j = rng.Intn(len(spans))
	}
	a, b := spans[i], spans[j]
	if a.start > b.start {
		a, b = b, a
	}

	left := expr[a.start:a.end]
	right := expr[b.start:b.end]
	return expr[:a.start] + right + expr[a.end:b.start] + left + expr[b.end:]
}

// DebugDescribeMutation is useful for tests and diagnostics.
func DebugDescribeMutation(before, after string) string {
	return fmt.Sprintf("before=%q after=%q", before, after)
}

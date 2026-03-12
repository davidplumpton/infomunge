package preprocessor

import (
	"fmt"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/sourcemap"
	"strings"
)

// MaxRecursionDepth limits how deeply nested if-else expressions can be.
const MaxRecursionDepth = 100

const (
	// MaxRewriteIterations is intentionally much higher than MaxTransformIterations:
	// the core rewriter is a single byte-walking pass over raw input and can recurse
	// into rewritten branch bodies, so large inputs legitimately require thousands of
	// iterations before we should suspect non-progress.
	MaxRewriteIterations = 10000
	// MaxTransformIterations bounds fixpoint loops in stage transforms where each
	// pass should converge quickly (typically in a small number of iterations).
	MaxTransformIterations = 100
)

// Options holds configuration for the preprocessor
type Options struct {
	// AllowMultilineIfElse enables multiline if/else branch parsing.
	AllowMultilineIfElse bool
}

// PrepareForParsing transforms InfoMunge syntax into Go-compatible syntax for the parser.
// Returns an error if the expression exceeds recursion depth limits.
func PrepareForParsing(body string, opts Options) (string, []int, error) {
	// Strip // line comments before any other processing, since the Go
	// expression parser does not support them.
	body = stripLineComments(body)

	// Process regex literals first, before any other transformations,
	// to prevent /[a-z]+/ from being interpreted as an array literal.
	// Track exact source positions through each early transform.
	body, regexMapping := replaceRegexLiteralsWithMapping(body)

	prevBody := body
	body = wrapImplicitObjectLiteralBodies(body)
	wrapMapping := inferStageMapping(prevBody, body)

	input, wrapped := wrapTopLevelObjectLiteral(body)
	inputMap := sourcemap.Identity(body)
	if wrapped {
		wrapperPositions := make([]int, len(input))
		for i := range wrapperPositions {
			switch {
			case i == 0:
				wrapperPositions[i] = 0
			case i == len(input)-1:
				if len(body) == 0 {
					wrapperPositions[i] = 0
				} else {
					wrapperPositions[i] = len(body) - 1
				}
			default:
				wrapperPositions[i] = i - 1
			}
		}
		inputMap = sourcemap.New(body, input, wrapperPositions)
	}
	rewriter := newRewriter(input, opts)
	result, mapping, err := rewriter.RewriteWithDepth(0)

	// Compose full chain: result → rewriter input → post-wrap → post-regex → original
	composed := inputMap.Compose(result, mapping).Positions()
	composed = composeMappings(wrapMapping, composed)
	composed = composeMappings(regexMapping, composed)

	return result, composed, err
}

// Internal implementation of the rewriter
type rewriter struct {
	input        string
	result       []byte
	mapping      []int
	stack        []byte
	inString     bool
	stringEscape bool
	pos          int
	opts         Options
	err          error
}

func newRewriter(input string, opts Options) *rewriter {
	return &rewriter{
		input:   input,
		result:  make([]byte, 0, len(input)*2),
		mapping: make([]int, 0, len(input)*2),
		stack:   make([]byte, 0, 8),
		opts:    opts,
	}
}

func (r *rewriter) append(char byte, originalPos int) {
	r.result = append(r.result, char)
	r.mapping = append(r.mapping, originalPos)
}

func (r *rewriter) appendString(s string, originalPos int) {
	for i := 0; i < len(s); i++ {
		r.append(s[i], originalPos)
	}
}

func (r *rewriter) appendMappedString(s string, mapping []int, fallback int) {
	for i := 0; i < len(s); i++ {
		orig := fallback
		if i < len(mapping) {
			orig = mapping[i]
		}
		r.append(s[i], orig)
	}
}

// truncate shrinks both result and mapping to length n, keeping them in sync.
func (r *rewriter) truncate(n int) {
	r.result = r.result[:n]
	r.mapping = r.mapping[:n]
}

// assertMappingInvariant sets r.err if result and mapping have different lengths.
// This is an O(1) programming-error detector; callers must check r.err afterward.
func (r *rewriter) assertMappingInvariant() {
	if len(r.result) != len(r.mapping) {
		r.err = fmt.Errorf("mapping invariant violated: len(result)=%d, len(mapping)=%d at pos %d",
			len(r.result), len(r.mapping), r.pos)
	}
}

func (r *rewriter) updateStringEscape(ch byte) {
	if !r.inString {
		return
	}
	if r.stringEscape {
		r.stringEscape = false
		return
	}
	if ch == '\\' {
		r.stringEscape = true
	}
}

func (r *rewriter) Rewrite() (string, []int, error) {
	return r.RewriteWithDepth(0)
}

// RewriteWithDepth is the internal rewrite function that tracks recursion depth.
func (r *rewriter) RewriteWithDepth(depth int) (string, []int, error) {
	if depth > MaxRecursionDepth {
		r.err = unifiederrors.ParseErrorf("expression exceeds maximum recursion depth of %d", MaxRecursionDepth)
		return string(r.result), r.mapping, r.err
	}

	iterCount := 0
	maxIterations := MaxRewriteIterations
	// Scale by input length so large but valid inputs do not trip the static cap.
	// The x4 factor provides slack for rewrites that emit additional tokens while
	// still detecting pathological non-progress quickly.
	if scaled := len(r.input) * 4; scaled > maxIterations {
		maxIterations = scaled
	}
	for r.pos = 0; r.pos < len(r.input); r.pos++ {
		iterCount++
		if iterCount > maxIterations {
			r.err = unifiederrors.ParseErrorf(
				"rewriter iteration limit exceeded at pos %d (limit %d, input %d)",
				r.pos,
				maxIterations,
				len(r.input),
			)
			return string(r.result), r.mapping, r.err
		}
		if r.err != nil {
			return string(r.result), r.mapping, r.err
		}
		char := r.input[r.pos]

		// Handle double-quoted strings
		if r.handleDoubleQuote(char) {
			continue
		}

		// Handle single-quoted strings (convert to double-quoted)
		if r.handleSingleQuote(char) {
			continue
		}

		// Inside string, just append and continue
		if r.inString {
			r.updateStringEscape(char)
			r.append(char, r.pos)
			continue
		}

		// Handle transformations outside of strings
		if r.handleBreak() {
			continue
		}

		if r.handleContinue() {
			continue
		}

		if r.handleWhileLoop() {
			continue
		}

		if r.handleIfElse() {
			continue
		}

		if r.handleAndOr() {
			continue
		}

		if r.handleNot() {
			continue
		}

		if r.handleCoerceEquals() {
			continue
		}

		if r.handleRangeBuiltin() {
			continue
		}

		if r.handleUsing() {
			continue
		}

		if r.handleNewline(char) {
			continue
		}

		if r.handleDoBlock() {
			continue
		}

		if r.handleUpdateExprBlock() {
			continue
		}

		if r.handleOpenBrace(char) {
			continue
		}

		if r.handleCloseBrace(char) {
			continue
		}

		if r.handleUnquotedKey(char) {
			continue
		}

		r.append(char, r.pos)
	}

	r.assertMappingInvariant()
	if r.err != nil {
		return string(r.result), r.mapping, r.err
	}

	result := string(r.result)
	if containsIfKeywordOutsideStrings(result) {
		if depth+1 > MaxRecursionDepth {
			r.err = unifiederrors.ParseErrorf("expression exceeds maximum recursion depth of %d", MaxRecursionDepth)
			return result, r.mapping, r.err
		}
		firstPassMapping := r.mapping
		newRewriter := newRewriter(result, r.opts)
		var secondPassMapping []int
		var err error
		result, secondPassMapping, err = newRewriter.RewriteWithDepth(depth + 1)
		if err != nil {
			r.mapping = secondPassMapping
			return result, r.mapping, err
		}
		// Compose mappings: secondPass → firstPass → original input
		composed := make([]int, len(secondPassMapping))
		for i, pos := range secondPassMapping {
			if pos >= 0 && pos < len(firstPassMapping) {
				composed[i] = firstPassMapping[pos]
			} else {
				composed[i] = pos
			}
		}
		r.mapping = composed
	}

	pipeline := CreateModularPostProcessingPipeline()
	var pipelineErr error
	result, r.mapping, pipelineErr = pipeline.Execute(result, r.mapping)
	if pipelineErr != nil {
		return result, r.mapping, pipelineErr
	}

	if r.err != nil {
		return result, r.mapping, r.err
	}

	return result, r.mapping, nil
}

// stripLineComments removes // line comments from each line, respecting
// string literals. Line count is preserved for diagnostics.
func stripLineComments(body string) string {
	lines := strings.Split(body, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		stripped := stripSingleLineComment(line)
		result = append(result, stripped)
	}
	return strings.Join(result, "\n")
}

// stripSingleLineComment removes a // comment from a single line,
// respecting string literals.
func stripSingleLineComment(line string) string {
	var sc ScanState

	for i := 0; i < len(line); i++ {
		ch := line[i]
		sc.Advance(ch)
		if !sc.InString() && ch == '/' {
			if i+1 < len(line) && line[i+1] == '/' {
				return line[:i] + strings.Repeat(" ", len(line)-i)
			}
			if isRegexContext(line, i) {
				regexEnd, _, _ := parseRegexLiteral(line, i)
				if regexEnd > i {
					i = regexEnd - 1
					continue
				}
			}
		}
	}
	return line
}

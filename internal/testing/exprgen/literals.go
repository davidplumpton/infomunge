package exprgen

import (
	"math"
	"strconv"
	"strings"

	"pgregory.net/rapid"
)

// formatFloat formats a float64 as a string that Go's lexer will tokenize as
// token.FLOAT (not token.INT). It appends ".0" when FormatFloat produces an
// integer-looking string without a decimal point or exponent.
func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.Contains(s, ".") && !strings.ContainsAny(s, "eE") {
		s += ".0"
	}
	return s
}

// IntLiteral returns a generator of infomunge integer literal source strings.
// Includes zero, positive, negative, and large values. Negative values are
// emitted as "-N" which the parser handles as a unary minus expression.
func IntLiteral() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		strategy := rapid.IntRange(0, 5).Draw(t, "strategy")
		switch strategy {
		case 0: // zero
			return "0"
		case 1: // small positive
			return strconv.Itoa(rapid.IntRange(1, 100).Draw(t, "n"))
		case 2: // small negative
			return strconv.Itoa(-rapid.IntRange(1, 100).Draw(t, "n"))
		case 3: // medium range
			return strconv.Itoa(rapid.IntRange(-1_000_000, 1_000_000).Draw(t, "n"))
		case 4: // minimum signed integer exercises unary literal handling
			return strconv.Itoa(math.MinInt)
		default: // large
			n := rapid.IntRange(math.MinInt+1, math.MaxInt).Draw(t, "n")
			return strconv.Itoa(n)
		}
	})
}

// FloatLiteral returns a generator of infomunge float literal source strings.
// Produces values like "0.0", "3.14", "-2.5", "1e10".
func FloatLiteral() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		strategy := rapid.IntRange(0, 4).Draw(t, "strategy")
		var f float64
		switch strategy {
		case 0: // zero
			return "0.0"
		case 1: // small
			f = rapid.Float64Range(-100.0, 100.0).Draw(t, "f")
		case 2: // very small (near zero)
			f = rapid.Float64Range(-1e-10, 1e-10).Draw(t, "f")
		case 3: // large
			f = rapid.Float64Range(-1e15, 1e15).Draw(t, "f")
		default: // full range
			f = rapid.Float64Range(-math.MaxFloat64, math.MaxFloat64).Draw(t, "f")
		}
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return "0.0"
		}
		return formatFloat(f)
	})
}

// StringLiteral returns a generator of infomunge string literal source strings.
// Produces Go-style double-quoted strings that round-trip through strconv.Unquote.
func StringLiteral() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		strategy := rapid.IntRange(0, 4).Draw(t, "strategy")
		var s string
		switch strategy {
		case 0: // empty
			s = ""
		case 1: // short ASCII
			s = rapid.StringOfN(rapid.RuneFrom([]rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 _-.,!?")), 1, 20, -1).Draw(t, "s")
		case 2: // with special chars
			s = rapid.StringN(0, 10, -1).Draw(t, "s")
		case 3: // unicode
			s = rapid.String().Draw(t, "s")
		default: // longer
			s = rapid.StringN(0, 50, -1).Draw(t, "s")
		}
		return strconv.Quote(s)
	})
}

// BoolLiteral returns a generator of infomunge boolean literal source strings.
func BoolLiteral() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"true", "false"})
}

// NullLiteral returns a generator of infomunge null literal source strings.
// Both "null" and "nil" are valid in infomunge.
func NullLiteral() *rapid.Generator[string] {
	return rapid.SampledFrom([]string{"null", "nil"})
}

// Literal returns a generator that produces any literal source string
// (integer, float, string, boolean, or null).
func Literal() *rapid.Generator[string] {
	return rapid.OneOf(
		IntLiteral(),
		FloatLiteral(),
		StringLiteral(),
		BoolLiteral(),
		NullLiteral(),
	)
}

// DWLiteral returns a generator that produces literals compatible with both
// infomunge and DataWeave. Unlike Literal(), it only generates "null" (never
// "nil", which DataWeave does not recognize).
func DWLiteral() *rapid.Generator[string] {
	return rapid.OneOf(
		IntLiteral(),
		FloatLiteral(),
		StringLiteral(),
		BoolLiteral(),
		rapid.Just("null"),
	)
}

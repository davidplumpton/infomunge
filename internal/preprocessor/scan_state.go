package preprocessor

import "infomunge/internal/stringutils"

// Go type literal prefixes used as rewrite targets in the preprocessor.
const (
	GoObjectPrefix      = "map[string]interface{}{"
	GoObjectPrefixSpace = "map[string]interface{} {"
	GoArrayPrefix       = "[]interface{}{"
	GoEmptyObject       = "map[string]interface{}{}"
)

// isEscapedAt checks if the character at position i in string s is escaped
// by counting consecutive backslashes before it. An odd number means escaped.
// Example: "hello\"" -> quote at end is escaped (1 backslash)
// Example: "hello\\"  -> quote at end is NOT escaped (2 backslashes = escaped backslash)
func isEscapedAt(s string, i int) bool {
	return stringutils.IsEscapedAt(s, i)
}

// ScanState is an alias for stringutils.ScanState, kept here for
// backward compatibility with preprocessor-internal code.
type ScanState = stringutils.ScanState

// BranchPositions holds the positions of true and false branches in an if-else expression.
type BranchPositions struct {
	TrueStart  int // Start of the true branch expression
	TrueEnd    int // End of the true branch expression (exclusive)
	FalseStart int // Start of the false branch (0 if no else)
	FalseEnd   int // End of the false branch (0 if no else)
}

// BranchScanResult holds the result of scanning for a branch endpoint.
type BranchScanResult struct {
	BranchEnd int // End position of the branch expression
	ElsePos   int // Position of "else" keyword (-1 if none)
	ErrPos    int // Position of unmatched closer (-1 if no error)
	OK        bool
}

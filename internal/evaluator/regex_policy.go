package evaluator

import "strings"

// looksLikeRegexPattern reports whether a string contains syntax that should be
// interpreted as a regular expression rather than as literal text.
func looksLikeRegexPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*+?()[]{}|^$\\") ||
		(len(pattern) > 1 && strings.Contains(pattern, "."))
}

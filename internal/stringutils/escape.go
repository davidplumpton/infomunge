package stringutils

// IsEscapedAt reports whether the byte at index i is escaped by an odd number
// of consecutive backslashes immediately before it.
func IsEscapedAt(s string, i int) bool {
	if i <= 0 || i > len(s) {
		return false
	}

	backslashCount := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		backslashCount++
	}
	return backslashCount%2 == 1
}

// IsEscapedRuneAt reports whether the rune at index i is escaped by an odd
// number of consecutive backslashes immediately before it.
func IsEscapedRuneAt(runes []rune, i int) bool {
	if i <= 0 || i > len(runes) {
		return false
	}

	backslashCount := 0
	for j := i - 1; j >= 0 && runes[j] == '\\'; j-- {
		backslashCount++
	}
	return backslashCount%2 == 1
}

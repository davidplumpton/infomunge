package preprocessor

// IsOperatorChar returns true if ch is an operator character.
func IsOperatorChar(ch byte) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/' ||
		ch == '=' || ch == '<' || ch == '>' || ch == '!' ||
		ch == '(' || ch == '[' || ch == '{' || ch == ','
}

// IsClosingBracket returns true if ch is a closing bracket.
func IsClosingBracket(ch byte) bool {
	return ch == ')' || ch == ']' || ch == '}'
}

// IsOpeningBracket returns true if ch is an opening bracket.
func IsOpeningBracket(ch byte) bool {
	return ch == '(' || ch == '[' || ch == '{'
}

// IsIdentifierStart returns true if ch can start an identifier.
func IsIdentifierStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// IsIdentifierPart returns true if ch can be part of an identifier.
func IsIdentifierPart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '#'
}

// IsTypeExprChar returns true if ch can be part of a type expression.
// Type expressions include identifiers, whitespace, '|' (for unions), '?' (for optionals), and '.' (for namespaces).
func IsTypeExprChar(ch byte) bool {
	if IsIdentifierPart(ch) {
		return true
	}
	switch ch {
	case ' ', '\t', '|', '?', '.':
		return true
	default:
		return false
	}
}

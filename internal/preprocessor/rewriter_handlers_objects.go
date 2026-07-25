package preprocessor

import (
	"strings"
	"unicode"
)

func (r *rewriter) handleNewline(char byte) bool {
	if (char == '\n' || char == '\r') && len(r.stack) > 0 {
		r.append(' ', r.pos)
		return true
	}
	return false
}

func (r *rewriter) handleOpenBrace(char byte) bool {
	if char == '{' {
		if rewritten, endPos, ok := r.tryRewriteConditionalObjectLiteral(); ok {
			r.appendString(rewritten, r.pos)
			r.pos = endPos
			return true
		}
		r.appendString(GoObjectPrefix, r.pos)
		r.stack = append(r.stack, '}')
		return true
	}
	if char == '[' {
		if r.isIndexer() {
			r.append('[', r.pos)
			r.stack = append(r.stack, ']')
			return true
		}
		r.appendString(GoArrayPrefix, r.pos)
		r.stack = append(r.stack, '>')
		return true
	}
	return false
}

func (r *rewriter) tryRewriteConditionalObjectLiteral() (string, int, bool) {
	braceStart, braceEnd, ok := findMatchingDelimited(r.input, r.pos, '{', '}', true)
	if !ok || braceStart != r.pos {
		return "", 0, false
	}

	content := r.input[braceStart+1 : braceEnd]
	entries := splitObjectEntries(content)
	if len(entries) == 0 {
		return "", 0, false
	}

	parsed, conditionalFound := parseConditionalObjectEntries(entries)
	if !conditionalFound {
		return "", 0, false
	}

	rewrittenEntries := make([]string, 0, len(parsed))
	for _, entry := range parsed {
		rewrittenEntry, ok := r.rewriteConditionalObjectEntry(entry)
		if !ok {
			return "", 0, true
		}
		rewrittenEntries = append(rewrittenEntries, rewrittenEntry)
	}

	if len(rewrittenEntries) == 0 {
		return GoEmptyObject, braceEnd, true
	}

	return "merge(" + strings.Join(rewrittenEntries, ", ") + ")", braceEnd, true
}

type conditionalObjectEntry struct {
	expr          string
	cond          string
	isConditional bool
}

func parseConditionalObjectEntries(entries []string) ([]conditionalObjectEntry, bool) {
	parsed := make([]conditionalObjectEntry, 0, len(entries))
	conditionalFound := false

	for _, entry := range entries {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}

		expr, cond, isConditional := splitConditionalEntry(trimmed)
		if !isConditional {
			expr = trimmed
		}

		parsed = append(parsed, conditionalObjectEntry{
			expr:          expr,
			cond:          cond,
			isConditional: isConditional,
		})
		if isConditional {
			conditionalFound = true
		}
	}

	return parsed, conditionalFound
}

func (r *rewriter) rewriteConditionalObjectEntry(entry conditionalObjectEntry) (string, bool) {
	expr := stripOuterParens(entry.expr)
	entryLiteral := "{" + expr + "}"
	rewrittenEntry, ok := r.rewriteBranch(entryLiteral)
	if !ok {
		return "", false
	}

	if !entry.isConditional {
		return rewrittenEntry, true
	}

	condRewritten, ok := r.rewriteBranch(entry.cond)
	if !ok {
		return "", false
	}
	return "__ifelse(" + condRewritten + ", " + rewrittenEntry + ", map[string]interface{}{})", true
}

func splitObjectEntries(input string) []string {
	entries := make([]string, 0, 4)
	state := ScanState{}
	start := 0
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if ch == ',' && state.AtTopLevel() {
			entries = append(entries, input[start:i])
			start = i + 1
			continue
		}
		state.Advance(ch)
	}

	if start <= len(input) {
		entries = append(entries, input[start:])
	}
	return entries
}

func splitConditionalEntry(entry string) (string, string, bool) {
	state := ScanState{}
	for i := 0; i < len(entry); i++ {
		ch := entry[i]
		if state.AtTopLevel() && ch == ' ' && strings.HasPrefix(entry[i:], " if") {
			afterIf := i + 3
			afterIf = skipSpace(entry, afterIf, false)
			if afterIf < len(entry) && entry[afterIf] == '(' {
				condStart := afterIf
				_, condEnd, ok := findMatchingDelimited(entry, condStart, '(', ')', false)
				if ok && strings.TrimSpace(entry[condEnd+1:]) == "" {
					expr := strings.TrimSpace(entry[:i])
					cond := strings.TrimSpace(entry[condStart+1 : condEnd])
					return expr, cond, true
				}
			}
		}
		state.Advance(ch)
	}

	return "", "", false
}

func stripOuterParens(expr string) string {
	trimmed := strings.TrimSpace(expr)
	if len(trimmed) < 2 || trimmed[0] != '(' || trimmed[len(trimmed)-1] != ')' {
		return trimmed
	}

	_, end, ok := findMatchingDelimited(trimmed, 0, '(', ')', false)
	if !ok || end != len(trimmed)-1 {
		return trimmed
	}
	return strings.TrimSpace(trimmed[1:end])
}

func (r *rewriter) isIndexer() bool {
	j := len(r.result) - 1
	for j >= 0 && unicode.IsSpace(rune(r.result[j])) {
		j--
	}
	if j < 0 {
		return false
	}
	if r.followsImplicitLambdaOperator(j) {
		return false
	}
	if r.followsDefaultOperator(j) {
		return false
	}
	if r.followsConfiguredBinaryOperator(j) {
		return false
	}
	last := r.result[j]
	// Implicit lambda parameters are rewritten after the core syntax pass, so
	// their selectors must already be classified as indexing here.
	return unicode.IsLetter(rune(last)) || unicode.IsDigit(rune(last)) || last == '}' || last == ']' || last == ')' || last == '"' || last == '\'' || last == '$'
}

func (r *rewriter) followsDefaultOperator(end int) bool {
	// The default operator is rewritten after array literals. Preserve the
	// leading space in this suffix so identifiers ending in "default" still
	// treat a following bracket as an index selector.
	return strings.HasSuffix(string(r.result[:end+1]), " default")
}

func (r *rewriter) followsConfiguredBinaryOperator(end int) bool {
	prefix := string(r.result[:end+1])
	for _, config := range binaryOperatorConfigs {
		operator := strings.TrimSpace(config.Operator)
		if strings.HasSuffix(prefix, " "+operator) {
			return true
		}
	}
	return false
}

func (r *rewriter) followsImplicitLambdaOperator(end int) bool {
	prefix := string(r.result[:end+1])
	for _, operator := range implicitLambdaOperators {
		if strings.HasSuffix(prefix, strings.TrimSuffix(operator, " ")) {
			return true
		}
	}
	return false
}

func (r *rewriter) handleCloseBrace(char byte) bool {
	if char == '}' || char == ']' {
		closingChar := char
		isIndexer := false
		if len(r.stack) == 0 {
			r.setInvariantError("closing brace without opener at position %d", r.pos)
			return true
		}

		expected := r.stack[len(r.stack)-1]
		switch expected {
		case ']':
			if char != ']' {
				r.setInvariantError("mismatched closing brace '%c' at position %d", char, r.pos)
				return true
			}
			isIndexer = true
		case '}':
			if char != '}' {
				r.setInvariantError("mismatched closing brace '%c' at position %d", char, r.pos)
				return true
			}
		case '>':
			if char != ']' {
				r.setInvariantError("mismatched closing brace '%c' at position %d", char, r.pos)
				return true
			}
			closingChar = '}'
		default:
			r.setInvariantError("unexpected stack entry '%c' at position %d", expected, r.pos)
			return true
		}
		r.stack = r.stack[:len(r.stack)-1]

		if !isIndexer {
			r.addTrailingCommaIfNeeded()
		}
		r.append(closingChar, r.pos)
		return true
	}
	return false
}

func (r *rewriter) addTrailingCommaIfNeeded() {
	j := len(r.result) - 1
	for j >= 0 && unicode.IsSpace(rune(r.result[j])) {
		j--
	}
	if j >= 0 {
		lastNonSpace := r.result[j]
		if lastNonSpace != '{' && lastNonSpace != '[' && lastNonSpace != ',' {
			r.append(',', r.pos)
		}
	}
}

func (r *rewriter) handleUnquotedKey(char byte) bool {
	if unicode.IsLetter(rune(char)) || char == '_' {
		k := r.pos
		for k < len(r.input) && (unicode.IsLetter(rune(r.input[k])) || unicode.IsDigit(rune(r.input[k])) || r.input[k] == '_' || r.input[k] == '#') {
			k++
		}
		word := r.input[r.pos:k]

		m := k
		for m < len(r.input) && unicode.IsSpace(rune(r.input[m])) {
			m++
		}

		if m < len(r.input) && r.input[m] == ':' {
			if m+1 < len(r.input) && r.input[m+1] == ':' {
				return false
			}
			r.append('"', r.pos)
			for i := 0; i < len(word); i++ {
				r.append(word[i], r.pos+i)
			}
			r.append('"', r.pos)
			r.pos = k - 1
			return true
		}
	}
	return false
}

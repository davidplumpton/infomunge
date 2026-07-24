package preprocessor

import "strconv"

func replaceAsOperatorWithMapping(s string) (string, []int) {
	return replaceTypedOperatorWithMapping(s, " as ", "__coerce", true)
}

func replaceIsOperatorWithMapping(s string) (string, []int) {
	return replaceTypedOperatorWithMapping(s, " is ", "__isType", false)
}

func replaceTypedOperatorWithMapping(s string, token string, funcName string, allowConfig bool) (string, []int) {
	buf := newMappedBuffer(len(s) + 32)
	var topState ScanState
	i := 0
	tokenLen := len(token)

	for i < len(s) {
		if !topState.InString() && i+tokenLen <= len(s) && s[i:i+tokenLen] == token {
			leftStart := findTypedOperatorLeftOperandStartBytes(buf.bytes)
			if leftStart >= buf.Len() {
				appendOriginalWithState(buf, &topState, s, i, i+tokenLen)
				i += tokenLen
				continue
			}

			leftTrimStart, _ := trimSpaceBounds(buf.String(), leftStart, buf.Len())
			if leftTrimStart >= buf.Len() {
				appendOriginalWithState(buf, &topState, s, i, i+tokenLen)
				i += tokenLen
				continue
			}

			right, ok := scanTypedOperatorRightSpan(s, i+tokenLen, allowConfig)
			if !ok {
				appendOriginalWithState(buf, &topState, s, i, i+tokenLen)
				i += tokenLen
				continue
			}

			leftBytes, leftMapping := buf.Slice(leftTrimStart)
			// Keep separator whitespace between a preceding operator/delimiter
			// and the rewritten typed operand. Later transforms match exact
			// spaced tokens such as " ++ " and " ~ ".
			buf.Truncate(leftTrimStart)

			buf.AppendLiteral(funcName+"(", i)
			buf.AppendBytes(leftBytes, leftMapping)
			buf.AppendLiteral(", ", i)
			appendQuotedTypeExpr(buf, s, right)
			if right.ConfigArg != "" {
				buf.AppendLiteral(", ", right.ConfigStart)
				buf.AppendOriginal(s, right.ConfigStart, right.ConfigEnd)
			}
			buf.AppendLiteral(")", typedOperatorCloseMappingPos(right))
			i = right.Next
			continue
		}

		appendOriginalWithState(buf, &topState, s, i, i+1)
		i++
	}

	return buf.String(), buf.mapping
}

func appendOriginalWithState(buf *mappedBuffer, state *ScanState, src string, start int, end int) {
	buf.AppendOriginal(src, start, end)
	for i := start; i < end; i++ {
		state.Advance(src[i])
	}
}

func appendQuotedTypeExpr(buf *mappedBuffer, src string, span typedOperatorRightSpan) {
	quoted := strconv.Quote(span.TypeExpr)
	if quoted == "\""+span.TypeExpr+"\"" {
		buf.AppendLiteral("\"", span.TypeStart)
		buf.AppendOriginal(src, span.TypeStart, span.TypeEnd)
		buf.AppendLiteral("\"", span.TypeEnd-1)
		return
	}
	buf.AppendLiteral(quoted, span.TypeStart)
}

func typedOperatorCloseMappingPos(span typedOperatorRightSpan) int {
	if span.ConfigEnd > span.ConfigStart && span.ConfigStart >= 0 {
		return span.ConfigEnd - 1
	}
	return span.TypeEnd - 1
}

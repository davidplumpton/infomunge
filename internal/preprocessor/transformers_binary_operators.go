package preprocessor

import (
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/stringutils"
)

const (
	binaryOpDefault     = "default"
	binaryOpOnNull      = "onNull"
	binaryOpThen        = "then"
	binaryOpUpdate      = "update"
	binaryOpFind        = "find"
	binaryOpConcatenate = "concatenate"
	binaryOpRemove      = "remove"
	binaryOpSplitBy     = "splitBy"
	binaryOpJoinBy      = "joinBy"
	binaryOpTo          = "to"
	binaryOpMatch       = "match"
	binaryOpContains    = "contains"
	binaryOpMatches     = "matches"
	binaryOpScan        = "scan"
	binaryOpRepeat      = "repeat"
	binaryOpMod         = "mod"
)

var binaryOperatorConfigs = map[string]stringutils.BinaryOperatorConfig{
	binaryOpDefault: {
		Operator:     " default ",
		FuncName:     "__default",
		RightStopOps: []string{" and "},
	},
	binaryOpOnNull: {
		Operator: " onNull ",
		FuncName: "onNull",
	},
	binaryOpThen: {
		Operator: " then ",
		FuncName: "then",
	},
	binaryOpUpdate: {
		Operator:     " ~ ",
		FuncName:     "__update",
		ExtraStops:   []rune{'~'},
		RightStopOps: []string{" ~ "},
	},
	binaryOpFind: {
		Operator: " find ",
		FuncName: "find",
	},
	binaryOpConcatenate: {
		Operator:     " ++ ",
		FuncName:     "__concat",
		ExtraStops:   []rune{'+'},
		RightStopOps: []string{" ++ "},
	},
	binaryOpRemove: {
		Operator:     " -- ",
		FuncName:     "__remove",
		ExtraStops:   []rune{'-'},
		RightStopOps: []string{" -- "},
	},
	binaryOpSplitBy: {
		Operator:     " splitBy ",
		FuncName:     "splitBy",
		RightStopOps: []string{" splitBy ", " ++ "},
	},
	binaryOpJoinBy: {
		Operator:     " joinBy ",
		FuncName:     "joinBy",
		RightStopOps: []string{" joinBy ", " ++ "},
	},
	binaryOpTo: {
		Operator:        " to ",
		FuncName:        "to",
		RightStopOps:    []string{" to "},
		UseMinimalStops: true, // Allow negative numbers like -2 to 5.
	},
	binaryOpMatch: {
		Operator:     " match ",
		FuncName:     "match",
		RightStopOps: []string{" match "},
	},
	binaryOpContains: {
		Operator:     " contains ",
		FuncName:     "contains",
		RightStopOps: []string{" && ", " || ", " and ", " or ", " contains "},
	},
	binaryOpMatches: {
		Operator:     " matches ",
		FuncName:     "matches",
		RightStopOps: []string{" matches "},
	},
	binaryOpScan: {
		Operator:     " scan ",
		FuncName:     "scan",
		RightStopOps: []string{" scan "},
	},
	binaryOpRepeat: {
		Operator:     " repeat ",
		FuncName:     "repeat",
		RightStopOps: []string{" repeat "},
	},
	binaryOpMod: {
		Operator:     " mod ",
		FuncName:     "mod",
		RightStopOps: []string{" mod ", " repeat ", "==", "!=", "<", ">", "<=", ">=", " matches "},
	},
}

type binaryOperatorScanOverrides struct {
	leftOperandStart   func([]byte) int
	rightStopPredicate func(string, int, int) bool
}

var binaryOperatorScanOverridesByKey = map[string]binaryOperatorScanOverrides{
	binaryOpMod: {
		leftOperandStart:   findModuloLeftOperandStartBytes,
		rightStopPredicate: shouldStopModuloRightOperand,
	},
}

func replaceConfiguredBinaryOperatorWithMapping(s string, key string) (string, []int, error) {
	return replaceConfiguredBinaryOperatorWithMappingAndPeers(s, key, nil)
}

func replaceConfiguredBinaryOperatorWithMappingAndPeers(s string, key string, peerOps []string) (string, []int, error) {
	config, ok := binaryOperatorConfigs[key]
	if !ok {
		return s, identityMapping(len(s)), unifiederrors.ParseErrorf("missing binary operator config: %s", key)
	}
	rightStopOps := mergeOperatorStops(config.RightStopOps, peerOps)

	buf := newMappedBuffer(len(s) + len(config.FuncName) + 4)
	inString := false
	i := 0
	opLen := len(config.Operator)
	stopBytes := defaultStopBytes(config.ExtraStops)
	if config.UseMinimalStops {
		stopBytes = minimalStopBytes()
	}
	scanOverrides := binaryOperatorScanOverridesByKey[key]

	for i < len(s) {
		if s[i] == '"' && !stringutils.IsEscapedAt(s, i) {
			inString = !inString
			buf.AppendOriginal(s, i, i+1)
			i++
			continue
		}

		if !inString && i+opLen <= len(s) && s[i:i+opLen] == config.Operator {
			leftStart := findLeftOperandStartBytesWithIgnoredOperators(buf.bytes, stopBytes, peerOps)
			if scanOverrides.leftOperandStart != nil {
				leftStart = scanOverrides.leftOperandStart(buf.bytes)
			}
			if leftStart >= buf.Len() {
				buf.AppendOriginal(s, i, i+opLen)
				i += opLen
				continue
			}

			leftTrimStart, _ := trimSpaceBounds(buf.String(), leftStart, buf.Len())
			if leftTrimStart >= buf.Len() {
				buf.AppendOriginal(s, i, i+opLen)
				i += opLen
				continue
			}

			leftBytes, leftMapping := buf.Slice(leftTrimStart)
			buf.Truncate(leftStart)
			var typedSuffixBytes []byte
			var typedSuffixMapping []int
			if key == binaryOpDefault {
				// DataWeave applies an ungrouped trailing type annotation to
				// the value selected by default, while an explicitly grouped
				// coercion remains the left operand of default.
				leftBytes, leftMapping, typedSuffixBytes, typedSuffixMapping = splitTrailingTypedOperator(leftBytes, leftMapping)
			}

			rightStart := i + opLen
			rightTrimStart, rightTrimEnd, rightEnd, ok := scanRightOperandBounds(s, rightStart, rightOperandScanConfig{
				StopOps:       rightStopOps,
				StopPredicate: scanOverrides.rightStopPredicate,
			})
			if !ok {
				buf.AppendBytes(leftBytes, leftMapping)
				buf.AppendBytes(typedSuffixBytes, typedSuffixMapping)
				buf.AppendOriginal(s, i, i+opLen)
				i += opLen
				continue
			}

			buf.AppendLiteral(config.FuncName+"(", i)
			buf.AppendBytes(leftBytes, leftMapping)
			buf.AppendLiteral(", ", i)
			buf.AppendOriginal(s, rightTrimStart, rightTrimEnd)
			buf.AppendLiteral(")", rightTrimEnd-1)
			buf.AppendBytes(typedSuffixBytes, typedSuffixMapping)
			i = rightEnd
			continue
		}

		buf.AppendOriginal(s, i, i+1)
		i++
	}

	return buf.String(), buf.mapping, nil
}

func splitTrailingTypedOperator(data []byte, mapping []int) ([]byte, []int, []byte, []int) {
	input := string(data)
	var state ScanState

	for pos := 0; pos < len(input); pos++ {
		if !state.InString() && state.Depth() == 0 {
			for _, candidate := range []struct {
				token       string
				allowConfig bool
			}{
				{token: " as ", allowConfig: true},
				{token: " is ", allowConfig: false},
			} {
				if pos+len(candidate.token) > len(input) || input[pos:pos+len(candidate.token)] != candidate.token {
					continue
				}
				span, ok := scanTypedOperatorRightSpan(input, pos+len(candidate.token), candidate.allowConfig)
				if !ok {
					continue
				}
				end := span.Next
				for end < len(input) && isWhitespace(input[end]) {
					end++
				}
				if end == len(input) {
					return data[:pos], mapping[:pos], data[pos:], mapping[pos:]
				}
			}
		}
		state.Advance(input[pos])
	}

	return data, mapping, nil, nil
}

func mergeOperatorStops(existing, peers []string) []string {
	merged := make([]string, 0, len(existing)+len(peers))
	seen := make(map[string]struct{}, len(existing)+len(peers))
	for _, stops := range [][]string{existing, peers} {
		for _, stop := range stops {
			if _, ok := seen[stop]; ok {
				continue
			}
			seen[stop] = struct{}{}
			merged = append(merged, stop)
		}
	}
	return merged
}

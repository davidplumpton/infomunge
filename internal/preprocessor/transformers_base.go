package preprocessor

import (
	"infomunge/internal/stringutils"
)

// replaceAndOrOutsideStrings replaces " and " and " or " with operators.
func replaceAndOrOutsideStrings(s string) string {
	return stringutils.ReplaceOperatorsOutsideStrings(s, []stringutils.OperatorReplacement{
		{Pattern: " and ", Replacement: []rune(" && ")},
		{Pattern: " or ", Replacement: []rune(" || ")},
	})
}

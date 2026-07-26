package preprocessor

// CollectionOperators is the canonical list of collection operators used for
// expression boundary lookahead.
var CollectionOperators = []string{
	"map", "filter", "reduce", "flatMap", "groupBy", "pluck",
	"sort", "orderBy", "maxBy", "minBy", "distinctBy", "filterObject", "mapObject",
}

func isCollectionOperatorAt(s string, pos int) bool {
	if pos < 0 || pos >= len(s) {
		return false
	}
	for _, op := range CollectionOperators {
		opWithSpace := op + " "
		if pos+len(opWithSpace) <= len(s) && s[pos:pos+len(opWithSpace)] == opWithSpace {
			return true
		}
	}
	return false
}

func isCollectionOperatorWithSpacesAt(s string, pos int) bool {
	if pos < 0 || pos >= len(s) || s[pos] != ' ' {
		return false
	}
	for _, op := range CollectionOperators {
		opWithSpaces := " " + op + " "
		if pos+len(opWithSpaces) <= len(s) && s[pos:pos+len(opWithSpaces)] == opWithSpaces {
			return true
		}
	}
	return false
}

func isArrayCollectionOperatorWithSpacesAt(s string, pos int) bool {
	if pos < 0 || pos >= len(s) || s[pos] != ' ' {
		return false
	}
	for _, op := range []string{
		"map", "filter", "reduce", "flatMap", "groupBy",
		"sort", "orderBy", "maxBy", "minBy", "distinctBy",
	} {
		opWithSpaces := " " + op + " "
		if pos+len(opWithSpaces) <= len(s) && s[pos:pos+len(opWithSpaces)] == opWithSpaces {
			return true
		}
	}
	return false
}

func isCollectionOperatorAtRunes(runes []rune, pos int) bool {
	if pos < 0 || pos >= len(runes) {
		return false
	}
	for _, op := range CollectionOperators {
		opRunes := []rune(op + " ")
		if pos+len(opRunes) <= len(runes) && string(runes[pos:pos+len(opRunes)]) == string(opRunes) {
			return true
		}
	}
	return false
}

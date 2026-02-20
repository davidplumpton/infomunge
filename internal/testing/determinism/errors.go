package determinism

import "regexp"

var pointerAddressPattern = regexp.MustCompile(`0x[0-9a-fA-F]+`)

// EqualErrors compares evaluation errors while ignoring unstable pointer-address text.
func EqualErrors(first, second error) bool {
	if first == nil || second == nil {
		return first == second
	}
	return normalizeErrorText(first.Error()) == normalizeErrorText(second.Error())
}

func normalizeErrorText(text string) string {
	return pointerAddressPattern.ReplaceAllString(text, "<ptr>")
}

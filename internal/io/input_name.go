package input

import (
	"regexp"
	"strings"

	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/evaluator"
)

var inputNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// NormalizeAndValidateInputName trims and validates an input variable name.
func NormalizeAndValidateInputName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", unifiederrors.ValidationError("input name is required")
	}
	if !inputNamePattern.MatchString(name) {
		return "", unifiederrors.ValidationErrorf("invalid input name %q: must match [A-Za-z_][A-Za-z0-9_]*", name)
	}
	if evaluator.IsReservedBindingName(name) {
		return "", unifiederrors.ValidationErrorf("input name %q is reserved for runtime metadata", name)
	}
	return name, nil
}

package errors

import (
	stderrors "errors"
	"go/token"
	"testing"
)

func TestConstructorsPreserveTypeAndMessage(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		errType ErrorType
		message string
	}{
		{name: "parse", err: ParseError("bad syntax"), errType: TypeParse, message: "bad syntax"},
		{name: "formatted eval", err: EvalErrorf("unknown %s", "name"), errType: TypeEval, message: "unknown name"},
		{name: "formatted IO", err: IOErrorf("read %s", "input"), errType: TypeIO, message: "read input"},
		{name: "validation", err: ValidationError("invalid value"), errType: TypeValidate, message: "invalid value"},
		{name: "internal", err: InternalError("unexpected state"), errType: TypeInternal, message: "unexpected state"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Type != tt.errType {
				t.Fatalf("Type = %q, want %q", tt.err.Type, tt.errType)
			}
			if tt.err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", tt.err.Message, tt.message)
			}
		})
	}
}

func TestEvalPositionalUsesFileSetPosition(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("example.im", -1, 20)
	file.AddLine(5)

	err := EvalPositional("unknown variable", file.Pos(7), fset)

	if err.Position == nil {
		t.Fatal("Position is nil")
	}
	if got, want := err.Error(), "EvalError at example.im:2:3: unknown variable"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestEvalPositionalPreservesGeneratedPositionWithoutFileSet(t *testing.T) {
	err := EvalPositional("unknown variable", token.Pos(7), nil)

	if err.Position != nil {
		t.Fatalf("Position = %#v, want nil until source-map resolution", err.Position)
	}
	if got, ok := err.GeneratedPosition(); !ok || got != token.Pos(7) {
		t.Fatalf("GeneratedPosition() = (%d, %t), want (7, true)", got, ok)
	}
	if got, want := err.Error(), "EvalError: unknown variable"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestEvalPositionalDoesNotAttachInvalidFileSetPosition(t *testing.T) {
	err := EvalPositional("unknown variable", token.Pos(7), token.NewFileSet())

	if err.Position != nil {
		t.Fatalf("Position = %#v, want nil for an unresolved file-set position", err.Position)
	}
	if got, ok := err.GeneratedPosition(); !ok || got != token.Pos(7) {
		t.Fatalf("GeneratedPosition() = (%d, %t), want (7, true)", got, ok)
	}
}

func TestWrapPreservesCauseAndMatchesType(t *testing.T) {
	cause := stderrors.New("decode failed")
	err := WrapValidationf(cause, "invalid %s", "payload")

	if !stderrors.Is(err, cause) {
		t.Fatal("wrapped error does not match its cause")
	}
	if !stderrors.Is(err, ValidationError("")) {
		t.Fatal("wrapped error does not match its error type")
	}
	if stderrors.Is(err, ParseError("")) {
		t.Fatal("wrapped error unexpectedly matches a different error type")
	}
	if got, want := err.Error(), "ValidationError: invalid payload"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestWithPositionUsesExplicitPosition(t *testing.T) {
	err := ParseError("bad token").WithPosition(token.Position{
		Filename: "input.im",
		Offset:   8,
		Line:     3,
		Column:   4,
	})

	if err.Position == nil {
		t.Fatal("Position is nil")
	}
	if got, want := err.Error(), "ParseError at input.im:3:4: bad token"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

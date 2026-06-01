package test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestCaptureStdoutDrainsLargeOutput(t *testing.T) {
	const outputSize = 256 * 1024
	want := strings.Repeat("x", outputSize)

	got, err := captureStdout(func() error {
		_, err := fmt.Fprint(os.Stdout, want)
		return err
	})
	if err != nil {
		t.Fatalf("captureStdout returned error: %v", err)
	}
	if got != want {
		t.Fatalf("captured output length = %d, want %d", len(got), len(want))
	}
}

func TestCaptureStdoutRestoresAfterError(t *testing.T) {
	oldStdout := os.Stdout
	wantErr := errors.New("runner failed")

	got, err := captureStdout(func() error {
		fmt.Fprint(os.Stdout, "before error")
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("captureStdout error = %v, want %v", err, wantErr)
	}
	if got != "before error" {
		t.Fatalf("captured output = %q, want %q", got, "before error")
	}
	if os.Stdout != oldStdout {
		t.Fatal("os.Stdout was not restored after error")
	}
}

func TestCaptureStdoutRestoresAndCleansUpAfterPanic(t *testing.T) {
	oldStdout := os.Stdout
	var recovered any

	func() {
		defer func() {
			recovered = recover()
		}()

		_, _ = captureStdout(func() error {
			fmt.Fprint(os.Stdout, "before panic")
			panic("controlled stdout capture panic")
		})
	}()

	if recovered != "controlled stdout capture panic" {
		t.Fatalf("recovered panic = %v", recovered)
	}
	if os.Stdout != oldStdout {
		t.Fatal("os.Stdout was not restored after panic")
	}

	got, err := captureStdout(func() error {
		fmt.Fprint(os.Stdout, "after panic")
		return nil
	})
	if err != nil {
		t.Fatalf("captureStdout after panic returned error: %v", err)
	}
	if got != "after panic" {
		t.Fatalf("captured output after panic = %q, want %q", got, "after panic")
	}
}

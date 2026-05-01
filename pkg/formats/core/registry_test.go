package core

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestRegistryCanRunWithoutCodecPackage(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterReader("application/x-test", func(content string) (interface{}, error) {
		return Object{"content": content}, nil
	})
	registry.RegisterWriter("application/x-test", func(result interface{}) (string, error) {
		return fmt.Sprintf("encoded:%v", result), nil
	})
	registry.RegisterExtension(".test", "application/x-test")

	reader, err := registry.GetReader("application/x-test")
	if err != nil {
		t.Fatalf("expected registered reader, got error: %v", err)
	}
	value, err := reader("payload")
	if err != nil {
		t.Fatalf("unexpected reader error: %v", err)
	}
	if got := value.(Object)["content"]; got != "payload" {
		t.Fatalf("expected payload value, got %v", got)
	}

	writer, err := registry.GetWriter("application/x-test")
	if err != nil {
		t.Fatalf("expected registered writer, got error: %v", err)
	}
	formatted, err := writer("payload")
	if err != nil {
		t.Fatalf("unexpected writer error: %v", err)
	}
	if formatted != "encoded:payload" {
		t.Fatalf("expected writer output, got %q", formatted)
	}
	if got := registry.DetectMimeType("file.TEST"); got != "application/x-test" {
		t.Fatalf("expected extension lookup, got %q", got)
	}
}

func TestRegistryOptionsHandlersFollowAliases(t *testing.T) {
	registry := NewRegistry()

	registry.RegisterReadOptionsHandler("application/x-test", func(content string, options Object) (interface{}, error) {
		return Object{"content": content, "mode": options["mode"]}, nil
	})
	registry.RegisterWriteOptionsHandler("application/x-test", func(result interface{}, options Object) (string, error) {
		return fmt.Sprintf("%v:%v", result, options["mode"]), nil
	})
	registry.RegisterOptionsAlias("application/vnd.test", "application/x-test")

	readHandler, ok := registry.GetReadOptionsHandler("application/vnd.test")
	if !ok {
		t.Fatal("expected read options handler through alias")
	}
	readValue, err := readHandler("payload", Object{"mode": "alias"})
	if err != nil {
		t.Fatalf("unexpected read handler error: %v", err)
	}
	if got := readValue.(Object)["mode"]; got != "alias" {
		t.Fatalf("expected alias option, got %v", got)
	}

	writeHandler, ok := registry.GetWriteOptionsHandler("application/vnd.test")
	if !ok {
		t.Fatal("expected write options handler through alias")
	}
	written, err := writeHandler("payload", Object{"mode": "alias"})
	if err != nil {
		t.Fatalf("unexpected write handler error: %v", err)
	}
	if written != "payload:alias" {
		t.Fatalf("expected aliased handler output, got %q", written)
	}
}

func TestRegistryConcurrentRegisterAndLookup(t *testing.T) {
	registry := NewRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			mime := fmt.Sprintf("application/x-test-%d", i)
			registry.RegisterReader(mime, func(content string) (interface{}, error) {
				return content, nil
			})
			registry.RegisterExtension(fmt.Sprintf(".x%d", i), mime)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				_, _ = registry.GetReader("application/x-test-0")
				_ = registry.DetectMimeType("sample.x0")
			}
		}()
	}
	wg.Wait()

	if _, err := registry.GetReader("application/x-test-0"); err != nil {
		t.Fatalf("expected concurrent registration to succeed, got error: %v", err)
	}
	if got := registry.DetectMimeType("sample.x0"); got != "application/x-test-0" {
		t.Fatalf("expected extension lookup after concurrent registration, got %q", got)
	}
}

func TestResultHelpers(t *testing.T) {
	ok := Ok("value")
	if !ok.IsOk() || ok.IsErr() {
		t.Fatalf("expected ok result")
	}
	if got := ok.OrElse("fallback"); got != "value" {
		t.Fatalf("expected ok value, got %q", got)
	}

	errResult := Err[string](errors.New("boom"))
	if errResult.IsOk() || !errResult.IsErr() {
		t.Fatalf("expected error result")
	}
	if got := errResult.OrElse("fallback"); got != "fallback" {
		t.Fatalf("expected fallback value, got %q", got)
	}
}

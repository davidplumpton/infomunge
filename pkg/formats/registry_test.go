package formats

import (
	"fmt"
	"sync"
	"testing"
)

func cloneReaders(src map[string]Reader) map[string]Reader {
	dst := make(map[string]Reader, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneWriters(src map[string]Writer) map[string]Writer {
	dst := make(map[string]Writer, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneObjectReaders(src map[string]ObjectReader) map[string]ObjectReader {
	dst := make(map[string]ObjectReader, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneArrayReaders(src map[string]ArrayReader) map[string]ArrayReader {
	dst := make(map[string]ArrayReader, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneExtensions(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func withIsolatedRegistry(t *testing.T) {
	t.Helper()

	registryMu.Lock()
	origReaders := cloneReaders(readers)
	origWriters := cloneWriters(writers)
	origObjectReaders := cloneObjectReaders(objectReaders)
	origArrayReaders := cloneArrayReaders(arrayReaders)
	origExtensions := cloneExtensions(extensions)

	readers = make(map[string]Reader)
	writers = make(map[string]Writer)
	objectReaders = make(map[string]ObjectReader)
	arrayReaders = make(map[string]ArrayReader)
	extensions = make(map[string]string)
	registryMu.Unlock()

	t.Cleanup(func() {
		registryMu.Lock()
		readers = origReaders
		writers = origWriters
		objectReaders = origObjectReaders
		arrayReaders = origArrayReaders
		extensions = origExtensions
		registryMu.Unlock()
	})
}

func TestRegisterReader_RuntimeRegistration(t *testing.T) {
	withIsolatedRegistry(t)

	RegisterReader("application/x-test", func(content string) (interface{}, error) {
		return "ok:" + content, nil
	})

	r, err := GetReader("application/x-test")
	if err != nil {
		t.Fatalf("expected registered reader, got error: %v", err)
	}
	value, err := r("payload")
	if err != nil {
		t.Fatalf("unexpected reader error: %v", err)
	}
	if value != "ok:payload" {
		t.Fatalf("expected transformed payload, got %v", value)
	}
}

func TestDetectMimeType_CaseInsensitiveExtension(t *testing.T) {
	withIsolatedRegistry(t)

	RegisterExtension(".json", "application/json")

	if got := DetectMimeType("FILE.JSON"); got != "application/json" {
		t.Fatalf("expected application/json for uppercase extension, got %q", got)
	}
	if got := MimeTypeForFormat("JSON"); got != "application/json" {
		t.Fatalf("expected application/json from format lookup, got %q", got)
	}
}

func TestRegistryConcurrentRegisterAndLookup(t *testing.T) {
	withIsolatedRegistry(t)

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			mime := fmt.Sprintf("application/x-test-%d", i)
			RegisterReader(mime, func(content string) (interface{}, error) {
				return content, nil
			})
			RegisterExtension(fmt.Sprintf(".x%d", i), mime)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				_, _ = GetReader("application/x-test-0")
				_ = DetectMimeType("sample.x0")
			}
		}()
	}
	wg.Wait()

	if _, err := GetReader("application/x-test-0"); err != nil {
		t.Fatalf("expected concurrent registration to succeed, got error: %v", err)
	}
	if got := DetectMimeType("sample.x0"); got != "application/x-test-0" {
		t.Fatalf("expected extension lookup after concurrent registration, got %q", got)
	}
}

func TestBinaryFormatLookup(t *testing.T) {
	if got := MimeTypeForFormat("binary"); got != "application/octet-stream" {
		t.Fatalf("expected application/octet-stream for format 'binary', got %q", got)
	}
	if got := MimeTypeForFormat("bin"); got != "application/octet-stream" {
		t.Fatalf("expected application/octet-stream for format 'bin', got %q", got)
	}
	if got := DetectMimeType("payload.bin"); got != "application/octet-stream" {
		t.Fatalf("expected application/octet-stream for .bin extension, got %q", got)
	}
}

func TestAvroFormatLookup(t *testing.T) {
	if got := MimeTypeForFormat("avro"); got != "application/avro" {
		t.Fatalf("expected application/avro for format 'avro', got %q", got)
	}
	if got := DetectMimeType("payload.avro"); got != "application/avro" {
		t.Fatalf("expected application/avro for .avro extension, got %q", got)
	}
}

package coremodules

import (
	"io/fs"
	"strings"
	"testing"
)

func TestStandardModuleSpecsCoverAllEmbeddedIMFiles(t *testing.T) {
	entries, err := fs.ReadDir(standardModuleFiles, ".")
	if err != nil {
		t.Fatalf("read embedded module directory: %v", err)
	}

	embeddedCount := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".im") {
			continue
		}
		embeddedCount++
	}

	if len(standardModuleFileBySpec) != embeddedCount {
		t.Fatalf("standard module spec map has %d entries, but found %d embedded .im files", len(standardModuleFileBySpec), embeddedCount)
	}
}

func TestSourceReturnsModuleContent(t *testing.T) {
	for spec := range standardModuleFileBySpec {
		content, err := Source(spec)
		if err != nil {
			t.Fatalf("source(%q) returned error: %v", spec, err)
		}
		if !strings.Contains(content, "%dw") {
			t.Fatalf("source(%q) did not return DataWeave module content", spec)
		}
	}
}

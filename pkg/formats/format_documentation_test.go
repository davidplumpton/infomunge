package formats

import (
	formatcore "infomunge/pkg/formats/core"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var builtInFormatRegistrations formatcore.RegistrationSnapshot

func TestMain(m *testing.M) {
	builtInFormatRegistrations = registry.RegistrationSnapshot()
	os.Exit(m.Run())
}

func TestFormatDocumentationMatchesBuiltInRegistry(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "FORMATS.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read format documentation: %v", err)
	}

	rows := parseFormatDocumentation(t, string(content))
	documentedReaders := make(map[string]struct{})
	documentedWriters := make(map[string]struct{})
	documentedExtensions := make(map[string]string)
	documentedAliases := make(map[string]string)
	documentedOptions := make(map[string]struct{})
	documentedMIMEs := make(map[string]string)

	for _, row := range rows {
		mimeTypes := append([]string{row.canonicalMIME}, row.aliases...)
		for _, mimeType := range mimeTypes {
			if previous, exists := documentedMIMEs[mimeType]; exists {
				t.Fatalf("MIME type %q is documented in both %q and %q", mimeType, previous, row.canonicalMIME)
			}
			documentedMIMEs[mimeType] = row.canonicalMIME
			if row.input {
				documentedReaders[mimeType] = struct{}{}
			}
			if row.output {
				documentedWriters[mimeType] = struct{}{}
			}
		}
		for _, alias := range row.aliases {
			documentedAliases[alias] = row.canonicalMIME
		}
		for _, extension := range row.extensions {
			if previous, exists := documentedExtensions[extension]; exists {
				t.Fatalf("extension %q is documented for both %q and %q", extension, previous, row.canonicalMIME)
			}
			documentedExtensions[extension] = row.canonicalMIME
		}
		if strings.Contains(strings.ToLower(row.fidelity), "options:") {
			documentedOptions[row.canonicalMIME] = struct{}{}
		}
	}

	assertStringSliceEqual(t, "reader MIME types", sortedSet(documentedReaders), builtInFormatRegistrations.Readers)
	assertStringSliceEqual(t, "writer MIME types", sortedSet(documentedWriters), builtInFormatRegistrations.Writers)
	if !reflect.DeepEqual(documentedExtensions, builtInFormatRegistrations.Extensions) {
		t.Errorf("documented extensions do not match registry\n documented: %#v\n registry:   %#v", documentedExtensions, builtInFormatRegistrations.Extensions)
	}

	for alias, canonical := range builtInFormatRegistrations.OptionMIMEAliases {
		if documentedAliases[alias] != canonical {
			t.Errorf("registry option alias %q -> %q is not documented in that canonical row", alias, canonical)
		}
	}

	optionMIMEs := append([]string{}, builtInFormatRegistrations.ReadOptionsMIMEs...)
	optionMIMEs = append(optionMIMEs, builtInFormatRegistrations.WriteOptionsMIMEs...)
	for _, mimeType := range optionMIMEs {
		if _, ok := documentedOptions[mimeType]; !ok {
			t.Errorf("registered option handler for %q has no Options: description", mimeType)
		}
	}
}

type documentedFormat struct {
	canonicalMIME string
	aliases       []string
	extensions    []string
	input         bool
	output        bool
	fidelity      string
}

func parseFormatDocumentation(t *testing.T, content string) []documentedFormat {
	t.Helper()

	const startMarker = "<!-- format-matrix:start -->"
	const endMarker = "<!-- format-matrix:end -->"
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if start == -1 || end == -1 || end <= start {
		t.Fatalf("format documentation must contain ordered %q and %q markers", startMarker, endMarker)
	}

	var rows []documentedFormat
	for _, line := range strings.Split(content[start+len(startMarker):end], "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) != 7 {
			t.Fatalf("format matrix row must contain 7 columns, got %d in %q", len(cells), line)
		}
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		if cells[0] == "Format" || strings.HasPrefix(cells[0], "---") {
			continue
		}

		row := documentedFormat{
			canonicalMIME: parseCodeValue(cells[1]),
			aliases:       parseCodeList(cells[2]),
			extensions:    parseCodeList(cells[3]),
			input:         parseSupportCell(t, "Input", cells[4], line),
			output:        parseSupportCell(t, "Output", cells[5], line),
			fidelity:      cells[6],
		}
		if row.canonicalMIME == "" {
			t.Fatalf("format matrix row has no canonical MIME type: %q", line)
		}
		if row.fidelity == "" {
			t.Fatalf("format matrix row has no fidelity or limitations: %q", line)
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		t.Fatal("format matrix contains no data rows")
	}
	return rows
}

func parseSupportCell(t *testing.T, column, value, line string) bool {
	t.Helper()
	switch value {
	case "Yes":
		return true
	case "No":
		return false
	default:
		t.Fatalf("%s support must be Yes or No in %q", column, line)
		return false
	}
}

func parseCodeList(value string) []string {
	if value == "—" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if parsed := parseCodeValue(part); parsed != "" {
			values = append(values, parsed)
		}
	}
	return values
}

func parseCodeValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), "`")
}

func sortedSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func assertStringSliceEqual(t *testing.T, label string, documented, registered []string) {
	t.Helper()
	if !reflect.DeepEqual(documented, registered) {
		t.Errorf("documented %s do not match registry\n documented: %q\n registry:   %q", label, documented, registered)
	}
}

package evaluator

import (
	"go/ast"
	"testing"
)

func TestMetadataSizeCountsUnicodeCharacters(t *testing.T) {
	call := &ast.CallExpr{Fun: &ast.Ident{Name: "__metadata"}}
	got, err := callBuiltinMetadata([]Value{"é🙂", "size"}, call)
	if err != nil {
		t.Fatalf("callBuiltinMetadata() error = %v", err)
	}
	if got != 2 {
		t.Fatalf("callBuiltinMetadata() = %v, want 2", got)
	}
}

func TestHaveSizeCountsUnicodeCharacters(t *testing.T) {
	result := haveSize(2)("é🙂")
	if !result.Success {
		t.Fatalf("haveSize(2) rejected two-character string: %s", result.Message)
	}
}

func TestStringCardinalityPreservesBinaryByteCounts(t *testing.T) {
	binary := []byte{0xc3, 0xa9, 0xf0, 0x9f, 0x99, 0x82}
	sizeOfCall := &ast.CallExpr{Fun: &ast.Ident{Name: "sizeOf"}}
	metadataCall := &ast.CallExpr{Fun: &ast.Ident{Name: "__metadata"}}

	size, err := callBuiltinSizeOf([]Value{binary}, sizeOfCall)
	if err != nil {
		t.Fatalf("callBuiltinSizeOf() error = %v", err)
	}
	if size != len(binary) {
		t.Fatalf("callBuiltinSizeOf() = %v, want %d", size, len(binary))
	}

	metadataSize, err := callBuiltinMetadata([]Value{binary, "size"}, metadataCall)
	if err != nil {
		t.Fatalf("callBuiltinMetadata() error = %v", err)
	}
	if metadataSize != len(binary) {
		t.Fatalf("callBuiltinMetadata() = %v, want %d", metadataSize, len(binary))
	}

	result := haveSize(len(binary))(binary)
	if !result.Success {
		t.Fatalf("haveSize(%d) rejected binary: %s", len(binary), result.Message)
	}
}

package formats

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadWithOptions_XLSXStructured_MultiSheetRoundTrip(t *testing.T) {
	input := Object{
		"Employees": Array{
			Object{"name": "Alice", "age": 30, "active": true},
			Object{"name": "Bob", "age": 41, "active": false},
		},
		"Sales": Array{
			Object{"region": "west", "amount": 12.5},
		},
	}

	encoded, err := FormatWithOptions(input, "application/xlsx", Object{"structured": true})
	if err != nil {
		t.Fatalf("unexpected format error: %v", err)
	}

	decoded, err := ReadWithOptions(encoded, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Object{"structured": true})
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}

	expected := Object{
		"Employees": Array{
			Object{"active": true, "age": 30.0, "name": "Alice"},
			Object{"active": false, "age": 41.0, "name": "Bob"},
		},
		"Sales": Array{
			Object{"amount": 12.5, "region": "west"},
		},
	}

	if !reflect.DeepEqual(decoded, expected) {
		t.Fatalf("expected %#v, got %#v", expected, decoded)
	}
}

func TestReadWithOptions_XLSXStructured_RequiresStructuredFlag(t *testing.T) {
	_, err := ReadWithOptions("PK\x03\x04not-a-real-xlsx", "application/xlsx", Object{"strict": true})
	if err == nil {
		t.Fatal("expected error when options are provided without structured=true")
	}
	if !strings.Contains(err.Error(), "structured=true") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestReadWithOptions_XLSXStructured_InvalidDuplicateHeaders(t *testing.T) {
	sheet := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>name</t></is></c><c r="B1" t="inlineStr"><is><t>name</t></is></c></row><row r="2"><c r="A2" t="inlineStr"><is><t>Alice</t></is></c><c r="B2" t="inlineStr"><is><t>Bob</t></is></c></row></sheetData></worksheet>`
	archive, err := buildXLSXArchive([]xlsxEncodedSheet{{name: "Sheet1", xml: sheet}})
	if err != nil {
		t.Fatalf("unexpected archive build error: %v", err)
	}

	_, err = ReadWithOptions(string(archive), "application/xlsx", Object{"structured": true})
	if err == nil {
		t.Fatal("expected error for duplicate header columns")
	}
	if !strings.Contains(err.Error(), "duplicate column") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestFormatWithOptions_XLSXStructured_ValidatesShape(t *testing.T) {
	t.Run("non object workbook", func(t *testing.T) {
		_, err := FormatWithOptions(Array{}, "application/xlsx", Object{"structured": true})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "expects an object of sheets") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("sheet is not an array", func(t *testing.T) {
		_, err := FormatWithOptions(Object{"Sheet1": Object{}}, "application/xlsx", Object{"structured": true})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "must be an array of row objects") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("row is not an object", func(t *testing.T) {
		_, err := FormatWithOptions(Object{"Sheet1": Array{"bad-row"}}, "application/xlsx", Object{"structured": true})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "row 1 must be an object") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

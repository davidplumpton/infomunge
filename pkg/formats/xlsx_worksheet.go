package formats

import (
	"encoding/xml"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"infomunge/pkg/values"
	"sort"
	"strconv"
	"strings"
)

// XML types for worksheet-level structures.

type xlsxWorksheetXML struct {
	SheetData xlsxSheetDataXML `xml:"sheetData"`
}

type xlsxSheetDataXML struct {
	Rows []xlsxRowXML `xml:"row"`
}

type xlsxRowXML struct {
	Cells []xlsxCellXML `xml:"c"`
}

// xlsxEncodedSheet pairs a sheet name with its rendered worksheet XML,
// used as an intermediate during archive assembly.
type xlsxEncodedSheet struct {
	name string
	xml  string
}

// decodeWorksheetRows parses a raw worksheet XML into an Array of row Objects,
// using the first row as headers.
func decodeWorksheetRows(sheetName string, data []byte, sharedStrings []string, options xlsxFormatOptions) (Array, error) {
	var worksheet xlsxWorksheetXML
	if err := xml.Unmarshal(data, &worksheet); err != nil {
		return nil, unifiederrors.ValidationErrorf("invalid xlsx worksheet for sheet %q: %v", sheetName, err)
	}

	if len(worksheet.SheetData.Rows) == 0 {
		if options.strict {
			return nil, unifiederrors.ValidationErrorf("xlsx sheet %q has no header row", sheetName)
		}
		return Array{}, nil
	}

	headerValues, err := decodeRowValues(worksheet.SheetData.Rows[0], sharedStrings, options)
	if err != nil {
		return nil, unifiederrors.ValidationErrorf("xlsx sheet %q: %v", sheetName, err)
	}
	headers, err := valuesToHeaders(headerValues)
	if err != nil {
		return nil, unifiederrors.ValidationErrorf("xlsx sheet %q: %v", sheetName, err)
	}
	if len(headers) == 0 {
		if options.strict {
			return nil, unifiederrors.ValidationErrorf("xlsx sheet %q has no header row", sheetName)
		}
		return Array{}, nil
	}

	rows := Array{}
	for idx, row := range worksheet.SheetData.Rows {
		if idx == 0 {
			continue
		}
		rowValues, err := decodeRowValues(row, sharedStrings, options)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("xlsx sheet %q: %v", sheetName, err)
		}
		if len(rowValues) > len(headers) {
			if options.strict {
				return nil, unifiederrors.ValidationErrorf("xlsx sheet %q row has %d columns but header defines %d", sheetName, len(rowValues), len(headers))
			}
			rowValues = rowValues[:len(headers)]
		}
		rowObj := values.NewObject(len(headers))
		nonNil := false
		for i, header := range headers {
			if i < len(rowValues) {
				values.SetObjectValue(rowObj, header, rowValues[i])
				if rowValues[i] != nil {
					nonNil = true
				}
			} else {
				values.SetObjectValue(rowObj, header, nil)
			}
		}
		if nonNil {
			rows = append(rows, rowObj)
		}
	}

	return rows, nil
}

// decodeRowValues extracts cell values from a single XML row, filling gaps
// for skipped columns based on cell references.
func decodeRowValues(row xlsxRowXML, sharedStrings []string, options xlsxFormatOptions) ([]interface{}, error) {
	values := []interface{}{}
	for idx, cell := range row.Cells {
		columnIndex := cellColumnIndex(cell.Ref, idx)
		for len(values) < columnIndex {
			values = append(values, nil)
		}
		value, err := decodeCellValue(cell, sharedStrings, options)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// valuesToHeaders converts the first row's values into a list of header
// strings, validating that all entries are non-empty and unique.
func valuesToHeaders(values []interface{}) ([]string, error) {
	headers := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, raw := range values {
		if raw == nil {
			return nil, unifiederrors.ValidationError("header row contains empty column names")
		}
		header := strings.TrimSpace(fmt.Sprint(raw))
		if header == "" {
			return nil, unifiederrors.ValidationError("header row contains empty column names")
		}
		if _, ok := seen[header]; ok {
			return nil, unifiederrors.ValidationErrorf("header row contains duplicate column %q", header)
		}
		seen[header] = struct{}{}
		headers = append(headers, header)
	}

	return headers, nil
}

// encodeWorksheetRows builds worksheet XML from an Array of row Objects,
// collecting column headers from all rows and sorting them alphabetically.
func encodeWorksheetRows(sheetName string, rows Array) (string, error) {
	headersSet := map[string]struct{}{}
	for idx, raw := range rows {
		obj, ok := raw.(Object)
		if !ok {
			return "", unifiederrors.ValidationErrorf("xlsx sheet %q row %d must be an object, got %T", sheetName, idx+1, raw)
		}
		for key := range obj {
			trimmed := strings.TrimSpace(key)
			if trimmed == "" {
				return "", unifiederrors.ValidationErrorf("xlsx sheet %q row %d contains an empty column name", sheetName, idx+1)
			}
			headersSet[key] = struct{}{}
		}
	}

	headers := make([]string, 0, len(headersSet))
	for key := range headersSet {
		headers = append(headers, key)
	}
	sort.Strings(headers)
	if len(rows) > 0 && len(headers) == 0 {
		return "", unifiederrors.ValidationErrorf("xlsx sheet %q rows must contain at least one column", sheetName)
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)

	if len(headers) > 0 {
		b.WriteString(`<row r="1">`)
		for col, header := range headers {
			writeInlineStringCell(&b, 1, col+1, header)
		}
		b.WriteString(`</row>`)
	}

	for rowIdx, raw := range rows {
		obj := raw.(Object)
		rowNumber := rowIdx + 2
		if len(headers) == 0 {
			continue
		}
		var rowBuilder strings.Builder
		hasValue := false
		for col, header := range headers {
			value, exists := obj[header]
			if !exists || value == nil {
				continue
			}
			cellXML, err := encodeCellXML(rowNumber, col+1, value)
			if err != nil {
				return "", unifiederrors.ValidationErrorf("xlsx sheet %q row %d column %q: %v", sheetName, rowIdx+1, header, err)
			}
			rowBuilder.WriteString(cellXML)
			hasValue = true
		}
		if !hasValue {
			continue
		}
		b.WriteString(`<row r="`)
		b.WriteString(strconv.Itoa(rowNumber))
		b.WriteString(`">`)
		b.WriteString(rowBuilder.String())
		b.WriteString(`</row>`)
	}

	b.WriteString(`</sheetData></worksheet>`)
	return b.String(), nil
}

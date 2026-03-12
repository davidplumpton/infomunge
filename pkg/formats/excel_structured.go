package formats

import (
	"archive/zip"
	unifiederrors "infomunge/internal/errors"
	"sort"
	"strings"
)

// xlsxFormatOptions controls structured vs binary XLSX handling and
// strictness of header/cell validation.
type xlsxFormatOptions struct {
	structured bool
	strict     bool
}

func readXLSXWithOptions(content string, options Object) (interface{}, error) {
	parsed, err := parseXLSXOptions(options)
	if err != nil {
		return nil, err
	}
	if !parsed.structured {
		return readBinary(content)
	}

	return decodeStructuredXLSX(content, parsed)
}

func formatXLSXWithOptions(result interface{}, options Object) (string, error) {
	parsed, err := parseXLSXOptions(options)
	if err != nil {
		return "", err
	}
	if !parsed.structured {
		return formatBinary(result)
	}

	return encodeStructuredXLSX(result)
}

func parseXLSXOptions(options Object) (xlsxFormatOptions, error) {
	parsed := xlsxFormatOptions{strict: true}
	if options == nil {
		return parsed, nil
	}

	for key, raw := range options {
		switch key {
		case "structured":
			v, ok := raw.(bool)
			if !ok {
				return xlsxFormatOptions{}, unifiederrors.ValidationErrorf("xlsx option structured must be a boolean, got %T", raw)
			}
			parsed.structured = v
		case "strict":
			v, ok := raw.(bool)
			if !ok {
				return xlsxFormatOptions{}, unifiederrors.ValidationErrorf("xlsx option strict must be a boolean, got %T", raw)
			}
			parsed.strict = v
		default:
			return xlsxFormatOptions{}, unifiederrors.ValidationErrorf("unsupported xlsx option: %s", key)
		}
	}

	if !parsed.structured && len(options) > 0 {
		if len(options) > 1 || options["structured"] != false {
			return xlsxFormatOptions{}, unifiederrors.ValidationError("xlsx options require structured=true")
		}
	}

	return parsed, nil
}

func decodeStructuredXLSX(content string, options xlsxFormatOptions) (interface{}, error) {
	reader, err := zip.NewReader(strings.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, unifiederrors.ValidationErrorf("invalid xlsx payload: %v", err)
	}

	files, err := readZipFiles(reader)
	if err != nil {
		return nil, err
	}

	meta, err := parseWorkbookMetadata(files)
	if err != nil {
		return nil, err
	}
	if len(meta.sheets) == 0 {
		return nil, unifiederrors.ValidationError("xlsx workbook has no sheets")
	}

	sharedStrings, err := parseSharedStrings(files)
	if err != nil {
		return nil, err
	}

	result := Object{}
	for _, sheet := range meta.sheets {
		sheetXML, ok := files[sheet.path]
		if !ok {
			return nil, unifiederrors.ValidationErrorf("xlsx sheet %q is missing worksheet file %s", sheet.name, sheet.path)
		}
		rows, err := decodeWorksheetRows(sheet.name, sheetXML, sharedStrings, options)
		if err != nil {
			return nil, err
		}
		result[sheet.name] = rows
	}

	return result, nil
}

func encodeStructuredXLSX(result interface{}) (string, error) {
	workbook, ok := result.(Object)
	if !ok {
		return "", unifiederrors.ValidationErrorf("xlsx structured output expects an object of sheets, got %T", result)
	}
	if len(workbook) == 0 {
		return "", unifiederrors.ValidationError("xlsx structured output requires at least one sheet")
	}

	sheetNames := make([]string, 0, len(workbook))
	for name := range workbook {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return "", unifiederrors.ValidationError("xlsx sheet names must be non-empty")
		}
		sheetNames = append(sheetNames, name)
	}
	sort.Strings(sheetNames)

	sheets := make([]xlsxEncodedSheet, 0, len(sheetNames))
	for _, sheetName := range sheetNames {
		rawRows := workbook[sheetName]
		rows, ok := rawRows.(Array)
		if !ok {
			return "", unifiederrors.ValidationErrorf("xlsx sheet %q must be an array of row objects, got %T", sheetName, rawRows)
		}
		sheetXML, err := encodeWorksheetRows(sheetName, rows)
		if err != nil {
			return "", err
		}
		sheets = append(sheets, xlsxEncodedSheet{name: sheetName, xml: sheetXML})
	}

	bytesOut, err := buildXLSXArchive(sheets)
	if err != nil {
		return "", err
	}
	return string(bytesOut), nil
}

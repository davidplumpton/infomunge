package formats

import (
	"bytes"
	"encoding/csv"
	unifiederrors "infomunge/internal/errors"
	"io"
	"sort"
	"strings"
)

const MaxCSVRows = 1000000 // Maximum number of rows in CSV to prevent memory exhaustion

func init() {
	RegisterReader("application/csv", readCSV)
	RegisterWriter("application/csv", formatCSV)
	RegisterExtension(".csv", "application/csv")
}

func readCSV(content string) (interface{}, error) {
	r := csv.NewReader(strings.NewReader(content))

	// Read header first
	header, err := r.Read()
	if err != nil {
		if err == io.EOF {
			return Array{}, nil
		}
		return nil, unifiederrors.WrapValidationf(err, "CSV parse error: %v", err)
	}

	headerLen := len(header)

	// Validate header: check for empty column names
	for i, col := range header {
		if strings.TrimSpace(col) == "" {
			return nil, unifiederrors.ValidationErrorf("CSV header has empty column name at position %d", i+1)
		}
	}

	var result Array
	rowNum := 1 // Start from 1 since we already read the header

	// Read rows one at a time to prevent memory exhaustion
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, unifiederrors.WrapValidationf(err, "CSV parse error: %v", err)
		}

		rowNum++
		if rowNum > MaxCSVRows {
			return nil, unifiederrors.ValidationErrorf("CSV row limit exceeded (max %d rows)", MaxCSVRows)
		}

		// Validate row length matches header
		if len(record) != headerLen {
			return nil, unifiederrors.ValidationErrorf("CSV row %d has %d columns, expected %d (matching header)", rowNum, len(record), headerLen)
		}

		obj := make(Object)
		for i, val := range record {
			obj[header[i]] = val
		}
		result = append(result, obj)
	}
	return result, nil
}

func formatCSV(result interface{}) (string, error) {
	items, ok := result.(Array)
	if !ok {
		return "", unifiederrors.ValidationErrorf("CSV output expects an array of objects, got %T", result)
	}

	if len(items) == 0 {
		return "", nil
	}

	// Collect headers from ALL items to handle heterogeneous data
	headerSet := make(map[string]struct{})
	for i, item := range items {
		m, ok := item.(Object)
		if !ok {
			return "", unifiederrors.ValidationErrorf("CSV output expects array elements to be objects, item %d is %T", i+1, item)
		}
		for k := range m {
			headerSet[k] = struct{}{}
		}
	}

	if len(headerSet) == 0 {
		return "", unifiederrors.ValidationError("CSV output expects at least one object with fields")
	}

	headers := make([]string, 0, len(headerSet))
	for k := range headerSet {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return "", unifiederrors.WrapValidationf(err, "CSV format error: %v", err)
	}

	for _, item := range items {
		m := item.(Object)
		record := make([]string, 0, len(headers))
		for _, h := range headers {
			if val, exists := m[h]; exists {
				record = append(record, valueToCSVString(val))
			} else {
				record = append(record, "")
			}
		}
		if err := w.Write(record); err != nil {
			return "", unifiederrors.WrapValidationf(err, "CSV format error: %v", err)
		}
	}
	w.Flush()
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

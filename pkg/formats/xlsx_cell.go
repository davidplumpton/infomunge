package formats

import (
	"bytes"
	"encoding/xml"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"strconv"
	"strings"
)

// XML types for cell-level structures.

type xlsxCellXML struct {
	Ref      string          `xml:"r,attr"`
	Type     string          `xml:"t,attr"`
	Value    string          `xml:"v"`
	Inline   *xlsxInlineXML  `xml:"is"`
	RichText []xlsxRunXML    `xml:"r"`
	Formula  *xlsxFormulaXML `xml:"f"`
}

type xlsxInlineXML struct {
	Text string       `xml:"t"`
	Runs []xlsxRunXML `xml:"r"`
}

type xlsxRunXML struct {
	Text string `xml:"t"`
}

type xlsxFormulaXML struct {
	Expr string `xml:",chardata"`
}

// decodeCellValue converts a raw XML cell into a Go value using the cell's
// type attribute and the workbook's shared-string table.
func decodeCellValue(cell xlsxCellXML, sharedStrings []string, options xlsxFormatOptions) (interface{}, error) {
	cellType := strings.TrimSpace(cell.Type)
	switch cellType {
	case "", "n":
		trimmed := strings.TrimSpace(cell.Value)
		if trimmed == "" {
			return nil, nil
		}
		n, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			if options.strict {
				return nil, unifiederrors.ValidationErrorf("invalid numeric cell value %q", trimmed)
			}
			return trimmed, nil
		}
		return n, nil
	case "s":
		idx, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil || idx < 0 || idx >= len(sharedStrings) {
			return nil, unifiederrors.ValidationErrorf("invalid shared string index %q", strings.TrimSpace(cell.Value))
		}
		return sharedStrings[idx], nil
	case "inlineStr":
		if cell.Inline == nil {
			return "", nil
		}
		if cell.Inline.Text != "" {
			return cell.Inline.Text, nil
		}
		var builder strings.Builder
		for _, run := range cell.Inline.Runs {
			builder.WriteString(run.Text)
		}
		return builder.String(), nil
	case "str":
		if strings.TrimSpace(cell.Value) != "" {
			return cell.Value, nil
		}
		if cell.Inline != nil {
			return cell.Inline.Text, nil
		}
		return "", nil
	case "b":
		trimmed := strings.TrimSpace(cell.Value)
		if trimmed == "1" {
			return true, nil
		}
		if trimmed == "0" {
			return false, nil
		}
		if options.strict {
			return nil, unifiederrors.ValidationErrorf("invalid boolean cell value %q", trimmed)
		}
		return strings.EqualFold(trimmed, "true"), nil
	case "d":
		if strings.TrimSpace(cell.Value) == "" {
			return nil, nil
		}
		return cell.Value, nil
	case "e":
		if options.strict {
			return nil, unifiederrors.ValidationError("formula/error cell type is not supported in structured xlsx mode")
		}
		return nil, nil
	default:
		if options.strict {
			return nil, unifiederrors.ValidationErrorf("unsupported xlsx cell type %q", cellType)
		}
		return strings.TrimSpace(cell.Value), nil
	}
}

// encodeCellXML renders a single cell value as an OOXML <c> element.
func encodeCellXML(row, col int, value interface{}) (string, error) {
	ref := encodeCellRef(row, col)
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>%s</t></is></c>`, ref, xmlEscape(v)), nil
	case bool:
		if v {
			return fmt.Sprintf(`<c r="%s" t="b"><v>1</v></c>`, ref), nil
		}
		return fmt.Sprintf(`<c r="%s" t="b"><v>0</v></c>`, ref), nil
	case int:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case int8:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case int16:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case int32:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case int64:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case uint:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case uint8:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case uint16:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case uint32:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case uint64:
		return fmt.Sprintf(`<c r="%s"><v>%d</v></c>`, ref, v), nil
	case float32:
		return fmt.Sprintf(`<c r="%s"><v>%s</v></c>`, ref, strconv.FormatFloat(float64(v), 'f', -1, 64)), nil
	case float64:
		return fmt.Sprintf(`<c r="%s"><v>%s</v></c>`, ref, strconv.FormatFloat(v, 'f', -1, 64)), nil
	default:
		return "", unifiederrors.ValidationErrorf("unsupported value type %T", value)
	}
}

// writeInlineStringCell writes an inline-string cell element to b.
func writeInlineStringCell(b *strings.Builder, row, col int, value string) {
	b.WriteString(`<c r="`)
	b.WriteString(encodeCellRef(row, col))
	b.WriteString(`" t="inlineStr"><is><t>`)
	b.WriteString(xmlEscape(value))
	b.WriteString(`</t></is></c>`)
}

// cellColumnIndex returns the zero-based column index from a cell reference
// like "B3". If the reference is empty or unparseable, fallback is returned.
func cellColumnIndex(ref string, fallback int) int {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return fallback
	}
	letters := make([]rune, 0, len(ref))
	for _, r := range ref {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			letters = append(letters, r)
			continue
		}
		break
	}
	if len(letters) == 0 {
		return fallback
	}
	column := 0
	for _, r := range letters {
		r = rune(strings.ToUpper(string(r))[0])
		column = column*26 + int(r-'A'+1)
	}
	if column <= 0 {
		return fallback
	}
	return column - 1
}

// encodeCellRef returns a cell reference like "A1" from a 1-based row and column.
func encodeCellRef(row, col int) string {
	if row < 1 {
		row = 1
	}
	if col < 1 {
		col = 1
	}
	return columnName(col) + strconv.Itoa(row)
}

// columnName returns the column letter(s) for a 1-based column number
// (1 → "A", 27 → "AA", etc.).
func columnName(col int) string {
	if col < 1 {
		return "A"
	}
	out := ""
	for col > 0 {
		col--
		out = string(rune('A'+(col%26))) + out
		col /= 26
	}
	return out
}

// xmlEscape escapes a string for safe inclusion in XML text or quoted attribute values.
func xmlEscape(value string) string {
	var b bytes.Buffer
	if err := xml.EscapeText(&b, []byte(value)); err != nil {
		return value
	}
	return b.String()
}

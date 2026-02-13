package formats

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	xlsxWorkbookPath     = "xl/workbook.xml"
	xlsxWorkbookRelsPath = "xl/_rels/workbook.xml.rels"
	xlsxSheetPrefix      = "xl/worksheets/sheet"
)

type xlsxFormatOptions struct {
	structured bool
	strict     bool
}

type xlsxWorkbookMeta struct {
	sheets []xlsxSheetMeta
}

type xlsxEncodedSheet struct {
	name string
	xml  string
}

type xlsxSheetMeta struct {
	name string
	path string
}

type xlsxWorkbookXML struct {
	Sheets xlsxWorkbookSheetsXML `xml:"sheets"`
}

type xlsxWorkbookSheetsXML struct {
	Entries []xlsxWorkbookSheetEntryXML `xml:"sheet"`
}

type xlsxWorkbookSheetEntryXML struct {
	Name string `xml:"name,attr"`
	RID  string `xml:"id,attr"`
}

type xlsxWorkbookRelsXML struct {
	Relationships []xlsxRelationshipXML `xml:"Relationship"`
}

type xlsxRelationshipXML struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type xlsxWorksheetXML struct {
	SheetData xlsxSheetDataXML `xml:"sheetData"`
}

type xlsxSheetDataXML struct {
	Rows []xlsxRowXML `xml:"row"`
}

type xlsxRowXML struct {
	Cells []xlsxCellXML `xml:"c"`
}

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

type xlsxSharedStringsXML struct {
	Items []xlsxSharedStringItemXML `xml:"si"`
}

type xlsxSharedStringItemXML struct {
	Text string       `xml:"t"`
	Runs []xlsxRunXML `xml:"r"`
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

func readZipFiles(reader *zip.Reader) (map[string][]byte, error) {
	files := make(map[string][]byte, len(reader.File))
	for _, file := range reader.File {
		h, err := file.Open()
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("unable to read xlsx file %s: %v", file.Name, err)
		}
		data, err := io.ReadAll(h)
		_ = h.Close()
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("unable to read xlsx file %s: %v", file.Name, err)
		}
		files[file.Name] = data
	}
	return files, nil
}

func parseWorkbookMetadata(files map[string][]byte) (xlsxWorkbookMeta, error) {
	workbookXML, ok := files[xlsxWorkbookPath]
	if !ok {
		return xlsxWorkbookMeta{}, unifiederrors.ValidationError("xlsx workbook is missing xl/workbook.xml")
	}
	relsXML, ok := files[xlsxWorkbookRelsPath]
	if !ok {
		return xlsxWorkbookMeta{}, unifiederrors.ValidationError("xlsx workbook is missing xl/_rels/workbook.xml.rels")
	}

	var workbook xlsxWorkbookXML
	if err := xml.Unmarshal(workbookXML, &workbook); err != nil {
		return xlsxWorkbookMeta{}, unifiederrors.ValidationErrorf("invalid xlsx workbook.xml: %v", err)
	}
	var rels xlsxWorkbookRelsXML
	if err := xml.Unmarshal(relsXML, &rels); err != nil {
		return xlsxWorkbookMeta{}, unifiederrors.ValidationErrorf("invalid xlsx workbook.xml.rels: %v", err)
	}

	relsByID := map[string]string{}
	for _, rel := range rels.Relationships {
		if rel.ID == "" || rel.Target == "" {
			continue
		}
		relsByID[rel.ID] = normalizeWorkbookTarget(rel.Target)
	}

	meta := xlsxWorkbookMeta{sheets: make([]xlsxSheetMeta, 0, len(workbook.Sheets.Entries))}
	for _, sheet := range workbook.Sheets.Entries {
		if strings.TrimSpace(sheet.Name) == "" {
			return xlsxWorkbookMeta{}, unifiederrors.ValidationError("xlsx workbook contains a sheet with an empty name")
		}
		sheetPath, ok := relsByID[sheet.RID]
		if !ok {
			return xlsxWorkbookMeta{}, unifiederrors.ValidationErrorf("xlsx workbook sheet %q has unresolved relationship %q", sheet.Name, sheet.RID)
		}
		meta.sheets = append(meta.sheets, xlsxSheetMeta{name: sheet.Name, path: sheetPath})
	}

	return meta, nil
}

func normalizeWorkbookTarget(target string) string {
	clean := strings.TrimSpace(strings.ReplaceAll(target, "\\", "/"))
	if clean == "" {
		return ""
	}
	clean = path.Clean(clean)
	if strings.HasPrefix(clean, "../") {
		clean = strings.TrimPrefix(clean, "../")
	}
	if strings.HasPrefix(clean, "xl/") {
		return clean
	}
	return path.Join("xl", clean)
}

func parseSharedStrings(files map[string][]byte) ([]string, error) {
	data, ok := files["xl/sharedStrings.xml"]
	if !ok {
		return nil, nil
	}
	var doc xlsxSharedStringsXML
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, unifiederrors.ValidationErrorf("invalid xlsx sharedStrings.xml: %v", err)
	}
	out := make([]string, 0, len(doc.Items))
	for _, item := range doc.Items {
		if item.Text != "" {
			out = append(out, item.Text)
			continue
		}
		var builder strings.Builder
		for _, run := range item.Runs {
			builder.WriteString(run.Text)
		}
		out = append(out, builder.String())
	}
	return out, nil
}

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
		values, err := decodeRowValues(row, sharedStrings, options)
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("xlsx sheet %q: %v", sheetName, err)
		}
		if len(values) > len(headers) {
			if options.strict {
				return nil, unifiederrors.ValidationErrorf("xlsx sheet %q row has %d columns but header defines %d", sheetName, len(values), len(headers))
			}
			values = values[:len(headers)]
		}
		rowObj := Object{}
		nonNil := false
		for i, header := range headers {
			if i < len(values) {
				rowObj[header] = values[i]
				if values[i] != nil {
					nonNil = true
				}
			} else {
				rowObj[header] = nil
			}
		}
		if nonNil {
			rows = append(rows, rowObj)
		}
	}

	return rows, nil
}

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

func writeInlineStringCell(b *strings.Builder, row, col int, value string) {
	b.WriteString(`<c r="`)
	b.WriteString(encodeCellRef(row, col))
	b.WriteString(`" t="inlineStr"><is><t>`)
	b.WriteString(xmlEscape(value))
	b.WriteString(`</t></is></c>`)
}

func buildXLSXArchive(sheets []xlsxEncodedSheet) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	write := func(name, content string) error {
		writer, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = writer.Write([]byte(content))
		return err
	}

	if err := write("[Content_Types].xml", buildContentTypesXML(len(sheets))); err != nil {
		_ = zw.Close()
		return nil, unifiederrors.ValidationErrorf("unable to build xlsx archive: %v", err)
	}
	if err := write("_rels/.rels", rootRelationshipsXML); err != nil {
		_ = zw.Close()
		return nil, unifiederrors.ValidationErrorf("unable to build xlsx archive: %v", err)
	}
	if err := write(xlsxWorkbookPath, buildWorkbookXML(sheets)); err != nil {
		_ = zw.Close()
		return nil, unifiederrors.ValidationErrorf("unable to build xlsx archive: %v", err)
	}
	if err := write(xlsxWorkbookRelsPath, buildWorkbookRelsXML(len(sheets))); err != nil {
		_ = zw.Close()
		return nil, unifiederrors.ValidationErrorf("unable to build xlsx archive: %v", err)
	}
	for idx, sheet := range sheets {
		name := fmt.Sprintf("%s%d.xml", xlsxSheetPrefix, idx+1)
		if err := write(name, sheet.xml); err != nil {
			_ = zw.Close()
			return nil, unifiederrors.ValidationErrorf("unable to build xlsx archive: %v", err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, unifiederrors.ValidationErrorf("unable to finalize xlsx archive: %v", err)
	}

	return buf.Bytes(), nil
}

func buildContentTypesXML(sheetCount int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	b.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	for i := 1; i <= sheetCount; i++ {
		b.WriteString(`<Override PartName="/xl/worksheets/sheet`)
		b.WriteString(strconv.Itoa(i))
		b.WriteString(`.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`)
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func buildWorkbookXML(sheets []xlsxEncodedSheet) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for i, sheet := range sheets {
		b.WriteString(`<sheet name="`)
		b.WriteString(xmlEscape(sheet.name))
		b.WriteString(`" sheetId="`)
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(`" r:id="rId`)
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteString(`"/>`)
	}
	b.WriteString(`</sheets></workbook>`)
	return b.String()
}

func buildWorkbookRelsXML(sheetCount int) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := 1; i <= sheetCount; i++ {
		b.WriteString(`<Relationship Id="rId`)
		b.WriteString(strconv.Itoa(i))
		b.WriteString(`" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet`)
		b.WriteString(strconv.Itoa(i))
		b.WriteString(`.xml"/>`)
	}
	b.WriteString(`</Relationships>`)
	return b.String()
}

const rootRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`

func xmlEscape(value string) string {
	var b bytes.Buffer
	if err := xml.EscapeText(&b, []byte(value)); err != nil {
		return value
	}
	return b.String()
}

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

func encodeCellRef(row, col int) string {
	if row < 1 {
		row = 1
	}
	if col < 1 {
		col = 1
	}
	return columnName(col) + strconv.Itoa(row)
}

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

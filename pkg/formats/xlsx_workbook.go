package formats

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	unifiederrors "infomunge/internal/errors"
	"infomunge/internal/readlimit"
	"path"
	"strconv"
	"strings"
)

// Constants for well-known paths inside an XLSX archive.
const (
	xlsxWorkbookPath     = "xl/workbook.xml"
	xlsxWorkbookRelsPath = "xl/_rels/workbook.xml.rels"
	xlsxSheetPrefix      = "xl/worksheets/sheet"
	MaxXLSXZipEntries    = 512
	MaxXLSXZipEntryBytes = 2 * 1024 * 1024
	MaxXLSXZipTotalBytes = 10 * 1024 * 1024
)

// Workbook-level model types.

type xlsxWorkbookMeta struct {
	sheets []xlsxSheetMeta
}

type xlsxSheetMeta struct {
	name string
	path string
}

// XML types for workbook and relationship documents.

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

// XML types for shared strings.

type xlsxSharedStringsXML struct {
	Items []xlsxSharedStringItemXML `xml:"si"`
}

type xlsxSharedStringItemXML struct {
	Text string       `xml:"t"`
	Runs []xlsxRunXML `xml:"r"`
}

// readZipFiles reads all entries from a zip.Reader into an in-memory map.
func readZipFiles(reader *zip.Reader) (map[string][]byte, error) {
	if len(reader.File) > MaxXLSXZipEntries {
		return nil, unifiederrors.ValidationErrorf("xlsx archive contains too many files (max %d)", MaxXLSXZipEntries)
	}

	files := make(map[string][]byte, len(reader.File))
	totalBytes := int64(0)
	for _, file := range reader.File {
		h, err := file.Open()
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("unable to read xlsx file %s: %v", file.Name, err)
		}
		data, tooLarge, err := readlimit.ReadAll(h, MaxXLSXZipEntryBytes)
		_ = h.Close()
		if err != nil {
			return nil, unifiederrors.ValidationErrorf("unable to read xlsx file %s: %v", file.Name, err)
		}
		if tooLarge {
			return nil, unifiederrors.ValidationErrorf("xlsx file %s exceeds maximum size of %d bytes", file.Name, MaxXLSXZipEntryBytes)
		}
		totalBytes += int64(len(data))
		if totalBytes > MaxXLSXZipTotalBytes {
			return nil, unifiederrors.ValidationErrorf("xlsx archive decompressed size exceeds maximum of %d bytes", MaxXLSXZipTotalBytes)
		}
		files[file.Name] = data
	}
	return files, nil
}

// parseWorkbookMetadata extracts sheet names and resolved paths from the
// workbook and relationship XML files.
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

// normalizeWorkbookTarget cleans a relationship target path so it is always
// rooted under "xl/" regardless of how it was expressed in the rels file.
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

// parseSharedStrings decodes the shared-string table from the archive,
// returning nil if the file is absent.
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

// buildXLSXArchive assembles a complete XLSX ZIP archive from a list of
// encoded sheets, including content types, relationships, and workbook files.
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

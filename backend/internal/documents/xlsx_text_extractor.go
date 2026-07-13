package documents

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type xlsxSheetDefinition struct {
	Name    string
	RelID   string
	ZipPath string
}

func extractXLSXText(filePath string) (string, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	zipEntries := indexZipEntries(reader.File)
	sharedStrings, err := readXLSSharedStrings(zipEntries)
	if err != nil {
		return "", err
	}

	sheetTargets, err := readXLSXWorkbookRelationships(zipEntries)
	if err != nil {
		return "", err
	}

	sheets, err := readXLSXSheetDefinitions(zipEntries, sheetTargets)
	if err != nil {
		return "", err
	}
	if len(sheets) == 0 {
		return "", nil
	}

	var builder strings.Builder
	for index, sheet := range sheets {
		sheetText, err := readXLSXWorksheetText(zipEntries, sheet.ZipPath, sharedStrings)
		if err != nil {
			return "", fmt.Errorf("%s: %w", sheet.ZipPath, err)
		}

		if index > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(formatXLSXSheetText(sheet.Name, sheetText))
	}

	if builder.Len() == 0 {
		return "", nil
	}

	return cleanXLSXText(builder.String()), nil
}

func indexZipEntries(files []*zip.File) map[string]*zip.File {
	entries := make(map[string]*zip.File, len(files))
	for _, file := range files {
		entries[normalizeXLSXZipEntryName(file.Name)] = file
	}
	return entries
}

func normalizeXLSXZipEntryName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "/")
	return strings.ToLower(name)
}

func readZipEntryText(entries map[string]*zip.File, normalizedPath string) ([]byte, error) {
	file, ok := entries[normalizedPath]
	if !ok {
		return nil, fmt.Errorf("zip entry not found: %s", normalizedPath)
	}

	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func readXLSSharedStrings(entries map[string]*zip.File) ([]string, error) {
	data, err := readZipEntryText(entries, "xl/sharedstrings.xml")
	if err != nil {
		if strings.Contains(err.Error(), "zip entry not found") {
			return nil, nil
		}
		return nil, err
	}

	return parseXLSSharedStrings(bytes.NewReader(data))
}

func parseXLSSharedStrings(reader io.Reader) ([]string, error) {
	decoder := xml.NewDecoder(reader)
	sharedStrings := make([]string, 0)
	var currentItem strings.Builder
	var inText bool
	var depth int

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch value := token.(type) {
		case xml.StartElement:
			switch value.Name.Local {
			case "si":
				currentItem.Reset()
				depth = 0
			case "t":
				inText = true
				depth++
			}
		case xml.CharData:
			if inText {
				currentItem.Write(value)
			}
		case xml.EndElement:
			switch value.Name.Local {
			case "t":
				depth--
				if depth <= 0 {
					inText = false
					depth = 0
				}
			case "si":
				sharedStrings = append(sharedStrings, currentItem.String())
			}
		}
	}

	return sharedStrings, nil
}

func readXLSXWorkbookRelationships(entries map[string]*zip.File) (map[string]string, error) {
	data, err := readZipEntryText(entries, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	relationships := make(map[string]string)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		startElement, ok := token.(xml.StartElement)
		if !ok || startElement.Name.Local != "Relationship" {
			continue
		}

		var relationshipID string
		var target string
		for _, attr := range startElement.Attr {
			switch attr.Name.Local {
			case "Id":
				relationshipID = attr.Value
			case "Target":
				target = attr.Value
			}
		}

		if relationshipID == "" || target == "" {
			continue
		}

		relationships[relationshipID] = normalizeXLSXRelationshipTarget(target)
	}

	return relationships, nil
}

func normalizeXLSXRelationshipTarget(target string) string {
	target = strings.ReplaceAll(target, "\\", "/")
	target = strings.TrimPrefix(target, "/")
	if strings.HasPrefix(target, "xl/") {
		return normalizeXLSXZipEntryName(target)
	}
	return normalizeXLSXZipEntryName(path.Join("xl", target))
}

func readXLSXSheetDefinitions(entries map[string]*zip.File, relationships map[string]string) ([]xlsxSheetDefinition, error) {
	data, err := readZipEntryText(entries, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	sheets := make([]xlsxSheetDefinition, 0)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		startElement, ok := token.(xml.StartElement)
		if !ok || startElement.Name.Local != "sheet" {
			continue
		}

		sheet := xlsxSheetDefinition{}
		for _, attr := range startElement.Attr {
			switch attr.Name.Local {
			case "name":
				sheet.Name = attr.Value
			case "id":
				sheet.RelID = attr.Value
			}
		}

		if sheet.Name == "" || sheet.RelID == "" {
			continue
		}

		zipPath, ok := relationships[sheet.RelID]
		if !ok {
			continue
		}
		sheet.ZipPath = zipPath
		sheets = append(sheets, sheet)
	}

	if len(sheets) == 0 {
		return fallbackXLSXSheetDefinitions(entries), nil
	}

	return sheets, nil
}

func fallbackXLSXSheetDefinitions(entries map[string]*zip.File) []xlsxSheetDefinition {
	paths := make([]string, 0)
	for normalizedPath := range entries {
		if strings.HasPrefix(normalizedPath, "xl/worksheets/") && path.Ext(normalizedPath) == ".xml" {
			paths = append(paths, normalizedPath)
		}
	}

	sort.Strings(paths)
	sheets := make([]xlsxSheetDefinition, 0, len(paths))
	for index, zipPath := range paths {
		sheets = append(sheets, xlsxSheetDefinition{
			Name:    fmt.Sprintf("Sheet%d", index+1),
			ZipPath: zipPath,
		})
	}

	return sheets
}

func readXLSXWorksheetText(entries map[string]*zip.File, zipPath string, sharedStrings []string) (string, error) {
	data, err := readZipEntryText(entries, zipPath)
	if err != nil {
		return "", err
	}

	return extractXLSXWorksheetText(bytes.NewReader(data), sharedStrings)
}

func extractXLSXWorksheetText(reader io.Reader, sharedStrings []string) (string, error) {
	decoder := xml.NewDecoder(reader)
	var rows []string
	var currentCells []string
	var cellType string
	var cellValue strings.Builder
	var inValue bool
	var inInlineString bool
	var inInlineText bool

	flushCell := func() {
		value := resolveXLSXCellValue(cellType, cellValue.String(), sharedStrings)
		cellType = ""
		cellValue.Reset()
		if value != "" {
			currentCells = append(currentCells, value)
		}
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch value := token.(type) {
		case xml.StartElement:
			switch value.Name.Local {
			case "row":
				currentCells = nil
			case "c":
				cellType = ""
				cellValue.Reset()
				for _, attr := range value.Attr {
					if attr.Name.Local == "t" {
						cellType = attr.Value
					}
				}
			case "v":
				inValue = true
			case "is":
				inInlineString = true
			case "t":
				if inInlineString {
					inInlineText = true
				}
			}
		case xml.CharData:
			if inValue || inInlineText {
				cellValue.Write(value)
			}
		case xml.EndElement:
			switch value.Name.Local {
			case "v":
				inValue = false
			case "t":
				inInlineText = false
			case "is":
				inInlineString = false
			case "c":
				flushCell()
			case "row":
				if len(currentCells) > 0 {
					rows = append(rows, strings.Join(currentCells, "\t"))
				}
				currentCells = nil
			}
		}
	}

	return strings.Join(rows, "\n"), nil
}

func resolveXLSXCellValue(cellType, rawValue string, sharedStrings []string) string {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return ""
	}

	switch cellType {
	case "s":
		index, err := strconv.Atoi(rawValue)
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return ""
		}
		return strings.TrimSpace(sharedStrings[index])
	case "inlineStr", "str":
		return strings.TrimSpace(rawValue)
	case "b":
		if rawValue == "1" {
			return "TRUE"
		}
		if rawValue == "0" {
			return "FALSE"
		}
		return rawValue
	default:
		return rawValue
	}
}

func formatXLSXSheetText(sheetName, sheetBody string) string {
	sheetName = strings.TrimSpace(sheetName)
	if sheetName == "" {
		sheetName = "Sheet"
	}

	sheetBody = strings.TrimSpace(sheetBody)
	if sheetBody == "" {
		return "[" + sheetName + "]"
	}

	return "[" + sheetName + "]\n" + sheetBody
}

func cleanXLSXText(value string) string {
	value = strings.ReplaceAll(value, "\x00", "")
	lines := strings.Split(value, "\n")
	normalizedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			normalizedLines = append(normalizedLines, trimmedLine)
			continue
		}

		cells := strings.Split(line, "\t")
		normalizedCells := make([]string, 0, len(cells))
		for _, cell := range cells {
			cell = strings.Join(strings.Fields(strings.TrimSpace(cell)), " ")
			if cell != "" {
				normalizedCells = append(normalizedCells, cell)
			}
		}
		if len(normalizedCells) > 0 {
			normalizedLines = append(normalizedLines, strings.Join(normalizedCells, "\t"))
		}
	}

	return strings.TrimSpace(norm.NFC.String(strings.Join(normalizedLines, "\n")))
}

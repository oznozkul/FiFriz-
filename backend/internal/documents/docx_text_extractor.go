package documents

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type docxXMLToken struct {
	XMLName xml.Name
	Text    string `xml:",chardata"`
}

func extractDOCXText(filePath string) (string, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var builder strings.Builder
	foundTextPart := false
	for _, file := range reader.File {
		if !isDOCXTextPart(file.Name) {
			continue
		}

		documentXML, err := file.Open()
		if err != nil {
			return "", err
		}

		text, err := extractDOCXXMLText(documentXML)
		closeErr := documentXML.Close()
		if err != nil {
			return "", fmt.Errorf("%s: %w", file.Name, err)
		}
		if closeErr != nil {
			return "", closeErr
		}

		text = strings.TrimSpace(text)
		if text != "" {
			if builder.Len() > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(text)
			foundTextPart = true
		}
	}

	if !foundTextPart {
		return "", nil
	}

	return cleanDOCXText(builder.String()), nil
}

func normalizeDOCXZipEntryName(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "/")
	return strings.ToLower(name)
}

func isDOCXTextPart(name string) bool {
	name = normalizeDOCXZipEntryName(name)
	if path.Ext(name) != ".xml" {
		return false
	}

	switch name {
	case "word/document.xml", "word/footnotes.xml", "word/endnotes.xml", "word/comments.xml":
		return true
	}

	return strings.HasPrefix(name, "word/header") || strings.HasPrefix(name, "word/footer")
}

func extractDOCXXMLText(reader io.Reader) (string, error) {
	decoder := xml.NewDecoder(reader)
	var builder strings.Builder
	var paragraphHasText bool

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
			case "t", "instrText", "delText":
				var textToken docxXMLToken
				if err := decoder.DecodeElement(&textToken, &value); err != nil {
					return "", err
				}
				if textToken.Text != "" {
					builder.WriteString(textToken.Text)
					paragraphHasText = true
				}
			case "tab":
				builder.WriteString("\t")
			case "br", "cr":
				builder.WriteString("\n")
			}
		case xml.EndElement:
			if value.Name.Local == "p" && paragraphHasText {
				builder.WriteString("\n")
				paragraphHasText = false
			}
		}
	}

	return builder.String(), nil
}

func cleanDOCXText(value string) string {
	text := strings.ReplaceAll(value, "\x00", "")
	lines := strings.Split(text, "\n")
	normalizedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			normalizedLines = append(normalizedLines, line)
		}
	}

	return strings.TrimSpace(norm.NFC.String(strings.Join(normalizedLines, "\n")))
}

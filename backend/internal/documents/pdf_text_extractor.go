package documents

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/ledongthuc/pdf"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/unicode/norm"
)

func extractPDFText(filePath string) (string, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	plainTextReader, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, plainTextReader); err != nil {
		return "", err
	}

	return normalizeExtractedText(buffer.Bytes()), nil
}

func normalizeExtractedText(rawText []byte) string {
	var text string
	if utf8.Valid(rawText) {
		text = string(rawText)
	} else if decoded, err := charmap.Windows1254.NewDecoder().Bytes(rawText); err == nil && utf8.Valid(decoded) {
		text = string(decoded)
	} else if decoded, err := charmap.ISO8859_9.NewDecoder().Bytes(rawText); err == nil && utf8.Valid(decoded) {
		text = string(decoded)
	} else {
		text = strings.ToValidUTF8(string(rawText), "")
	}

	text = strings.ToValidUTF8(text, "")
	text = strings.ReplaceAll(text, "\x00", "")
	text = replaceCommonLigatures(text)
	text = strings.Join(strings.Fields(text), " ")
	text = repairCommonTurkishReplacement(text)
	text = repairFragmentedTurkishReplacement(text)
	text = norm.NFC.String(text)

	return strings.TrimSpace(text)
}

func replaceCommonLigatures(text string) string {
	replacer := strings.NewReplacer(
		"\uFB00", "ff",
		"\uFB01", "fi",
		"\uFB02", "fl",
		"\uFB03", "ffi",
		"\uFB04", "ffl",
	)
	return replacer.Replace(text)
}

func repairCommonTurkishReplacement(text string) string {
	replacer := strings.NewReplacer(
		"K \uFFFD raya", "Kiraya",
		"K\uFFFDraya", "Kiraya",
		"k \uFFFD raya", "kiraya",
		"k\uFFFDraya", "kiraya",
		"K \uFFFD racı", "Kiracı",
		"k \uFFFD racı", "kiracı",
		"K \uFFFD raci", "Kiraci",
		"k \uFFFD raci", "kiraci",
		"K \uFFFD ml \uFFFD k", "Kimlik",
		"k \uFFFD ml \uFFFD k", "kimlik",
		"Sah \uFFFD b \uFFFD", "Sahibi",
		"sah \uFFFD b \uFFFD", "sahibi",
		"m \uFFFD lk", "m\u00fclk",
		"m\uFFFDlk", "m\u00fclk",
		"M \uFFFD lk", "M\u00fclk",
		"M\uFFFDlk", "M\u00fclk",
		"Da \uFFFD re", "Daire",
		"Da\uFFFDre", "Daire",
		"da \uFFFD re", "daire",
		"da\uFFFDre", "daire",
		"s \uFFFD zle \uFFFD me", "s\u00f6zle\u015fme",
		"s\uFFFDzle\uFFFDme", "s\u00f6zle\u015fme",
		"S \uFFFD zle \uFFFD me", "S\u00f6zle\u015fme",
		"S\uFFFDzle\uFFFDme", "S\u00f6zle\u015fme",
		"sozle \uFFFD me", "sozle\u015fme",
		"sozle\uFFFDme", "sozle\u015fme",
		"Sozle \uFFFD me", "Sozle\u015fme",
		"Sozle\uFFFDme", "Sozle\u015fme",
	)

	return replacer.Replace(text)
}

func repairFragmentedTurkishReplacement(text string) string {
	text = strings.ReplaceAll(text, "\uFFFD ", "i ")
	text = strings.ReplaceAll(text, " \uFFFD", " i")
	text = strings.ReplaceAll(text, "\uFFFD", "\u0131")

	fragmentPattern := regexp.MustCompile(`(\p{L})\s+([iı])\s+(\p{L})`)
	for {
		repaired := fragmentPattern.ReplaceAllString(text, "$1$2$3")
		if repaired == text {
			break
		}
		text = repaired
	}

	replacer := strings.NewReplacer(
		"havalesii le", "havalesi ile",
		"iile", "i ile",
		"\u0131ile", "\u0131 ile",
		"verdi\u011finibir", "verdi\u011fini bir",
		"yaz\u0131li", "yaz\u0131l\u0131",
		"belgeibraz", "belge ibraz",
		"Bilirki\u015firaporu", "Bilirki\u015fi raporu",
		"bilirki\u015firaporu", "bilirki\u015fi raporu",
		"kay\u0131tlari", "kay\u0131tlar\u0131",
		"diiital", "dijital",
		"Diiital", "Dijital",
		"Tara\u0131ar", "Taraflar",
		"tara\u0131ar", "taraflar",
		"yaz\u0131l\u0131", "yaz\u0131l\u0131",
		"a\u00e7\u0131k", "a\u00e7\u0131k",
	)

	return replacer.Replace(text)
}

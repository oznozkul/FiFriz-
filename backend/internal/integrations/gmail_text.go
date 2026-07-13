package integrations

import (
	"regexp"
	"strings"
)

var htmlTagPattern = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
var htmlElementPattern = regexp.MustCompile(`(?is)<[^>]+>`)

func htmlToPlainText(html string) string {
	html = strings.TrimSpace(html)
	if html == "" {
		return ""
	}
	html = htmlTagPattern.ReplaceAllString(html, " ")
	html = strings.ReplaceAll(html, "<br>", "\n")
	html = strings.ReplaceAll(html, "<br/>", "\n")
	html = strings.ReplaceAll(html, "<br />", "\n")
	html = strings.ReplaceAll(html, "</p>", "\n")
	html = strings.ReplaceAll(html, "</div>", "\n")
	html = htmlElementPattern.ReplaceAllString(html, " ")
	html = strings.Join(strings.Fields(html), " ")
	return strings.TrimSpace(html)
}

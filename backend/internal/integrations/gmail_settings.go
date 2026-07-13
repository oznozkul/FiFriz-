package integrations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type GmailSettings struct {
	Query             string
	SyncFromDate      time.Time
	ImportAttachments bool
	ImportEmailBody   bool
	AllowedExtensions []string
	MaxEmailsPerSync  int
}

func defaultGmailSettings() GmailSettings {
	return GmailSettings{
		Query:             "has:attachment",
		SyncFromDate:      time.Now().UTC().AddDate(0, -1, 0),
		ImportAttachments: true,
		ImportEmailBody:   true,
		AllowedExtensions: []string{"pdf", "docx", "xlsx", "png", "jpg", "jpeg"},
		MaxEmailsPerSync:  DefaultMaxEmailsPerSync,
	}
}

func ParseGmailSettings(raw []byte) (GmailSettings, error) {
	settings := defaultGmailSettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return settings, fmt.Errorf("invalid Gmail settings")
	}

	if query := stringValue(payload["query"]); query != "" {
		settings.Query = query
	}
	if syncFromDate := stringValue(payload["syncFromDate"]); syncFromDate != "" {
		parsed, err := time.Parse("2006-01-02", syncFromDate)
		if err != nil {
			return settings, fmt.Errorf("invalid syncFromDate")
		}
		settings.SyncFromDate = parsed.UTC()
	}
	if importAttachments, ok := payload["importAttachments"].(bool); ok {
		settings.ImportAttachments = importAttachments
	}
	if importEmailBody, ok := payload["importEmailBody"].(bool); ok {
		settings.ImportEmailBody = importEmailBody
	}
	if extensions, ok := payload["allowedExtensions"].([]any); ok && len(extensions) > 0 {
		settings.AllowedExtensions = normalizeExtensions(parseStringSlice(extensions))
	}
	if maxEmails := intValue(payload["maxEmailsPerSync"]); maxEmails > 0 {
		settings.MaxEmailsPerSync = maxEmails
	}

	return settings, nil
}

func (s GmailSettings) BuildQuery() string {
	parts := make([]string, 0, 3)
	if trimmed := strings.TrimSpace(s.Query); trimmed != "" {
		parts = append(parts, trimmed)
	} else if s.ImportAttachments {
		parts = append(parts, "has:attachment")
	}
	if !s.SyncFromDate.IsZero() {
		date := s.SyncFromDate.UTC()
		parts = append(parts, fmt.Sprintf("after:%d/%02d/%02d", date.Year(), int(date.Month()), date.Day()))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func (s GmailSettings) IsExtensionAllowed(filename string) bool {
	ext := strings.TrimPrefix(strings.ToLower(extensionFromFilename(filename)), ".")
	if ext == "" {
		return false
	}
	for _, allowed := range s.AllowedExtensions {
		if allowed == ext {
			return true
		}
	}
	return false
}

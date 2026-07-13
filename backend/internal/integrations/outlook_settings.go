package integrations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type OutlookSettings struct {
	Folder            string
	SyncFromDate      time.Time
	ImportAttachments bool
	ImportEmailBody   bool
	AllowedExtensions []string
	MaxEmailsPerSync  int
}

func defaultOutlookSettings() OutlookSettings {
	return OutlookSettings{
		Folder:            "Inbox",
		SyncFromDate:      time.Now().UTC().AddDate(0, -1, 0),
		ImportAttachments: true,
		ImportEmailBody:   true,
		AllowedExtensions: []string{"pdf", "docx", "xlsx", "png", "jpg", "jpeg"},
		MaxEmailsPerSync:  DefaultMaxEmailsPerSync,
	}
}

func ParseOutlookSettings(raw []byte) (OutlookSettings, error) {
	settings := defaultOutlookSettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return settings, fmt.Errorf("invalid Outlook settings")
	}

	if folder := stringValue(payload["folder"]); folder != "" {
		settings.Folder = folder
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

func (s OutlookSettings) BuildFilter() string {
	parts := make([]string, 0, 2)
	if !s.SyncFromDate.IsZero() {
		date := s.SyncFromDate.UTC().Format("2006-01-02T15:04:05Z")
		parts = append(parts, fmt.Sprintf("receivedDateTime ge %s", date))
	}
	if s.ImportAttachments && !s.ImportEmailBody {
		parts = append(parts, "hasAttachments eq true")
	}
	return strings.Join(parts, " and ")
}

func (s OutlookSettings) MailFolderPath() string {
	folder := strings.TrimSpace(s.Folder)
	if folder == "" {
		folder = "Inbox"
	}
	return strings.ToLower(folder)
}

func (s OutlookSettings) IsExtensionAllowed(filename string) bool {
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

func outlookSettingsToMap(settings OutlookSettings) map[string]any {
	return map[string]any{
		"folder":            settings.Folder,
		"syncFromDate":      settings.SyncFromDate.Format("2006-01-02"),
		"importAttachments": settings.ImportAttachments,
		"importEmailBody":   settings.ImportEmailBody,
		"allowedExtensions": settings.AllowedExtensions,
		"maxEmailsPerSync":  settings.MaxEmailsPerSync,
	}
}

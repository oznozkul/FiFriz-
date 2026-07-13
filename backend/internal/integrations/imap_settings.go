package integrations

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type IMAPSettings struct {
	Host              string
	Port              int
	Username          string
	UseTLS            bool
	SecurityMode      string
	TLSServerName     string
	InsecureSkipTLS   bool
	Mailbox           string
	SyncFromDate      time.Time
	ImportAttachments bool
	ImportEmailBody   bool
	AllowedExtensions []string
	MaxEmailsPerSync  int
}

func ParseIMAPSettings(raw []byte) (IMAPSettings, error) {
	settings := defaultIMAPSettings()
	if len(raw) == 0 {
		return settings, fmt.Errorf("IMAP settings are not configured")
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return settings, fmt.Errorf("invalid IMAP settings")
	}

	if host := stringValue(payload["host"]); host != "" {
		settings.Host = NormalizeIMAPHost(host)
	}
	if port := intValue(payload["port"]); port > 0 {
		settings.Port = port
	}
	if username := stringValue(payload["username"]); username != "" {
		settings.Username = username
	}
	if useTLS, ok := payload["useTLS"].(bool); ok {
		settings.UseTLS = useTLS
	}
	if securityMode := stringValue(payload["securityMode"]); securityMode != "" {
		settings.SecurityMode = normalizeSecurityMode(securityMode)
	}
	if tlsServerName := stringValue(payload["tlsServerName"]); tlsServerName != "" {
		settings.TLSServerName = NormalizeIMAPHost(tlsServerName)
	}
	if insecureSkipTLS, ok := payload["insecureSkipTLS"].(bool); ok {
		settings.InsecureSkipTLS = insecureSkipTLS
	}
	if mailbox := stringValue(payload["mailbox"]); mailbox != "" {
		settings.Mailbox = mailbox
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

	if settings.Host == "" || settings.Username == "" {
		return settings, fmt.Errorf("IMAP host and username are required")
	}

	return settings, nil
}

func applyIMAPSyncFromDate(settings *IMAPSettings, input map[string]any) error {
	if settings == nil || len(input) == 0 {
		return nil
	}
	syncFromDate := stringValue(input["syncFromDate"])
	if syncFromDate == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", syncFromDate)
	if err != nil {
		return fmt.Errorf("invalid syncFromDate")
	}
	settings.SyncFromDate = parsed.UTC()
	return nil
}

func mergeIMAPUserSettings(existing []byte, updates map[string]any) ([]byte, error) {
	current := decodeJSONMap(existing)
	if current == nil {
		current = map[string]any{}
	}

	syncFromDate := stringValue(updates["syncFromDate"])
	if syncFromDate == "" {
		return nil, fmt.Errorf("syncFromDate is required")
	}
	if _, err := time.Parse("2006-01-02", syncFromDate); err != nil {
		return nil, fmt.Errorf("invalid syncFromDate")
	}
	current["syncFromDate"] = syncFromDate
	if maxEmails := intValue(updates["maxEmailsPerSync"]); maxEmails > 0 {
		current["maxEmailsPerSync"] = maxEmails
	}
	return encodeJSONMap(current)
}

func preservedIMAPSyncFromDate(raw []byte) (time.Time, bool) {
	if len(raw) == 0 {
		return time.Time{}, false
	}
	if settings, err := ParseIMAPSettings(raw); err == nil && !settings.SyncFromDate.IsZero() {
		return settings.SyncFromDate, true
	}
	payload := decodeJSONMap(raw)
	syncFromDate := stringValue(payload["syncFromDate"])
	if syncFromDate == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", syncFromDate)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func defaultIMAPSettings() IMAPSettings {
	return IMAPSettings{
		Port:              993,
		UseTLS:            true,
		Mailbox:           "INBOX",
		SyncFromDate:      time.Now().UTC().AddDate(0, -1, 0),
		ImportAttachments: true,
		ImportEmailBody:   true,
		AllowedExtensions: []string{"pdf", "docx", "xlsx", "png", "jpg", "jpeg"},
		MaxEmailsPerSync:  DefaultMaxEmailsPerSync,
	}
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strings.TrimSpace(strconv.FormatInt(int64(typed), 10))
	default:
		return ""
	}
}

func intValue(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func parseStringSlice(values []any) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text := stringValue(value); text != "" {
			result = append(result, text)
		}
	}
	return result
}

func normalizeExtensions(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		ext := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), ".")
		if ext == "" {
			continue
		}
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		result = append(result, ext)
	}
	return result
}

func (s IMAPSettings) IsExtensionAllowed(filename string) bool {
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

// NormalizeIMAPHost strips accidental protocol/port suffixes from user input.
func NormalizeIMAPHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "imaps://")
	host = strings.TrimPrefix(host, "imap://")
	if idx := strings.Index(host, "/"); idx >= 0 {
		host = host[:idx]
	}
	if idx := strings.Index(host, ":"); idx >= 0 {
		host = host[:idx]
	}
	return strings.TrimSpace(host)
}

func normalizeSecurityMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "ssl", "starttls", "none":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return ""
	}
}

// ConnectionMode returns ssl, starttls, or none.
func (s IMAPSettings) ConnectionMode() string {
	if mode := normalizeSecurityMode(s.SecurityMode); mode != "" {
		return mode
	}
	if !s.UseTLS {
		return "none"
	}
	if s.Port == 143 {
		return "starttls"
	}
	return "ssl"
}

func extensionFromFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ""
	}
	if idx := strings.LastIndex(filename, "."); idx >= 0 && idx < len(filename)-1 {
		return filename[idx+1:]
	}
	return ""
}

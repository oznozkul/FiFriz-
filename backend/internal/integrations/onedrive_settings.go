package integrations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type OneDriveSettings struct {
	FolderID          *string
	SyncFromDate      time.Time
	AllowedExtensions []string
	MaxFilesPerSync   int
	IncludeSubfolders bool
}

func defaultOneDriveSettings() OneDriveSettings {
	return OneDriveSettings{
		FolderID:          nil,
		SyncFromDate:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		AllowedExtensions: []string{"pdf", "docx", "xlsx", "png", "jpg", "jpeg"},
		MaxFilesPerSync:   DefaultMaxFilesPerSync,
		IncludeSubfolders: true,
	}
}

func ParseOneDriveSettings(raw []byte) (OneDriveSettings, error) {
	settings := defaultOneDriveSettings()
	if len(raw) == 0 {
		return settings, nil
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return settings, fmt.Errorf("invalid OneDrive settings")
	}

	if folderID, ok := payload["folderId"].(string); ok {
		trimmed := strings.TrimSpace(folderID)
		if trimmed != "" {
			settings.FolderID = &trimmed
		}
	}
	if syncFromDate := stringValue(payload["syncFromDate"]); syncFromDate != "" {
		parsed, err := time.Parse("2006-01-02", syncFromDate)
		if err != nil {
			return settings, fmt.Errorf("invalid syncFromDate")
		}
		settings.SyncFromDate = parsed.UTC()
	}
	if extensions, ok := payload["allowedExtensions"].([]any); ok && len(extensions) > 0 {
		settings.AllowedExtensions = normalizeExtensions(parseStringSlice(extensions))
	}
	if maxFiles := intValue(payload["maxFilesPerSync"]); maxFiles > 0 {
		settings.MaxFilesPerSync = maxFiles
	}
	if includeSubfolders, ok := payload["includeSubfolders"].(bool); ok {
		settings.IncludeSubfolders = includeSubfolders
	}

	return settings, nil
}

func (s OneDriveSettings) RootFolderID() string {
	if s.FolderID != nil {
		if trimmed := strings.TrimSpace(*s.FolderID); trimmed != "" {
			return trimmed
		}
	}
	return "root"
}

func (s OneDriveSettings) IsExtensionAllowed(filename string, mimeType string) bool {
	ext := strings.TrimPrefix(strings.ToLower(extensionFromFilename(filename)), ".")
	if ext == "" {
		ext = extensionFromGraphMimeType(mimeType)
	}
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

func oneDriveSettingsToMap(settings OneDriveSettings) map[string]any {
	result := map[string]any{
		"syncFromDate":      settings.SyncFromDate.Format("2006-01-02"),
		"allowedExtensions": settings.AllowedExtensions,
		"maxFilesPerSync":   settings.MaxFilesPerSync,
		"includeSubfolders": settings.IncludeSubfolders,
	}
	if settings.FolderID != nil && strings.TrimSpace(*settings.FolderID) != "" {
		result["folderId"] = strings.TrimSpace(*settings.FolderID)
	} else {
		result["folderId"] = nil
	}
	return result
}

func extensionFromGraphMimeType(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "application/pdf":
		return "pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "xlsx"
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	default:
		return ""
	}
}

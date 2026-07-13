package search

import (
	"strings"
)

var staticActions = []GlobalActionResult{
	{ID: "upload-document", Label: "Belge Yükle", Kind: "action"},
	{ID: "new-folder", Label: "Yeni Klasör", Path: "/folders", Kind: "navigate"},
	{ID: "new-tag", Label: "Yeni Etiket", Path: "/tags", Kind: "navigate"},
	{ID: "go-dashboard", Label: "Dashboard'a Git", Path: "/dashboard", Kind: "navigate"},
	{ID: "go-trash", Label: "Çöp Kutusuna Git", Path: "/trash", Kind: "navigate"},
	{ID: "go-integrations", Label: "Entegrasyonlara Git", Path: "/integrations", Kind: "navigate"},
	{ID: "open-notifications", Label: "Bildirimleri Aç", Path: "/notifications", Kind: "navigate"},
	{ID: "open-profile", Label: "Profili Aç", Path: "/settings/profile", Kind: "navigate"},
}

var staticSettings = []GlobalSettingResult{
	{ID: "settings-profile", Label: "Profil", Path: "/settings/profile"},
	{ID: "settings-security", Label: "Güvenlik", Path: "/settings/security"},
	{ID: "settings-workspace", Label: "Workspace", Path: "/settings/workspace"},
	{ID: "settings-members", Label: "Üyeler", Path: "/settings/members"},
	{ID: "settings-storage", Label: "Depolama", Path: "/settings/storage"},
	{ID: "settings-billing", Label: "Faturalandırma", Path: "/settings/billing"},
}

func filterActions(query string) []GlobalActionResult {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return staticActions
	}
	out := make([]GlobalActionResult, 0)
	for _, item := range staticActions {
		if strings.Contains(strings.ToLower(item.Label), query) {
			out = append(out, item)
		}
	}
	return out
}

func filterSettings(query string) []GlobalSettingResult {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return staticSettings
	}
	out := make([]GlobalSettingResult, 0)
	for _, item := range staticSettings {
		if strings.Contains(strings.ToLower(item.Label), query) {
			out = append(out, item)
		}
	}
	return out
}

func matchesQuery(value, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(value), query)
}

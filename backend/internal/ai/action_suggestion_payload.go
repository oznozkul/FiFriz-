package ai

import (
	"fmt"
	"strings"
	"time"

	"fifriz/internal/ai/tools"
	"fifriz/internal/documents"

	"github.com/google/uuid"
)

func withToolPayload(toolName string, input map[string]interface{}, extras map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		"toolName": toolName,
		"input":    input,
	}
	for key, value := range extras {
		payload[key] = value
	}
	return payload
}

func toolInputFromPayload(payload map[string]interface{}) (string, map[string]interface{}, bool) {
	toolName, ok := payload["toolName"].(string)
	if !ok || strings.TrimSpace(toolName) == "" {
		return "", nil, false
	}
	rawInput, ok := payload["input"].(map[string]interface{})
	if !ok {
		return strings.TrimSpace(toolName), map[string]interface{}{}, true
	}
	return strings.TrimSpace(toolName), rawInput, true
}

func toolInputForSuggestion(suggestionType string, documentID uuid.UUID, payload map[string]interface{}) (string, map[string]interface{}, bool) {
	if toolName, input, ok := toolInputFromPayload(payload); ok {
		return toolName, input, true
	}
	return legacyToolInputForSuggestion(suggestionType, documentID, payload)
}

func legacyToolInputForSuggestion(suggestionType string, documentID uuid.UUID, payload map[string]interface{}) (string, map[string]interface{}, bool) {
	documentIDValue := documentID.String()
	switch suggestionType {
	case ActionSuggestionTypeAddTag:
		tagName, _ := payload["tagName"].(string)
		if strings.TrimSpace(tagName) == "" {
			return "", nil, false
		}
		return tools.ToolAddTagToDocument, map[string]interface{}{
			"documentId": documentIDValue,
			"tagName":      strings.TrimSpace(tagName),
		}, true
	case ActionSuggestionTypeMoveFolder:
		folderID, _ := payload["folderId"].(string)
		if strings.TrimSpace(folderID) == "" {
			return "", nil, false
		}
		return tools.ToolMoveDocumentToFolder, map[string]interface{}{
			"documentId": documentIDValue,
			"folderId":   strings.TrimSpace(folderID),
		}, true
	case ActionSuggestionTypeCreateReminder:
		title, _ := payload["dateTitle"].(string)
		suggestedDate, _ := payload["suggestedDate"].(string)
		if strings.TrimSpace(title) == "" || strings.TrimSpace(suggestedDate) == "" {
			return "", nil, false
		}
		daysBefore := 30
		if value, ok := payload["daysBefore"].(float64); ok && int(value) > 0 {
			daysBefore = int(value)
		}
		return tools.ToolCreateImportantDate, map[string]interface{}{
			"documentId":         documentIDValue,
			"title":              strings.TrimSpace(title),
			"date":               strings.TrimSpace(suggestedDate),
			"reminderDaysBefore": daysBefore,
		}, true
	case ActionSuggestionTypeStartApproval:
		return "", nil, false
	case ActionSuggestionTypeLinkRelated:
		targetID, _ := payload["targetDocumentId"].(string)
		if strings.TrimSpace(targetID) == "" {
			return "", nil, false
		}
		return tools.ToolCreateDocumentRelationship, map[string]interface{}{
			"documentId":       documentIDValue,
			"targetDocumentId": strings.TrimSpace(targetID),
			"relationshipType": documents.RelationshipTypeRelated,
		}, true
	case ActionSuggestionTypeRegenerateMetadata:
		return tools.ToolReprocessDocument, map[string]interface{}{"documentId": documentIDValue}, true
	case ActionSuggestionTypeGenerateSummary:
		return tools.ToolGenerateSummary, map[string]interface{}{"documentId": documentIDValue}, true
	default:
		return "", nil, false
	}
}

func toolConfirmationText(toolName string, payload map[string]interface{}) string {
	switch toolName {
	case tools.ToolAddTagToDocument:
		return fmt.Sprintf("Bu belgeye \"%v\" etiketi eklenecek.", payload["tagName"])
	case tools.ToolMoveDocumentToFolder:
		folderName := payload["folderName"]
		if folderName == nil {
			folderName = payload["folderId"]
		}
		return fmt.Sprintf("Bu belge \"%v\" klasörüne taşınacak.", folderName)
	case tools.ToolCreateImportantDate:
		return fmt.Sprintf("Bu belge için \"%v\" başlıklı önemli tarih oluşturulacak.", payload["dateTitle"])
	case tools.ToolCreateWorkflow:
		return fmt.Sprintf("Bu belge için \"%v\" onay süreci başlatılacak.", payload["workflowTitle"])
	case tools.ToolCreateDocumentRelationship:
		return fmt.Sprintf("Bu belge \"%v\" ile ilişkilendirilecek.", payload["targetDocumentName"])
	case tools.ToolGenerateSummary:
		return "Bu belge için AI özeti oluşturulacak."
	case tools.ToolReprocessDocument:
		return "Bu belge yeniden işlenecek ve analiz süreci başlatılacak."
	case tools.ToolFavoriteDocument:
		return "Bu belge favorilere eklenecek."
	default:
		return "Seçili AI önerisi uygulanacak."
	}
}

func buildToolPayloadExtras(panel string) map[string]interface{} {
	return map[string]interface{}{"panel": panel}
}

func reminderPayloadExtras(reminder *reminderSuggestion, documentID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		"panel":          "dates",
		"daysBefore":     reminder.DaysBefore,
		"dateTitle":      reminder.Title,
		"suggestedDate":  reminder.Date,
		"documentId":     documentID.String(),
		"confirmMessage": fmt.Sprintf("%s tarihi için %d gün önce hatırlatıcı oluşturulacak.", reminder.Date, reminder.DaysBefore),
	}
}

func parseSuggestedDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

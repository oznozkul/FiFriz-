package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"fifriz/internal/ai/tools"
)

type agentLLMResponse struct {
	Answer  string              `json:"answer"`
	Actions []agentLLMAction    `json:"actions"`
}

type agentLLMAction struct {
	ToolName    string                 `json:"toolName"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
}

func agentSystemPrompt(registry *tools.Registry) string {
	var builder strings.Builder
	builder.WriteString(`Sen FiFriz workspace AI asistanısın.

KURALLAR:
- Kullanıcı onayı olmadan hiçbir aksiyon çalıştırma; yalnızca öneri üret.
- Bilmediğin belge veya veri hakkında uydurma; yalnızca verilen bağlama dayan.
- Yeterli bilgi yoksa bunu açıkça belirt.
- Paylaşım, silme veya yetki dışı işlemler için aksiyon önerme; kullanıcıyı ilgili ekrana yönlendir.
- Cevaplarını Türkçe ve anlaşılır yaz.
- Aksiyon öneriyorsan yalnızca aşağıdaki araçları kullan.

KULLANILABİLİR ARAÇLAR:
`)
	for _, tool := range registry.List() {
		builder.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
	}

	builder.WriteString(`
YANIT FORMATI:
Yalnızca geçerli JSON döndür, başka metin ekleme:
{
  "answer": "Kullanıcıya Türkçe cevap",
  "actions": [
    {
      "toolName": "araç_adı",
      "title": "Öneri başlığı",
      "description": "Kısa açıklama",
      "input": { }
    }
  ]
}

Aksiyon gerekmiyorsa "actions" boş dizi olmalı.
input içinde documentId, folderId, tagName gibi alanları bağlamdaki gerçek kimliklerle doldur.
`)
	return strings.TrimSpace(builder.String())
}

func buildAgentUserPrompt(message string, contextText string, chunks []string) string {
	var builder strings.Builder
	builder.WriteString("WORKSPACE BAĞLAMI:\n")
	builder.WriteString(contextText)
	builder.WriteString("\n\n")

	if len(chunks) > 0 {
		builder.WriteString("İLGİLİ BELGE PARÇALARI:\n")
		for index, chunk := range chunks {
			builder.WriteString(fmt.Sprintf("\n[%d]\n%s\n", index+1, chunk))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("KULLANICI MESAJI:\n")
	builder.WriteString(message)
	return builder.String()
}

func parseAgentLLMResponse(raw string) (agentLLMResponse, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return agentLLMResponse{}, fmt.Errorf("empty agent response")
	}

	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(strings.TrimSpace(trimmed), "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	var response agentLLMResponse
	if err := json.Unmarshal([]byte(trimmed), &response); err != nil {
		start := strings.Index(trimmed, "{")
		end := strings.LastIndex(trimmed, "}")
		if start >= 0 && end > start {
			if err := json.Unmarshal([]byte(trimmed[start:end+1]), &response); err != nil {
				return agentLLMResponse{Answer: trimmed}, nil
			}
			return response, nil
		}
		return agentLLMResponse{Answer: trimmed}, nil
	}
	return response, nil
}

func suggestionTypeForTool(toolName string) string {
	switch toolName {
	case tools.ToolAddTagToDocument:
		return ActionSuggestionTypeAddTag
	case tools.ToolMoveDocumentToFolder:
		return ActionSuggestionTypeMoveFolder
	case tools.ToolCreateImportantDate:
		return ActionSuggestionTypeCreateReminder
	case tools.ToolCreateWorkflow:
		return ActionSuggestionTypeStartApproval
	case tools.ToolCreateDocumentRelationship:
		return ActionSuggestionTypeLinkRelated
	case tools.ToolGenerateSummary:
		return ActionSuggestionTypeGenerateSummary
	case tools.ToolReprocessDocument:
		return ActionSuggestionTypeRegenerateMetadata
	case tools.ToolFavoriteDocument:
		return ActionSuggestionTypeFavoriteDocument
	default:
		return toolName
	}
}

func panelForTool(toolName string) string {
	switch toolName {
	case tools.ToolAddTagToDocument:
		return "tags"
	case tools.ToolMoveDocumentToFolder:
		return "folder"
	case tools.ToolCreateImportantDate:
		return "dates"
	case tools.ToolCreateWorkflow:
		return "workflow"
	case tools.ToolCreateDocumentRelationship:
		return "related"
	case tools.ToolGenerateSummary:
		return "summary"
	case tools.ToolReprocessDocument:
		return "aiMetadata"
	case tools.ToolFavoriteDocument:
		return ""
	default:
		return ""
	}
}

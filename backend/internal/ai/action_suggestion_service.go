package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"fifriz/internal/ai/tools"
	"fifriz/internal/approvals"
	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/folders"
	"fifriz/internal/tags"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidDocumentID      = errors.New("invalid document id")
	ErrDocumentNotFound       = errors.New("document not found")
	ErrActionSuggestionDenied = errors.New("forbidden")
	ErrInvalidSuggestionID    = errors.New("invalid suggestion id")
	ErrSuggestionNotPending   = errors.New("action suggestion is not pending")
	ErrToolExecutionFailed    = errors.New("tool execution failed")
)

type DocumentViewer interface {
	CanView(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
}

type ActionSuggestionService struct {
	repository         *ActionSuggestionRepository
	documents          *documents.Repository
	folders            *folders.Repository
	tags               *tags.Repository
	approvals          *approvals.Repository
	access             DocumentViewer
	audit              audit.Recorder
	toolExecutor       *tools.Executor
}

func NewActionSuggestionService(
	repository *ActionSuggestionRepository,
	documentRepository *documents.Repository,
	folderRepository *folders.Repository,
	tagRepository *tags.Repository,
	approvalRepository *approvals.Repository,
	access DocumentViewer,
	auditRecorder audit.Recorder,
	toolExecutor *tools.Executor,
) *ActionSuggestionService {
	return &ActionSuggestionService{
		repository:   repository,
		documents:    documentRepository,
		folders:      folderRepository,
		tags:         tagRepository,
		approvals:    approvalRepository,
		access:       access,
		audit:        auditRecorder,
		toolExecutor: toolExecutor,
	}
}

func (s *ActionSuggestionService) ListForUser(ctx context.Context) ([]ActionSuggestionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.repository.ListPendingByUser(ctx, authContext.TenantID, authContext.UserID, nil, 50)
	if err != nil {
		return nil, err
	}
	return s.filterAccessibleSuggestions(ctx, authContext, items)
}

func (s *ActionSuggestionService) ListForDocument(ctx context.Context, documentIDRaw string) ([]ActionSuggestionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	documentID, document, err := s.loadViewableDocument(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}
	items, err := s.repository.ListPendingByUser(ctx, authContext.TenantID, authContext.UserID, &documentID, 20)
	if err != nil {
		return nil, err
	}
	_ = document
	return toActionSuggestionResponses(items)
}

func (s *ActionSuggestionService) GenerateForDocument(ctx context.Context, documentIDRaw string) ([]ActionSuggestionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	documentID, document, err := s.loadViewableDocument(ctx, authContext, documentIDRaw)
	if err != nil {
		return nil, err
	}

	if err := s.repository.DeletePendingByDocument(ctx, authContext.TenantID, authContext.UserID, documentID); err != nil {
		return nil, err
	}

	candidates, err := s.buildDocumentSuggestions(ctx, authContext, document)
	if err != nil {
		return nil, err
	}
	if len(candidates) > 0 {
		if err := s.repository.CreateBatch(ctx, candidates); err != nil {
			return nil, err
		}
	}

	userID := authContext.UserID
	audit.LogDocumentAction(ctx, s.audit, authContext.TenantID, &userID, audit.ActionAIActionSuggestionsGenerated, documentID, "ai action suggestions generated", map[string]interface{}{
		"count": len(candidates),
	})

	items, err := s.repository.ListPendingByUser(ctx, authContext.TenantID, authContext.UserID, &documentID, 20)
	if err != nil {
		return nil, err
	}
	return toActionSuggestionResponses(items)
}

func (s *ActionSuggestionService) Accept(ctx context.Context, suggestionIDRaw string) (*ActionSuggestionAcceptResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	suggestionID, err := uuid.Parse(strings.TrimSpace(suggestionIDRaw))
	if err != nil {
		return nil, ErrInvalidSuggestionID
	}

	item, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, suggestionID)
	if err != nil {
		return nil, err
	}
	if item.Status != ActionSuggestionStatusPending {
		return nil, ErrSuggestionNotPending
	}

	payload := map[string]interface{}{}
	if len(item.ActionPayload) > 0 {
		_ = json.Unmarshal(item.ActionPayload, &payload)
	}

	var documentID *uuid.UUID
	if item.DocumentID != nil {
		document, loadErr := s.documents.GetByID(ctx, authContext.TenantID, *item.DocumentID)
		if loadErr != nil {
			if errors.Is(loadErr, gorm.ErrRecordNotFound) {
				return nil, ErrDocumentNotFound
			}
			return nil, loadErr
		}
		if err := s.requireViewAccess(ctx, authContext, document); err != nil {
			return nil, err
		}
		documentID = item.DocumentID
	}

	response := ActionSuggestionAcceptResponse{Executed: false}
	toolName, toolInput, hasTool := "", map[string]interface{}{}, false
	if documentID != nil {
		toolName, toolInput, hasTool = toolInputForSuggestion(item.SuggestionType, *documentID, payload)
	}

	if hasTool && s.toolExecutor != nil {
		execution, execErr := s.toolExecutor.Execute(ctx, authContext, toolName, toolInput, tools.ExecuteOptions{
			SuggestionID: &suggestionID,
			DocumentID:   documentID,
		})
		if execErr != nil {
			return nil, execErr
		}
		response.Executed = true
		response.Execution = execution
	}

	if err := s.repository.UpdateStatus(ctx, authContext.TenantID, authContext.UserID, suggestionID, ActionSuggestionStatusAccepted); err != nil {
		return nil, err
	}
	item.Status = ActionSuggestionStatusAccepted

	userID := authContext.UserID
	if documentID != nil {
		metadata := map[string]interface{}{
			"suggestionId":   suggestionID,
			"suggestionType": item.SuggestionType,
			"status":         ActionSuggestionStatusAccepted,
			"executed":       response.Executed,
		}
		if response.Execution != nil {
			metadata["toolName"] = response.Execution.ToolName
			metadata["input"] = toolInput
		}
		audit.LogDocumentAction(ctx, s.audit, authContext.TenantID, &userID, audit.ActionAIActionSuggestionAccepted, *documentID, "ai action suggestion accepted", metadata)
	}

	suggestionResponse, err := toActionSuggestionResponse(*item)
	if err != nil {
		return nil, err
	}
	response.Suggestion = suggestionResponse
	return &response, nil
}

func (s *ActionSuggestionService) Reject(ctx context.Context, suggestionIDRaw string) (*ActionSuggestionResponse, error) {
	return s.updateSuggestionStatus(ctx, suggestionIDRaw, ActionSuggestionStatusRejected, audit.ActionAIActionSuggestionRejected)
}

func (s *ActionSuggestionService) updateSuggestionStatus(ctx context.Context, suggestionIDRaw, status, auditAction string) (*ActionSuggestionResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	suggestionID, err := uuid.Parse(strings.TrimSpace(suggestionIDRaw))
	if err != nil {
		return nil, ErrInvalidSuggestionID
	}

	item, err := s.repository.GetByID(ctx, authContext.TenantID, authContext.UserID, suggestionID)
	if err != nil {
		return nil, err
	}
	if item.Status != ActionSuggestionStatusPending {
		return nil, ErrSuggestionNotPending
	}

	if item.DocumentID != nil {
		document, loadErr := s.documents.GetByID(ctx, authContext.TenantID, *item.DocumentID)
		if loadErr != nil {
			if errors.Is(loadErr, gorm.ErrRecordNotFound) {
				return nil, ErrDocumentNotFound
			}
			return nil, loadErr
		}
		if err := s.requireViewAccess(ctx, authContext, document); err != nil {
			return nil, err
		}
	}

	if err := s.repository.UpdateStatus(ctx, authContext.TenantID, authContext.UserID, suggestionID, status); err != nil {
		return nil, err
	}
	item.Status = status

	userID := authContext.UserID
	if item.DocumentID != nil {
		audit.LogDocumentAction(ctx, s.audit, authContext.TenantID, &userID, auditAction, *item.DocumentID, "ai action suggestion updated", map[string]interface{}{
			"suggestionId":   suggestionID,
			"suggestionType": item.SuggestionType,
			"status":         status,
		})
	}

	response, err := toActionSuggestionResponse(*item)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *ActionSuggestionService) filterAccessibleSuggestions(ctx context.Context, authContext auth.Context, items []AIActionSuggestion) ([]ActionSuggestionResponse, error) {
	filtered := make([]AIActionSuggestion, 0, len(items))
	for _, item := range items {
		if item.DocumentID == nil {
			filtered = append(filtered, item)
			continue
		}
		document, err := s.documents.GetByID(ctx, authContext.TenantID, *item.DocumentID)
		if err != nil {
			continue
		}
		allowed, err := s.access.CanView(ctx, authContext, document)
		if err != nil || !allowed {
			continue
		}
		filtered = append(filtered, item)
	}
	return toActionSuggestionResponses(filtered)
}

func (s *ActionSuggestionService) buildDocumentSuggestions(ctx context.Context, authContext auth.Context, document *documents.Document) ([]AIActionSuggestion, error) {
	suggestions := make([]AIActionSuggestion, 0, 8)
	appendSuggestion := func(suggestionType, title string, description *string, payload map[string]interface{}) {
		suggestions = append(suggestions, s.newSuggestion(authContext, document.ID, suggestionType, title, description, payload))
	}

	tagMap, err := s.tags.ListTagsByDocumentIDs(ctx, authContext.TenantID, []uuid.UUID{document.ID})
	if err != nil {
		return nil, err
	}
	documentTags := tagMap[document.ID]

	if tagName := suggestedTagForDocumentType(document.AIDocumentType); tagName != "" && !hasTagName(documentTags, tagName) {
		extras := buildToolPayloadExtras("tags")
		extras["tagName"] = tagName
		extras["confirmMessage"] = toolConfirmationText(tools.ToolAddTagToDocument, extras)
		appendSuggestion(
			ActionSuggestionTypeAddTag,
			fmt.Sprintf("Bu belgeye \"%s\" etiketi ekle", tagName),
			stringPtr(fmt.Sprintf("AI belge türü %s olarak sınıflandırıldı.", humanDocumentType(document.AIDocumentType))),
			withToolPayload(tools.ToolAddTagToDocument, map[string]interface{}{
				"documentId": document.ID.String(),
				"tagName":    tagName,
			}, extras),
		)
	}

	if folderSuggestion := s.suggestFolderMove(ctx, authContext.TenantID, document); folderSuggestion != nil {
		extras := buildToolPayloadExtras("folder")
		extras["folderName"] = folderSuggestion.Name
		extras["confirmMessage"] = toolConfirmationText(tools.ToolMoveDocumentToFolder, extras)
		appendSuggestion(
			ActionSuggestionTypeMoveFolder,
			fmt.Sprintf("Bu belgeyi \"%s\" klasörüne taşı", folderSuggestion.Name),
			stringPtr("Belge türüne uygun klasör önerisi."),
			withToolPayload(tools.ToolMoveDocumentToFolder, map[string]interface{}{
				"documentId": document.ID.String(),
				"folderId":   folderSuggestion.ID.String(),
			}, extras),
		)
	}

	if reminder := s.suggestReminder(ctx, authContext.TenantID, document.ID); reminder != nil {
		extras := reminderPayloadExtras(reminder, document.ID)
		appendSuggestion(
			ActionSuggestionTypeCreateReminder,
			fmt.Sprintf("Bu belge için %d gün önce hatırlatıcı oluştur", reminder.DaysBefore),
			stringPtr(fmt.Sprintf("Önerilen tarih: %s", reminder.Date)),
			withToolPayload(tools.ToolCreateImportantDate, map[string]interface{}{
				"documentId":         document.ID.String(),
				"title":              reminder.Title,
				"date":               reminder.Date,
				"reminderDaysBefore": reminder.DaysBefore,
			}, extras),
		)
	}

	hasPendingWorkflow, err := s.approvals.HasPendingWorkflow(ctx, authContext.TenantID, document.ID)
	if err != nil {
		return nil, err
	}
	if !hasPendingWorkflow && document.ApprovalStatus == documents.DocumentApprovalStatusNone && isApprovalCandidate(document.AIDocumentType) {
		appendSuggestion(
			ActionSuggestionTypeStartApproval,
			"Bu belgeye onay süreci başlat",
			stringPtr("Sözleşme veya fatura gibi belgeler için onay akışı önerilir."),
			map[string]interface{}{
				"workflowTitle":  fmt.Sprintf("%s onayı", document.OriginalFileName),
				"documentId":     document.ID.String(),
				"panel":          "workflow",
				"confirmMessage": "Onay sürecini başlatmak için onaycı seçmeniz gerekecek.",
			},
		)
	}

	relatedRows, err := s.documents.ListSuggestedRelatedDocuments(ctx, authContext.TenantID, document.ID)
	if err != nil {
		return nil, err
	}
	if len(relatedRows) > 0 {
		top := relatedRows[0]
		extras := buildToolPayloadExtras("related")
		extras["targetDocumentName"] = top.TargetOriginalFileName
		extras["confirmMessage"] = toolConfirmationText(tools.ToolCreateDocumentRelationship, extras)
		appendSuggestion(
			ActionSuggestionTypeLinkRelated,
			"Benzer belgelerle ilişkilendir",
			stringPtr(fmt.Sprintf("%s belgesiyle yüksek benzerlik bulundu.", top.TargetOriginalFileName)),
			withToolPayload(tools.ToolCreateDocumentRelationship, map[string]interface{}{
				"documentId":       document.ID.String(),
				"targetDocumentId": top.TargetDocumentID.String(),
				"relationshipType": documents.RelationshipTypeRelated,
			}, extras),
		)
	}

	if documents.IsReadyStatus(document.Status) {
		if _, err := s.documents.GetAIMetadataByDocumentID(ctx, authContext.TenantID, document.ID); errors.Is(err, documents.ErrDocumentAIMetadataNotFound) {
			extras := buildToolPayloadExtras("aiMetadata")
			extras["confirmMessage"] = toolConfirmationText(tools.ToolReprocessDocument, extras)
			appendSuggestion(
				ActionSuggestionTypeRegenerateMetadata,
				"Metadata eksik, yeniden analiz et",
				stringPtr("Belge işlendi ancak AI metadata henüz oluşturulmadı."),
				withToolPayload(tools.ToolReprocessDocument, map[string]interface{}{
					"documentId": document.ID.String(),
				}, extras),
			)
		} else if err != nil {
			return nil, err
		}

		if _, err := s.documents.GetSummaryByDocumentID(ctx, authContext.TenantID, document.ID); errors.Is(err, documents.ErrDocumentSummaryNotFound) {
			extras := buildToolPayloadExtras("summary")
			extras["confirmMessage"] = toolConfirmationText(tools.ToolGenerateSummary, extras)
			appendSuggestion(
				ActionSuggestionTypeGenerateSummary,
				"Özet oluştur",
				stringPtr("Bu belge için henüz AI özeti oluşturulmadı."),
				withToolPayload(tools.ToolGenerateSummary, map[string]interface{}{
					"documentId": document.ID.String(),
				}, extras),
			)
		} else if err != nil {
			return nil, err
		}
	}

	return suggestions, nil
}

type reminderSuggestion struct {
	Title      string
	Date       string
	DaysBefore int
}

func (s *ActionSuggestionService) suggestReminder(ctx context.Context, tenantID, documentID uuid.UUID) *reminderSuggestion {
	existing, err := s.documents.ListImportantDatesByDocument(ctx, tenantID, documentID)
	if err == nil && len(existing) > 0 {
		return nil
	}

	entities, err := s.documents.ListEntitiesByDocument(ctx, tenantID, documentID)
	if err != nil {
		return nil
	}

	for _, entity := range entities {
		if entity.EntityType != "date" {
			continue
		}
		parsed, parseErr := time.Parse("2006-01-02", strings.TrimSpace(entity.Value))
		if parseErr != nil {
			continue
		}
		if parsed.Before(time.Now().UTC()) {
			continue
		}
		label := "Önemli tarih"
		if entity.Label != nil && strings.TrimSpace(*entity.Label) != "" {
			label = strings.TrimSpace(*entity.Label)
		}
		return &reminderSuggestion{
			Title:      label,
			Date:       parsed.Format("2006-01-02"),
			DaysBefore: 30,
		}
	}
	return nil
}

func (s *ActionSuggestionService) suggestFolderMove(ctx context.Context, tenantID uuid.UUID, document *documents.Document) *folders.Folder {
	targetName := folderNameForDocumentType(document.AIDocumentType)
	if targetName == "" {
		return nil
	}
	if document.FolderID != nil {
		folder, err := s.folders.GetByID(ctx, tenantID, *document.FolderID)
		if err == nil && strings.EqualFold(strings.TrimSpace(folder.Name), targetName) {
			return nil
		}
	}
	allFolders, err := s.folders.List(ctx, tenantID)
	if err != nil {
		return nil
	}
	for i := range allFolders {
		if strings.EqualFold(strings.TrimSpace(allFolders[i].Name), targetName) {
			return &allFolders[i]
		}
	}
	for i := range allFolders {
		if strings.Contains(strings.ToLower(allFolders[i].Name), strings.ToLower(targetName)) {
			return &allFolders[i]
		}
	}
	return nil
}

func (s *ActionSuggestionService) newSuggestion(
	authContext auth.Context,
	documentID uuid.UUID,
	suggestionType, title string,
	description *string,
	payload map[string]interface{},
) AIActionSuggestion {
	raw, _ := json.Marshal(payload)
	docID := documentID
	return AIActionSuggestion{
		TenantID:       authContext.TenantID,
		UserID:         authContext.UserID,
		DocumentID:     &docID,
		SuggestionType: suggestionType,
		Title:          title,
		Description:    description,
		ActionPayload:  raw,
		Status:         ActionSuggestionStatusPending,
	}
}

func (s *ActionSuggestionService) loadViewableDocument(ctx context.Context, authContext auth.Context, documentIDRaw string) (uuid.UUID, *documents.Document, error) {
	documentID, err := uuid.Parse(strings.TrimSpace(documentIDRaw))
	if err != nil {
		return uuid.Nil, nil, ErrInvalidDocumentID
	}
	document, err := s.documents.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil, ErrDocumentNotFound
		}
		return uuid.Nil, nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return uuid.Nil, nil, err
	}
	return documentID, document, nil
}

func (s *ActionSuggestionService) requireViewAccess(ctx context.Context, authContext auth.Context, document *documents.Document) error {
	allowed, err := s.access.CanView(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrActionSuggestionDenied
	}
	return nil
}

func suggestedTagForDocumentType(documentType *string) string {
	if documentType == nil {
		return ""
	}
	switch strings.TrimSpace(*documentType) {
	case "contract":
		return "Sözleşme"
	case "invoice":
		return "Fatura"
	case "insurance_policy":
		return "Sigorta"
	case "court_decision":
		return "Hukuki"
	case "warranty_document":
		return "Garanti"
	default:
		return ""
	}
}

func folderNameForDocumentType(documentType *string) string {
	if documentType == nil {
		return ""
	}
	switch strings.TrimSpace(*documentType) {
	case "contract", "court_decision":
		return "Hukuki Belgeler"
	case "invoice":
		return "Faturalar"
	case "insurance_policy":
		return "Sigorta Belgeleri"
	default:
		return ""
	}
}

func humanDocumentType(documentType *string) string {
	if documentType == nil {
		return "belge"
	}
	labels := map[string]string{
		"contract":         "sözleşme",
		"invoice":          "fatura",
		"insurance_policy": "sigorta poliçesi",
		"court_decision":   "mahkeme kararı",
	}
	if label, ok := labels[strings.TrimSpace(*documentType)]; ok {
		return label
	}
	return strings.TrimSpace(*documentType)
}

func isApprovalCandidate(documentType *string) bool {
	if documentType == nil {
		return false
	}
	switch strings.TrimSpace(*documentType) {
	case "contract", "invoice", "insurance_policy":
		return true
	default:
		return false
	}
}

func hasTagName(items []tags.TagBrief, name string) bool {
	needle := strings.ToLower(strings.TrimSpace(name))
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item.Name)) == needle {
			return true
		}
	}
	return false
}

func stringPtr(value string) *string {
	return &value
}

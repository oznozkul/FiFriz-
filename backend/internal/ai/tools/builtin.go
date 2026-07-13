package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"fifriz/internal/approvals"
	"fifriz/internal/auth"
	"fifriz/internal/documents"

	"github.com/google/uuid"
)

const (
	ToolAddTagToDocument         = "add_tag_to_document"
	ToolMoveDocumentToFolder     = "move_document_to_folder"
	ToolCreateImportantDate      = "create_important_date"
	ToolCreateWorkflow           = "create_workflow"
	ToolCreateDocumentRelationship = "create_document_relationship"
	ToolGenerateSummary          = "generate_summary"
	ToolReprocessDocument        = "reprocess_document"
	ToolFavoriteDocument         = "favorite_document"
)

func RegisterBuiltInTools(registry *Registry, deps Dependencies) {
	if registry == nil {
		return
	}
	registry.Register(&addTagTool{deps: deps})
	registry.Register(&moveFolderTool{deps: deps})
	registry.Register(&createImportantDateTool{deps: deps})
	registry.Register(&createWorkflowTool{deps: deps})
	registry.Register(&createRelationshipTool{deps: deps})
	registry.Register(&generateSummaryTool{deps: deps})
	registry.Register(&reprocessDocumentTool{deps: deps})
	registry.Register(&favoriteDocumentTool{deps: deps})
}

type addTagTool struct{ deps Dependencies }

func (t *addTagTool) Name() string { return ToolAddTagToDocument }
func (t *addTagTool) Description() string {
	return "Belgeye etiket ekler."
}
func (t *addTagTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId", "tagName"},
		"properties": map[string]interface{}{
			"documentId": map[string]string{"type": "string"},
			"tagName":    map[string]string{"type": "string"},
		},
	}
}
func (t *addTagTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	document, err := t.loadEditableDocument(ctx, user, input)
	if err != nil {
		return err
	}
	_ = document
	if _, err := parseRequiredString(input, "tagName"); err != nil {
		return err
	}
	if t.deps.Tags == nil {
		return ErrToolNotAllowed
	}
	return nil
}
func (t *addTagTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	tagName, err := parseRequiredString(input, "tagName")
	if err != nil {
		return nil, err
	}
	tag, err := t.deps.Tags.FindOrCreateTag(ctx, user.TenantID, user.UserID, tagName)
	if err != nil {
		return nil, err
	}
	if err := t.deps.Tags.AppendDocumentTag(ctx, user.TenantID, documentID, tag.ID); err != nil {
		return nil, err
	}
	return map[string]interface{}{"documentId": documentID.String(), "tagId": tag.ID.String(), "tagName": tag.Name}, nil
}
func (t *addTagTool) loadEditableDocument(ctx context.Context, user auth.Context, input map[string]interface{}) (*documents.Document, error) {
	return loadEditableDocument(ctx, t.deps, user, input)
}

type moveFolderTool struct{ deps Dependencies }

func (t *moveFolderTool) Name() string { return ToolMoveDocumentToFolder }
func (t *moveFolderTool) Description() string {
	return "Belgeyi klasöre taşır."
}
func (t *moveFolderTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId", "folderId"},
		"properties": map[string]interface{}{
			"documentId": map[string]string{"type": "string"},
			"folderId":   map[string]string{"type": "string"},
		},
	}
}
func (t *moveFolderTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	_, err := loadEditableDocument(ctx, t.deps, user, input)
	if err != nil {
		return err
	}
	if _, err := parseRequiredString(input, "folderId"); err != nil {
		return err
	}
	return nil
}
func (t *moveFolderTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	document, err := loadEditableDocument(ctx, t.deps, user, input)
	if err != nil {
		return nil, err
	}
	folderID, err := parseRequiredString(input, "folderId")
	if err != nil {
		return nil, err
	}
	response, err := t.deps.Documents.Update(ctx, document.ID.String(), documents.UpdateDocumentRequest{
		OriginalFileName: document.OriginalFileName,
		FolderID:         &folderID,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"documentId": response.ID.String(), "folderId": folderID}, nil
}

type createImportantDateTool struct{ deps Dependencies }

func (t *createImportantDateTool) Name() string { return ToolCreateImportantDate }
func (t *createImportantDateTool) Description() string {
	return "Belge için önemli tarih oluşturur."
}
func (t *createImportantDateTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId", "title", "date"},
		"properties": map[string]interface{}{
			"documentId":         map[string]string{"type": "string"},
			"title":              map[string]string{"type": "string"},
			"date":               map[string]string{"type": "string"},
			"reminderDaysBefore": map[string]string{"type": "integer"},
		},
	}
}
func (t *createImportantDateTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	_, err := loadEditableDocument(ctx, t.deps, user, input)
	if err != nil {
		return err
	}
	if _, err := parseRequiredString(input, "title"); err != nil {
		return err
	}
	if _, err := parseRequiredString(input, "date"); err != nil {
		return err
	}
	return nil
}
func (t *createImportantDateTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	title, err := parseRequiredString(input, "title")
	if err != nil {
		return nil, err
	}
	date, err := parseRequiredString(input, "date")
	if err != nil {
		return nil, err
	}
	response, err := t.deps.Documents.CreateImportantDate(ctx, documentID.String(), documents.CreateDocumentImportantDateRequest{
		Title:              title,
		Date:               date,
		ReminderDaysBefore: parseOptionalInt(input, "reminderDaysBefore"),
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"importantDateId": response.ID.String(), "documentId": documentID.String()}, nil
}

type createWorkflowTool struct{ deps Dependencies }

func (t *createWorkflowTool) Name() string { return ToolCreateWorkflow }
func (t *createWorkflowTool) Description() string {
	return "Belge için onay süreci başlatır."
}
func (t *createWorkflowTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId", "title", "steps"},
		"properties": map[string]interface{}{
			"documentId": map[string]string{"type": "string"},
			"title":      map[string]string{"type": "string"},
			"steps":      map[string]string{"type": "array"},
		},
	}
}
func (t *createWorkflowTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	_, err := loadViewableDocument(ctx, t.deps, user, input)
	if err != nil {
		return err
	}
	if _, err := parseRequiredString(input, "title"); err != nil {
		return err
	}
	if _, err := parseWorkflowSteps(input); err != nil {
		return err
	}
	if t.deps.Approvals == nil {
		return ErrToolNotAllowed
	}
	return nil
}
func (t *createWorkflowTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	title, err := parseRequiredString(input, "title")
	if err != nil {
		return nil, err
	}
	steps, err := parseWorkflowSteps(input)
	if err != nil {
		return nil, err
	}
	response, err := t.deps.Approvals.CreateForDocument(ctx, documentID.String(), approvals.CreateWorkflowRequest{
		Title: title,
		Steps: steps,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"workflowId": response.ID.String(), "documentId": documentID.String()}, nil
}

type createRelationshipTool struct{ deps Dependencies }

func (t *createRelationshipTool) Name() string { return ToolCreateDocumentRelationship }
func (t *createRelationshipTool) Description() string {
	return "İki belge arasında ilişki oluşturur."
}
func (t *createRelationshipTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId", "targetDocumentId"},
		"properties": map[string]interface{}{
			"documentId":       map[string]string{"type": "string"},
			"targetDocumentId": map[string]string{"type": "string"},
			"relationshipType": map[string]string{"type": "string"},
		},
	}
}
func (t *createRelationshipTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	_, err := loadViewableDocument(ctx, t.deps, user, input)
	if err != nil {
		return err
	}
	if _, err := parseRequiredString(input, "targetDocumentId"); err != nil {
		return err
	}
	return nil
}
func (t *createRelationshipTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	targetID, err := parseRequiredString(input, "targetDocumentId")
	if err != nil {
		return nil, err
	}
	relationshipType := parseOptionalString(input, "relationshipType")
	if relationshipType == "" {
		relationshipType = documents.RelationshipTypeRelated
	}
	edge, err := t.deps.Documents.CreateRelationship(ctx, documentID.String(), documents.CreateDocumentRelationshipRequest{
		TargetDocumentID: targetID,
		RelationshipType: relationshipType,
	})
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"relationshipId": edge.ID.String(), "documentId": documentID.String()}, nil
}

type generateSummaryTool struct{ deps Dependencies }

func (t *generateSummaryTool) Name() string { return ToolGenerateSummary }
func (t *generateSummaryTool) Description() string { return "Belge özeti oluşturur." }
func (t *generateSummaryTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId"},
		"properties": map[string]interface{}{
			"documentId": map[string]string{"type": "string"},
		},
	}
}
func (t *generateSummaryTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	_, err := loadViewableDocument(ctx, t.deps, user, input)
	return err
}
func (t *generateSummaryTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	_, err = t.deps.Documents.GenerateSummary(ctx, documentID.String())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"documentId": documentID.String(), "summaryGenerated": true}, nil
}

type reprocessDocumentTool struct{ deps Dependencies }

func (t *reprocessDocumentTool) Name() string { return ToolReprocessDocument }
func (t *reprocessDocumentTool) Description() string { return "Belge metin çıkarma ve işleme sürecini yeniden başlatır." }
func (t *reprocessDocumentTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId"},
		"properties": map[string]interface{}{
			"documentId": map[string]string{"type": "string"},
		},
	}
}
func (t *reprocessDocumentTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	_, err := loadEditableDocument(ctx, t.deps, user, input)
	return err
}
func (t *reprocessDocumentTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	response, err := t.deps.Documents.ReprocessText(ctx, documentID.String())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"documentId": response.ID.String(), "status": string(response.Status)}, nil
}

type favoriteDocumentTool struct{ deps Dependencies }

func (t *favoriteDocumentTool) Name() string { return ToolFavoriteDocument }
func (t *favoriteDocumentTool) Description() string { return "Belgeyi favorilere ekler." }
func (t *favoriteDocumentTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"required": []string{"documentId"},
		"properties": map[string]interface{}{
			"documentId": map[string]string{"type": "string"},
		},
	}
}
func (t *favoriteDocumentTool) CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error {
	_, err := loadViewableDocument(ctx, t.deps, user, input)
	return err
}
func (t *favoriteDocumentTool) Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	response, err := t.deps.Documents.Favorite(ctx, documentID.String())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"documentId": response.ID.String(), "isFavorite": response.IsFavorite}, nil
}

func loadViewableDocument(ctx context.Context, deps Dependencies, user auth.Context, input map[string]interface{}) (*documents.Document, error) {
	documentID, err := parseDocumentID(input)
	if err != nil {
		return nil, err
	}
	document, err := deps.DocumentReader.GetByID(ctx, user.TenantID, documentID)
	if err != nil {
		return nil, fmt.Errorf("%w: document not found", ErrInvalidToolInput)
	}
	allowed, err := deps.Access.CanView(ctx, user, document)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrToolNotAllowed
	}
	return document, nil
}

func loadEditableDocument(ctx context.Context, deps Dependencies, user auth.Context, input map[string]interface{}) (*documents.Document, error) {
	document, err := loadViewableDocument(ctx, deps, user, input)
	if err != nil {
		return nil, err
	}
	allowed, err := deps.Access.CanEdit(ctx, user, document)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrToolNotAllowed
	}
	return document, nil
}

func parseWorkflowSteps(input map[string]interface{}) ([]approvals.CreateWorkflowStepInput, error) {
	rawSteps, ok := input["steps"].([]interface{})
	if !ok || len(rawSteps) == 0 {
		return nil, fmt.Errorf("%w: steps required", ErrInvalidToolInput)
	}
	steps := make([]approvals.CreateWorkflowStepInput, 0, len(rawSteps))
	for index, raw := range rawSteps {
		stepMap, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: invalid step", ErrInvalidToolInput)
		}
		approverRaw, ok := stepMap["approverUserId"].(string)
		if !ok || strings.TrimSpace(approverRaw) == "" {
			return nil, fmt.Errorf("%w: approverUserId required", ErrInvalidToolInput)
		}
		approverID, err := uuid.Parse(strings.TrimSpace(approverRaw))
		if err != nil {
			return nil, fmt.Errorf("%w: invalid approverUserId", ErrInvalidToolInput)
		}
		order := index + 1
		if value := parseOptionalInt(stepMap, "stepOrder"); value != nil {
			order = *value
		}
		steps = append(steps, approvals.CreateWorkflowStepInput{
			ApproverUserID: approverID,
			StepOrder:      order,
		})
	}
	return steps, nil
}

func reminderDateFromSuggestion(suggestedDate string, daysBefore int) (string, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(suggestedDate))
	if err != nil {
		return "", err
	}
	reminderDate := parsed.AddDate(0, 0, -daysBefore)
	return reminderDate.Format("2006-01-02"), nil
}

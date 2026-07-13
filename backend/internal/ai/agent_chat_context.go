package ai

import (
	"context"
	"fmt"
	"strings"

	"fifriz/internal/approvals"
	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/folders"
	"fifriz/internal/permissions"
	"fifriz/internal/tags"

	"github.com/google/uuid"
)

const (
	agentContextFolderLimit = 30
	agentContextTagLimit    = 30
)

type agentContextBuilder struct {
	documents *documents.Repository
	folders   *folders.Repository
	tags      *tags.Repository
	approvals *approvals.Repository
}

func (b *agentContextBuilder) build(
	ctx context.Context,
	authContext auth.Context,
	request AgentChatRequest,
) (string, error) {
	var sections []string

	sections = append(sections, fmt.Sprintf("Kullanıcı rolü: %s", authContext.Role))
	sections = append(sections, fmt.Sprintf("Tenant ID: %s", authContext.TenantID))

	if request.DocumentID != nil && strings.TrimSpace(*request.DocumentID) != "" {
		if section, err := b.buildDocumentSection(ctx, authContext, strings.TrimSpace(*request.DocumentID)); err != nil {
			return "", err
		} else if section != "" {
			sections = append(sections, section)
		}
	}

	if request.FolderID != nil && strings.TrimSpace(*request.FolderID) != "" {
		if section, err := b.buildFolderSection(ctx, authContext, strings.TrimSpace(*request.FolderID)); err != nil {
			return "", err
		} else if section != "" {
			sections = append(sections, section)
		}
	}

	if section, err := b.buildFoldersSection(ctx, authContext.TenantID); err != nil {
		return "", err
	} else if section != "" {
		sections = append(sections, section)
	}

	if section, err := b.buildTagsSection(ctx, authContext.TenantID); err != nil {
		return "", err
	} else if section != "" {
		sections = append(sections, section)
	}

	if section, err := b.buildPendingApprovalsSection(ctx, authContext); err != nil {
		return "", err
	} else if section != "" {
		sections = append(sections, section)
	}

	return strings.Join(sections, "\n\n"), nil
}

func (b *agentContextBuilder) buildDocumentSection(ctx context.Context, authContext auth.Context, documentIDRaw string) (string, error) {
	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return "", ErrInvalidDocumentFilter
	}
	document, err := b.documents.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		return "", err
	}

	lines := []string{
		"AKTİF BELGE:",
		fmt.Sprintf("- documentId: %s", document.ID),
		fmt.Sprintf("- dosya adı: %s", document.OriginalFileName),
	}
	if document.AIDocumentType != nil && strings.TrimSpace(*document.AIDocumentType) != "" {
		lines = append(lines, fmt.Sprintf("- sınıflandırma: %s", strings.TrimSpace(*document.AIDocumentType)))
	}
	if document.FolderID != nil {
		if folder, err := b.folders.GetByID(ctx, authContext.TenantID, *document.FolderID); err == nil {
			lines = append(lines, fmt.Sprintf("- mevcut klasör: %s (folderId: %s)", folder.Name, folder.ID))
		}
	}

	if b.tags != nil {
		tagMap, err := b.tags.ListTagsByDocumentIDs(ctx, authContext.TenantID, []uuid.UUID{document.ID})
		if err == nil {
			if tagsForDoc, ok := tagMap[document.ID]; ok && len(tagsForDoc) > 0 {
				names := make([]string, 0, len(tagsForDoc))
				for _, tag := range tagsForDoc {
					names = append(names, tag.Name)
				}
				lines = append(lines, fmt.Sprintf("- etiketler: %s", strings.Join(names, ", ")))
			}
		}
	}

	return strings.Join(lines, "\n"), nil
}

func (b *agentContextBuilder) buildFolderSection(ctx context.Context, authContext auth.Context, folderIDRaw string) (string, error) {
	folderID, err := uuid.Parse(folderIDRaw)
	if err != nil {
		return "", ErrInvalidFolderFilter
	}
	folder, err := b.folders.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("AKTİF KLASÖR:\n- folderId: %s\n- ad: %s", folder.ID, folder.Name), nil
}

func (b *agentContextBuilder) buildFoldersSection(ctx context.Context, tenantID uuid.UUID) (string, error) {
	if b.folders == nil {
		return "", nil
	}
	items, err := b.folders.List(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "KLASÖRLER: (yok)", nil
	}
	if len(items) > agentContextFolderLimit {
		items = items[:agentContextFolderLimit]
	}
	lines := []string{"KLASÖRLER:"}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("- %s (folderId: %s)", item.Name, item.ID))
	}
	return strings.Join(lines, "\n"), nil
}

func (b *agentContextBuilder) buildTagsSection(ctx context.Context, tenantID uuid.UUID) (string, error) {
	if b.tags == nil {
		return "", nil
	}
	items, err := b.tags.List(ctx, tenantID)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "ETİKETLER: (yok)", nil
	}
	if len(items) > agentContextTagLimit {
		items = items[:agentContextTagLimit]
	}
	lines := []string{"ETİKETLER:"}
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("- %s (tagId: %s)", item.Name, item.ID))
	}
	return strings.Join(lines, "\n"), nil
}

func (b *agentContextBuilder) buildPendingApprovalsSection(ctx context.Context, authContext auth.Context) (string, error) {
	if b.approvals == nil {
		return "", nil
	}
	pending, err := b.approvals.ListPendingForApprover(ctx, authContext.TenantID, authContext.UserID)
	if err != nil {
		return "", err
	}
	if len(pending) == 0 {
		return "ONAY BEKLEYEN BELGELER: (yok)", nil
	}
	lines := []string{fmt.Sprintf("ONAY BEKLEYEN BELGELER (%d):", len(pending))}
	limit := 8
	if len(pending) < limit {
		limit = len(pending)
	}
	for _, item := range pending[:limit] {
		lines = append(lines, fmt.Sprintf("- %s (documentId: %s, workflowId: %s)", item.DocumentName, item.DocumentID, item.WorkflowID))
	}
	return strings.Join(lines, "\n"), nil
}

func memberAccessUserIDForAgent(authContext auth.Context) *uuid.UUID {
	if permissions.HasTenantWideAccess(permissions.Role(authContext.Role)) {
		return nil
	}
	userID := authContext.UserID
	return &userID
}

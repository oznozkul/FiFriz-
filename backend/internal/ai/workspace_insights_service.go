package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"fifriz/internal/approvals"
	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
)

var classificationTypeLabels = map[string]string{
	"contract":         "sözleşme",
	"invoice":          "fatura",
	"insurance_policy": "sigorta poliçesi",
}

type WorkspaceInsightsService struct {
	repository *WorkspaceInsightsRepository
	approvals  *approvals.Repository
	cache      *cache.Service
}

func NewWorkspaceInsightsService(
	repository *WorkspaceInsightsRepository,
	approvalRepository *approvals.Repository,
	cacheService *cache.Service,
) *WorkspaceInsightsService {
	return &WorkspaceInsightsService{
		repository: repository,
		approvals:  approvalRepository,
		cache:      cacheService,
	}
}

func (s *WorkspaceInsightsService) List(ctx context.Context) ([]WorkspaceInsight, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	return cache.GetOrSetJSON(ctx, s.cache, cache.WorkspaceInsightsKey(authContext.TenantID, authContext.UserID), cache.TTLWorkspaceInsights, func() ([]WorkspaceInsight, error) {
		return s.buildInsights(ctx, authContext)
	})
}

func (s *WorkspaceInsightsService) buildInsights(ctx context.Context, authContext auth.Context) ([]WorkspaceInsight, error) {
	tenantWide := permissions.HasTenantWideAccess(permissions.Role(authContext.Role))
	insights := make([]WorkspaceInsight, 0, 10)

	if insight, ok, err := s.buildUpcomingDatesInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildPendingApprovalsInsight(ctx, authContext); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildRecentImportantInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildClassifiedDocumentsInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildMissingMetadataInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildFailedProcessingInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildRelatedSuggestionsInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildPopularTagsInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	if insight, ok, err := s.buildActiveFoldersInsight(ctx, authContext.TenantID, authContext.UserID, tenantWide); err != nil {
		return nil, err
	} else if ok {
		insights = append(insights, insight)
	}

	sort.SliceStable(insights, func(i, j int) bool {
		return insightPriorityRank(insights[i].Priority) < insightPriorityRank(insights[j].Priority)
	})

	if insights == nil {
		insights = []WorkspaceInsight{}
	}
	return insights, nil
}

func insightPriorityRank(priority string) int {
	switch priority {
	case InsightPriorityHigh:
		return 0
	case InsightPriorityMedium:
		return 1
	default:
		return 2
	}
}

func (s *WorkspaceInsightsService) buildUpcomingDatesInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	contractCount, err := s.repository.CountContractUpcomingDates(ctx, tenantID, userID, tenantWide, upcomingInsightDays)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}

	totalCount, err := s.repository.CountUpcomingImportantDates(ctx, tenantID, userID, tenantWide, upcomingInsightDays)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if totalCount == 0 {
		return WorkspaceInsight{}, false, nil
	}

	docs, err := s.repository.ListUpcomingImportantDateDocuments(ctx, tenantID, userID, tenantWide, upcomingInsightDays, insightDocumentLimit)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}

	description := fmt.Sprintf("%d belgede önemli tarih %d gün içinde.", totalCount, upcomingInsightDays)
	title := "Yakında tarihi gelen belgeler"
	if contractCount > 0 {
		title = fmt.Sprintf("%d sözleşmenin süresi %d gün içinde doluyor", contractCount, upcomingInsightDays)
		if contractCount == 1 {
			title = fmt.Sprintf("1 sözleşmenin süresi %d gün içinde doluyor", upcomingInsightDays)
		}
		description = fmt.Sprintf("Sözleşme ve diğer belgelerde toplam %d yaklaşan tarih var.", totalCount)
	}

	return WorkspaceInsight{
		Title:            title,
		Description:      description,
		Type:             InsightTypeUpcomingDates,
		Priority:         InsightPriorityMedium,
		RelatedDocuments: toInsightDocuments(docs),
		ActionLabel:      "Önemli Tarihleri Gör",
		ActionURL:        "/important-dates",
	}, true, nil
}

func (s *WorkspaceInsightsService) buildPendingApprovalsInsight(ctx context.Context, authContext auth.Context) (WorkspaceInsight, bool, error) {
	pending, err := s.approvals.ListPendingForApprover(ctx, authContext.TenantID, authContext.UserID)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if len(pending) == 0 {
		return WorkspaceInsight{}, false, nil
	}

	related := make([]WorkspaceInsightRelatedDocument, 0, len(pending))
	seen := make(map[uuid.UUID]struct{}, len(pending))
	for _, item := range pending {
		if _, exists := seen[item.DocumentID]; exists {
			continue
		}
		seen[item.DocumentID] = struct{}{}
		related = append(related, WorkspaceInsightRelatedDocument{
			DocumentID:       item.DocumentID,
			OriginalFileName: item.DocumentName,
		})
		if len(related) >= insightDocumentLimit {
			break
		}
	}

	title := fmt.Sprintf("%d onay süreci bekliyor", len(pending))
	if len(pending) == 1 {
		title = "1 onay süreci bekliyor"
	}

	return WorkspaceInsight{
		Title:            title,
		Description:      "Onayınızı bekleyen belge süreçlerini inceleyin.",
		Type:             InsightTypePendingApprovals,
		Priority:         InsightPriorityHigh,
		RelatedDocuments: related,
		ActionLabel:      "Onayları Gör",
		ActionURL:        "/approvals",
	}, true, nil
}

func (s *WorkspaceInsightsService) buildRecentImportantInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	docs, err := s.repository.ListRecentImportantDocuments(ctx, tenantID, userID, tenantWide, insightDocumentLimit)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if len(docs) == 0 {
		return WorkspaceInsight{}, false, nil
	}

	return WorkspaceInsight{
		Title:            "Son yüklenen önemli belgeler",
		Description:      fmt.Sprintf("Son %d günde %d önemli belge yüklendi.", recentImportantDays, len(docs)),
		Type:             InsightTypeRecentImportant,
		Priority:         InsightPriorityLow,
		RelatedDocuments: toInsightDocuments(docs),
		ActionLabel:      "Belgeleri Gör",
		ActionURL:        "/documents",
	}, true, nil
}

func (s *WorkspaceInsightsService) buildClassifiedDocumentsInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	counts, err := s.repository.ListClassificationCounts(ctx, tenantID, userID, tenantWide)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if len(counts) == 0 {
		return WorkspaceInsight{}, false, nil
	}

	parts := make([]string, 0, len(counts))
	for _, item := range counts {
		label := classificationTypeLabels[item.DocumentType]
		if label == "" {
			label = item.DocumentType
		}
		parts = append(parts, fmt.Sprintf("%d %s", item.Count, label))
	}

	actionURL := "/collections/contracts"
	if len(counts) == 1 {
		switch counts[0].DocumentType {
		case "invoice":
			actionURL = "/collections/invoices"
		case "insurance_policy":
			actionURL = "/collections/insurance_policies"
		}
	}

	return WorkspaceInsight{
		Title:            "AI sınıflandırılmış belgeler",
		Description:      fmt.Sprintf("Erişebildiğiniz belgeler arasında %s.", strings.Join(parts, ", ")),
		Type:             InsightTypeClassifiedDocuments,
		Priority:         InsightPriorityLow,
		RelatedDocuments: []WorkspaceInsightRelatedDocument{},
		ActionLabel:      "Koleksiyonları Gör",
		ActionURL:        actionURL,
	}, true, nil
}

func (s *WorkspaceInsightsService) buildMissingMetadataInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	count, err := s.repository.CountMissingMetadata(ctx, tenantID, userID, tenantWide)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if count == 0 {
		return WorkspaceInsight{}, false, nil
	}

	docs, err := s.repository.ListMissingMetadataDocuments(ctx, tenantID, userID, tenantWide, insightDocumentLimit)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}

	title := fmt.Sprintf("%d belge metadata için analiz bekliyor", count)
	if count == 1 {
		title = "1 belge metadata için analiz bekliyor"
	}

	actionURL := "/documents"
	if len(docs) > 0 {
		actionURL = fmt.Sprintf("/documents/%s", docs[0].DocumentID.String())
	}

	return WorkspaceInsight{
		Title:            title,
		Description:      "İşlenmiş ancak AI metadata çıkarımı tamamlanmamış belgeler.",
		Type:             InsightTypeMissingMetadata,
		Priority:         InsightPriorityMedium,
		RelatedDocuments: toInsightDocuments(docs),
		ActionLabel:      "Belgeyi Aç",
		ActionURL:        actionURL,
	}, true, nil
}

func (s *WorkspaceInsightsService) buildFailedProcessingInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	count, err := s.repository.CountFailedDocuments(ctx, tenantID, userID, tenantWide)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if count == 0 {
		return WorkspaceInsight{}, false, nil
	}

	docs, err := s.repository.ListFailedDocuments(ctx, tenantID, userID, tenantWide, insightDocumentLimit)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}

	title := fmt.Sprintf("%d belge işlenemedi", count)
	if count == 1 {
		title = "1 belge işlenemedi"
	}

	actionURL := "/documents"
	if len(docs) > 0 {
		actionURL = fmt.Sprintf("/documents/%s", docs[0].DocumentID.String())
	}

	return WorkspaceInsight{
		Title:            title,
		Description:      "Bu belgelerde metin çıkarma veya işleme adımı başarısız oldu.",
		Type:             InsightTypeFailedProcessing,
		Priority:         InsightPriorityHigh,
		RelatedDocuments: toInsightDocuments(docs),
		ActionLabel:      "Belgeyi İncele",
		ActionURL:        actionURL,
	}, true, nil
}

func (s *WorkspaceInsightsService) buildRelatedSuggestionsInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	count, err := s.repository.CountHighSimilaritySuggestions(ctx, tenantID, userID, tenantWide)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if count == 0 {
		return WorkspaceInsight{}, false, nil
	}

	pairs, err := s.repository.ListHighSimilaritySuggestions(ctx, tenantID, userID, tenantWide, insightDocumentLimit)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}

	related := make([]WorkspaceInsightRelatedDocument, 0, len(pairs)*2)
	seen := make(map[uuid.UUID]struct{})
	for _, pair := range pairs {
		appendInsightDoc := func(id uuid.UUID, name, ext string) {
			if _, exists := seen[id]; exists {
				return
			}
			seen[id] = struct{}{}
			related = append(related, WorkspaceInsightRelatedDocument{
				DocumentID:       id,
				OriginalFileName: name,
				FileExtension:    ext,
			})
		}
		appendInsightDoc(pair.SourceDocumentID, pair.SourceOriginalFileName, pair.SourceFileExtension)
		appendInsightDoc(pair.TargetDocumentID, pair.TargetOriginalFileName, pair.TargetFileExtension)
	}

	title := fmt.Sprintf("%d belge çifti birbirine çok benziyor", count)
	if count == 1 {
		title = "1 belge çifti birbirine çok benziyor"
	}
	if count == 2 {
		title = "2 belge birbirine çok benziyor"
	}

	actionURL := "/documents"
	if len(pairs) > 0 {
		actionURL = fmt.Sprintf("/documents/%s?panel=related", pairs[0].SourceDocumentID.String())
	}

	return WorkspaceInsight{
		Title:            title,
		Description:      "AI ilişkili belge önerileri incelemenizi bekliyor.",
		Type:             InsightTypeRelatedSuggestions,
		Priority:         InsightPriorityMedium,
		RelatedDocuments: related,
		ActionLabel:      "Önerileri Gör",
		ActionURL:        actionURL,
	}, true, nil
}

func (s *WorkspaceInsightsService) buildPopularTagsInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	tags, err := s.repository.ListTopTags(ctx, tenantID, userID, tenantWide, 5)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if len(tags) == 0 {
		return WorkspaceInsight{}, false, nil
	}

	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, fmt.Sprintf("%s (%d)", tag.TagName, tag.UsageCount))
	}

	return WorkspaceInsight{
		Title:            "Sık kullanılan etiketler",
		Description:      strings.Join(names, ", "),
		Type:             InsightTypePopularTags,
		Priority:         InsightPriorityLow,
		RelatedDocuments: []WorkspaceInsightRelatedDocument{},
		ActionLabel:      "Etiketleri Gör",
		ActionURL:        "/tags",
	}, true, nil
}

func (s *WorkspaceInsightsService) buildActiveFoldersInsight(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	tenantWide bool,
) (WorkspaceInsight, bool, error) {
	folders, err := s.repository.ListActiveFolders(ctx, tenantID, userID, tenantWide, 5)
	if err != nil {
		return WorkspaceInsight{}, false, err
	}
	if len(folders) == 0 {
		return WorkspaceInsight{}, false, nil
	}

	names := make([]string, 0, len(folders))
	for _, folder := range folders {
		names = append(names, fmt.Sprintf("%s (%d)", folder.FolderName, folder.DocumentCount))
	}

	return WorkspaceInsight{
		Title:            "En aktif klasörler",
		Description:      strings.Join(names, ", "),
		Type:             InsightTypeActiveFolders,
		Priority:         InsightPriorityLow,
		RelatedDocuments: []WorkspaceInsightRelatedDocument{},
		ActionLabel:      "Klasörleri Gör",
		ActionURL:        "/folders",
	}, true, nil
}

func toInsightDocuments(rows []insightDocumentRow) []WorkspaceInsightRelatedDocument {
	result := make([]WorkspaceInsightRelatedDocument, 0, len(rows))
	for _, row := range rows {
		result = append(result, WorkspaceInsightRelatedDocument{
			DocumentID:       row.DocumentID,
			OriginalFileName: row.OriginalFileName,
			FileExtension:    row.FileExtension,
		})
	}
	return result
}

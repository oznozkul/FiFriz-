package search

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/documents"
	"fifriz/internal/folders"
	"fifriz/internal/integrations"
	"fifriz/internal/sharing"
	"fifriz/internal/smartcollections"
	"fifriz/internal/tags"

	"github.com/google/uuid"
)

const globalResultLimit = 8

var ErrInvalidSearchQuery = errors.New("invalid search query")

type Service struct {
	repository     *Repository
	documents      *documents.Service
	folderRepo     *folders.Repository
	tagRepo        *tags.Repository
	authRepo       *auth.Repository
	integrationRepo *integrations.Repository
	sharingService *sharing.Service
	cache          *cache.Service
}

func NewService(
	repository *Repository,
	documentsService *documents.Service,
	folderRepo *folders.Repository,
	tagRepo *tags.Repository,
	authRepo *auth.Repository,
	integrationRepo *integrations.Repository,
	sharingService *sharing.Service,
	cacheService *cache.Service,
) *Service {
	return &Service{
		repository:      repository,
		documents:       documentsService,
		folderRepo:      folderRepo,
		tagRepo:         tagRepo,
		authRepo:        authRepo,
		integrationRepo: integrationRepo,
		sharingService:  sharingService,
		cache:           cacheService,
	}
}

func (s *Service) Global(ctx context.Context, query string) (*GlobalSearchResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	if query == "" {
		response := newEmptyGlobalSearchResponse()
		response.Actions = filterActions("")
		response.Settings = filterSettings("")
		return &response, nil
	}

	cacheKey := cache.GlobalSearchKey(authContext.TenantID, authContext.UserID, hashQuery(query))
	cached, err := cache.GetOrSetJSON(ctx, s.cache, cacheKey, cache.TTLGlobalSearch, func() (GlobalSearchResponse, error) {
		return s.searchUncached(ctx, authContext, query)
	})
	if err != nil {
		return nil, err
	}
	return &cached, nil
}

func (s *Service) searchUncached(ctx context.Context, authContext auth.Context, query string) (GlobalSearchResponse, error) {
	response := newEmptyGlobalSearchResponse()
	response.Actions = limitActions(filterActions(query), globalResultLimit)
	response.Settings = limitSettings(filterSettings(query), globalResultLimit)

	docResults, err := s.documents.SearchForGlobal(ctx, query)
	if err == nil {
		docIDs := make([]uuid.UUID, 0, len(docResults))
		for _, item := range docResults {
			docIDs = append(docIDs, item.DocumentID)
		}
		meta, metaErr := s.repository.DocumentFolderMeta(ctx, authContext.TenantID, docIDs)
		if metaErr != nil {
			meta = map[uuid.UUID]documentFolderMeta{}
		}
		for index, item := range docResults {
			if index >= globalResultLimit {
				break
			}
			row := meta[item.DocumentID]
			updatedAt := item.CreatedAt
			if !row.UpdatedAt.IsZero() {
				updatedAt = row.UpdatedAt
			}
			response.Documents = append(response.Documents, GlobalDocumentResult{
				DocumentID:       item.DocumentID,
				Title:            item.Title,
				OriginalFileName: item.OriginalFileName,
				FileExtension:    item.FileExtension,
				MatchedSnippet:   item.MatchedSnippet,
				FolderID:         row.FolderID,
				FolderName:       row.FolderName,
				UpdatedAt:        updatedAt,
			})
		}
	}

	folderCounts, _ := s.repository.FolderDocumentCounts(ctx, authContext.TenantID)
	folderItems, err := s.folderRepo.List(ctx, authContext.TenantID)
	if err == nil {
		for _, folder := range folderItems {
			if !matchesQuery(folder.Name, query) {
				continue
			}
			if len(response.Folders) >= globalResultLimit {
				break
			}
			response.Folders = append(response.Folders, GlobalFolderResult{
				ID:            folder.ID,
				Name:          folder.Name,
				DocumentCount: folderCounts[folder.ID],
			})
		}
	}

	tagCounts, _ := s.repository.TagUsageCounts(ctx, authContext.TenantID)
	tagItems, err := s.tagRepo.List(ctx, authContext.TenantID)
	aliasesByTag, _ := s.tagRepo.ListAllAliases(ctx, authContext.TenantID)
	if err == nil {
		for _, tagItem := range tagItems {
			if tagItem.Status == tags.StatusArchived || tagItem.Status == tags.StatusHidden {
				continue
			}
			nameMatch := matchesQuery(tagItem.Name, query)
			aliasMatch := false
			for _, alias := range aliasesByTag[tagItem.ID] {
				if matchesQuery(alias, query) {
					aliasMatch = true
					break
				}
			}
			if !nameMatch && !aliasMatch {
				continue
			}
			if len(response.Tags) >= globalResultLimit {
				break
			}
			response.Tags = append(response.Tags, GlobalTagResult{
				ID:         tagItem.ID,
				Name:       tagItem.Name,
				UsageCount: tagCounts[tagItem.ID],
			})
		}
	}

	for _, definition := range smartcollections.AllDefinitions() {
		if !matchesQuery(definition.Title, query) && !matchesQuery(definition.Description, query) {
			continue
		}
		if len(response.Collections) >= globalResultLimit {
			break
		}
		response.Collections = append(response.Collections, GlobalCollectionResult{
			Key:         definition.Key,
			Title:       definition.Title,
			Description: definition.Description,
			Icon:        definition.Icon,
		})
	}

	if s.sharingService != nil {
		sharedItems, err := s.sharingService.ListSharedWithMe(ctx)
		if err == nil {
			for _, item := range sharedItems {
				if !matchesQuery(item.DocumentName, query) && !matchesQuery(item.SharedByFullName, query) {
					continue
				}
				if len(response.SharedDocuments) >= globalResultLimit {
					break
				}
				response.SharedDocuments = append(response.SharedDocuments, GlobalSharedDocumentResult{
					ShareID:          item.ShareID,
					DocumentID:       item.DocumentID,
					DocumentName:     item.DocumentName,
					SharedByFullName: item.SharedByFullName,
					CreatedAt:        item.CreatedAt,
				})
			}
		}
	}

	if s.integrationRepo != nil {
		integrationItems, err := s.integrationRepo.ListByUser(ctx, authContext.TenantID, authContext.UserID)
		if err == nil {
			for _, item := range integrationItems {
				label := item.DisplayName + " " + item.Provider
				if item.EmailAddress != nil {
					label += " " + *item.EmailAddress
				}
				if !matchesQuery(label, query) {
					continue
				}
				if len(response.Integrations) >= globalResultLimit {
					break
				}
				response.Integrations = append(response.Integrations, GlobalIntegrationResult{
					ID:          item.ID,
					Provider:    item.Provider,
					DisplayName: item.DisplayName,
					Status:      item.Status,
				})
			}
		}
	}

	if authContext.Role == auth.TenantRoleOwner || authContext.Role == auth.TenantRoleAdmin {
		members, err := s.authRepo.ListMembersByTenantID(ctx, authContext.TenantID)
		if err == nil {
			for _, member := range members {
				label := member.User.FullName + " " + member.User.Email
				if !matchesQuery(label, query) {
					continue
				}
				if len(response.Users) >= globalResultLimit {
					break
				}
				response.Users = append(response.Users, GlobalUserResult{
					UserID:   member.UserID,
					FullName: member.User.FullName,
					Email:    member.User.Email,
					Role:     string(member.Role),
				})
			}
		}
	}

	return response, nil
}

func hashQuery(query string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(query))))
	return hex.EncodeToString(sum[:8])
}

func limitActions(items []GlobalActionResult, limit int) []GlobalActionResult {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitSettings(items []GlobalSettingResult, limit int) []GlobalSettingResult {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

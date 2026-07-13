package documents

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/embedding"
	domain "fifriz/internal/events/events"
	"fifriz/internal/tags"

	"github.com/google/uuid"
)

const hybridSearchResultLimit = 50

func (s *Service) HybridSearch(ctx context.Context, query string) ([]HybridSearchResultResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrInvalidSearchQuery
	}

	cacheKey := cache.HybridSearchKey(authContext.TenantID, authContext.UserID, hashHybridSearchQuery(query))
	results, err := cache.GetOrSetJSON(ctx, s.cache, cacheKey, cache.TTLHybridSearch, func() ([]HybridSearchResultResponse, error) {
		return s.hybridSearchUncached(ctx, authContext, query)
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Service) hybridSearchUncached(ctx context.Context, authContext auth.Context, query string) ([]HybridSearchResultResponse, error) {
	accessUserID := s.memberAccessUserID(authContext)
	candidates := make(map[uuid.UUID]*hybridSearchCandidate)
	terms := extractNormalizedSearchTerms(query)
	needle := strings.Join(terms, " ")

	fullTextQuery := buildFullTextSearchQuery(query)
	if fullTextQuery != "" {
		fullTextResults, err := s.repository.SearchTextRanked(ctx, authContext.TenantID, fullTextQuery, accessUserID)
		if err != nil {
			return nil, err
		}
		for _, item := range fullTextResults {
			candidate := ensureHybridCandidate(candidates, item.DocumentID)
			score := normalizeHybridScore(item.Score)
			if score > candidate.fullTextScore {
				candidate.fullTextScore = score
				candidate.matchedSnippet = formatSearchSnippet(item.FullText, query, item.MatchedSnippet)
			}
		}
	}

	prefixQuery := buildPrefixSearchQuery(query)
	if prefixQuery != "" && prefixQuery != fullTextQuery {
		prefixResults, err := s.repository.SearchTextRanked(ctx, authContext.TenantID, prefixQuery, accessUserID)
		if err != nil {
			return nil, err
		}
		for _, item := range prefixResults {
			candidate := ensureHybridCandidate(candidates, item.DocumentID)
			score := normalizeHybridScore(item.Score * 0.95)
			if score > candidate.fullTextScore {
				candidate.fullTextScore = score
				if candidate.matchedSnippet == "" {
					candidate.matchedSnippet = formatSearchSnippet(item.FullText, query, item.MatchedSnippet)
				}
			}
		}
	}

	if s.embeddingProvider != nil {
		embeddingValues, err := s.embeddingProvider.CreateEmbedding(ctx, query)
		if err == nil && len(embeddingValues) == embedding.DefaultEmbeddingDimensions {
			semanticDocs, err := s.repository.SearchDocumentsByEmbeddingGrouped(ctx, authContext.TenantID, embeddingValues, 30, accessUserID)
			if err != nil {
				return nil, err
			}
			for _, item := range semanticDocs {
				candidate := ensureHybridCandidate(candidates, item.DocumentID)
				score := normalizeHybridScore(item.Similarity)
				if score > candidate.semanticScore {
					candidate.semanticScore = score
				}
			}

			semanticChunks, err := s.repository.SearchSemanticChunks(ctx, authContext.TenantID, embeddingValues, 40, accessUserID)
			if err != nil {
				return nil, err
			}
			chunkMap := make(map[uuid.UUID][]string)
			for _, chunk := range semanticChunks {
				content := strings.TrimSpace(chunk.Content)
				if content == "" {
					continue
				}
				if len(chunkMap[chunk.DocumentID]) < 3 {
					chunkMap[chunk.DocumentID] = append(chunkMap[chunk.DocumentID], formatSnippet(content))
				}
				candidate := ensureHybridCandidate(candidates, chunk.DocumentID)
				score := normalizeHybridScore(chunk.Similarity)
				if score > candidate.semanticScore {
					candidate.semanticScore = score
				}
			}
			for documentID, chunks := range chunkMap {
				candidate := ensureHybridCandidate(candidates, documentID)
				candidate.matchedChunks = chunks
			}
		}
	}

	metadataQuery := buildPrefixSearchQuery(query)
	if metadataQuery == "" {
		metadataQuery = fullTextQuery
	}
	if metadataQuery != "" {
		metadataResults, err := s.repository.SearchMetadataRanked(ctx, authContext.TenantID, metadataQuery, accessUserID)
		if err != nil {
			return nil, err
		}
		for _, item := range metadataResults {
			candidate := ensureHybridCandidate(candidates, item.DocumentID)
			score := normalizeHybridScore(item.Score)
			if score > candidate.metadataScore {
				candidate.metadataScore = score
			}
		}
	}

	entityTerms := buildHybridEntitySearchPattern(query)
	if len(entityTerms) > 0 {
		entityResults, err := s.repository.SearchEntitiesRanked(ctx, authContext.TenantID, entityTerms, accessUserID)
		if err != nil {
			return nil, err
		}
		seenEntities := make(map[uuid.UUID]map[string]struct{})
		for _, item := range entityResults {
			candidate := ensureHybridCandidate(candidates, item.DocumentID)
			score := normalizeHybridScore(item.Score)
			if score > candidate.entityScore {
				candidate.entityScore = score
			}
			if seenEntities[item.DocumentID] == nil {
				seenEntities[item.DocumentID] = make(map[string]struct{})
			}
			key := entityDedupKey(item.EntityType, item.Value)
			if _, exists := seenEntities[item.DocumentID][key]; exists {
				continue
			}
			seenEntities[item.DocumentID][key] = struct{}{}
			candidate.matchedEntities = append(candidate.matchedEntities, HybridSearchMatchedEntity{
				EntityType: item.EntityType,
				Value:      item.Value,
				Label:      item.Label,
			})
		}
	}

	if len(candidates) == 0 {
		return []HybridSearchResultResponse{}, nil
	}

	documentIDs := make([]uuid.UUID, 0, len(candidates))
	for documentID := range candidates {
		documentIDs = append(documentIDs, documentID)
	}

	contextRows, err := s.repository.ListHybridSearchDocumentContexts(ctx, authContext.TenantID, documentIDs)
	if err != nil {
		return nil, err
	}
	contextMap := make(map[uuid.UUID]HybridSearchDocumentContext, len(contextRows))
	for _, row := range contextRows {
		contextMap[row.DocumentID] = row
	}

	tagMap, err := s.tags.MapTagsForDocuments(ctx, authContext.TenantID, documentIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for documentID, candidate := range candidates {
		contextRow, ok := contextMap[documentID]
		if !ok {
			delete(candidates, documentID)
			continue
		}

		candidate.filenameScore = fieldMatchScore(contextRow.OriginalFileName, needle)
		candidate.folderScore = fieldMatchScore(contextRow.FolderName, needle)
		if contextRow.AIDocumentType != nil {
			candidate.classificationScore = fieldMatchScore(*contextRow.AIDocumentType, needle)
		}

		for _, tag := range tagMap[documentID] {
			tagScore := fieldMatchScore(tag.Name, needle)
			if tagScore > candidate.tagScore {
				candidate.tagScore = tagScore
			}
		}

		if contextRow.IsFavorite {
			candidate.favoriteBoost = 0.03
		}
		if now.Sub(contextRow.CreatedAt) <= 30*24*time.Hour {
			candidate.recentBoost = 0.02
		}
	}

	ranked := make([]HybridSearchResultResponse, 0, len(candidates))
	for documentID, candidate := range candidates {
		contextRow := contextMap[documentID]
		finalScore := candidate.finalScore()
		if finalScore <= 0 {
			continue
		}
		tagsForDocument := tagMap[documentID]
		if tagsForDocument == nil {
			tagsForDocument = []tags.TagBrief{}
		}
		ranked = append(ranked, HybridSearchResultResponse{
			DocumentID:       documentID,
			OriginalFileName: contextRow.OriginalFileName,
			Title:            contextRow.OriginalFileName,
			FileExtension:    contextRow.FileExtension,
			FolderName:       contextRow.FolderName,
			Tags:             tagsForDocument,
			SourceType:       contextRow.SourceType,
			AIDocumentType:   contextRow.AIDocumentType,
			MatchedSnippet:   candidate.matchedSnippet,
			MatchedChunks:    candidate.matchedChunks,
			MatchedEntities:  candidate.matchedEntities,
			Score:            finalScore,
			MatchType:        candidate.matchType(),
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].OriginalFileName < ranked[j].OriginalFileName
		}
		return ranked[i].Score > ranked[j].Score
	})

	if len(ranked) > hybridSearchResultLimit {
		ranked = ranked[:hybridSearchResultLimit]
	}
	if ranked == nil {
		ranked = []HybridSearchResultResponse{}
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentSearched(authContext.TenantID, authContext.UserID, query, len(ranked)))
	return ranked, nil
}

func ensureHybridCandidate(candidates map[uuid.UUID]*hybridSearchCandidate, documentID uuid.UUID) *hybridSearchCandidate {
	candidate, ok := candidates[documentID]
	if ok {
		return candidate
	}
	candidate = &hybridSearchCandidate{documentID: documentID}
	candidates[documentID] = candidate
	return candidate
}

func hashHybridSearchQuery(query string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(query))))
	return hex.EncodeToString(sum[:8])
}

package documents

import (
	"time"

	"fifriz/internal/tags"

	"github.com/google/uuid"
)

type HybridSearchMatchedEntity struct {
	EntityType string  `json:"entityType"`
	Value      string  `json:"value"`
	Label      *string `json:"label,omitempty"`
}

type HybridSearchResultResponse struct {
	DocumentID       uuid.UUID                   `json:"documentId"`
	OriginalFileName string                      `json:"originalFileName"`
	Title            string                      `json:"title"`
	FileExtension    string                      `json:"fileExtension"`
	FolderName       string                      `json:"folderName"`
	Tags             []tags.TagBrief             `json:"tags"`
	SourceType       string                      `json:"sourceType"`
	AIDocumentType   *string                     `json:"aiDocumentType,omitempty"`
	MatchedSnippet   string                      `json:"matchedSnippet"`
	MatchedChunks    []string                    `json:"matchedChunks"`
	MatchedEntities  []HybridSearchMatchedEntity `json:"matchedEntities"`
	Score            float64                     `json:"score"`
	MatchType        string                      `json:"matchType"`
}

const (
	HybridMatchTypeFullText  = "full_text"
	HybridMatchTypeSemantic  = "semantic"
	HybridMatchTypeMetadata  = "metadata"
	HybridMatchTypeEntity    = "entity"
	HybridMatchTypeHybrid    = "hybrid"
)

type hybridSearchCandidate struct {
	documentID       uuid.UUID
	fullTextScore    float64
	semanticScore    float64
	metadataScore    float64
	entityScore      float64
	tagScore         float64
	folderScore      float64
	filenameScore    float64
	classificationScore float64
	favoriteBoost    float64
	recentBoost      float64
	matchedSnippet   string
	matchedChunks    []string
	matchedEntities  []HybridSearchMatchedEntity
	signalCount      int
}

func (c *hybridSearchCandidate) finalScore() float64 {
	score := c.fullTextScore*0.24 +
		c.semanticScore*0.24 +
		c.metadataScore*0.14 +
		c.entityScore*0.14 +
		c.tagScore*0.08 +
		c.folderScore*0.05 +
		c.filenameScore*0.05 +
		c.classificationScore*0.06 +
		c.favoriteBoost +
		c.recentBoost
	if score > 1 {
		return 1
	}
	return score
}

func (c *hybridSearchCandidate) matchType() string {
	active := 0
	var only string
	record := func(kind string, value float64) {
		if value <= 0 {
			return
		}
		active++
		if only == "" {
			only = kind
		}
	}
	record(HybridMatchTypeFullText, c.fullTextScore)
	record(HybridMatchTypeSemantic, c.semanticScore)
	record(HybridMatchTypeMetadata, c.metadataScore)
	record(HybridMatchTypeEntity, c.entityScore)
	if c.tagScore > 0 || c.folderScore > 0 || c.filenameScore > 0 || c.classificationScore > 0 {
		active++
		if only == "" {
			only = HybridMatchTypeHybrid
		}
	}
	if active > 1 {
		return HybridMatchTypeHybrid
	}
	if only != "" {
		return only
	}
	return HybridMatchTypeHybrid
}

type HybridSearchDocumentContext struct {
	DocumentID       uuid.UUID `gorm:"column:document_id"`
	OriginalFileName string    `gorm:"column:original_file_name"`
	FileExtension    string    `gorm:"column:file_extension"`
	FolderName       string    `gorm:"column:folder_name"`
	SourceType       string    `gorm:"column:source_type"`
	AIDocumentType   *string   `gorm:"column:ai_document_type"`
	IsFavorite       bool      `gorm:"column:is_favorite"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

type HybridSearchRankedResult struct {
	DocumentID     uuid.UUID `gorm:"column:document_id"`
	Score          float64   `gorm:"column:score"`
	MatchedSnippet string    `gorm:"column:matched_snippet"`
	FullText       string    `gorm:"column:full_text"`
}

type HybridSearchEntityHit struct {
	DocumentID uuid.UUID `gorm:"column:document_id"`
	Value      string    `gorm:"column:value"`
	Label      *string   `gorm:"column:label"`
	EntityType string    `gorm:"column:entity_type"`
	Score      float64   `gorm:"column:score"`
}

type HybridSearchSemanticChunk struct {
	DocumentID uuid.UUID `gorm:"column:document_id"`
	Content    string    `gorm:"column:content"`
	Similarity float64   `gorm:"column:similarity"`
}

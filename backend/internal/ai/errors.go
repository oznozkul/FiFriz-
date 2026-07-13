package ai

import (
	"errors"
)

var (
	ErrInvalidQuestion       = errors.New("question is required")
	ErrChatProviderDisabled  = errors.New("chat provider is disabled")
	ErrEmbeddingDisabled     = errors.New("embedding provider is disabled")
	ErrInvalidDocumentFilter = errors.New("invalid document id filter")
	ErrInvalidFolderFilter   = errors.New("invalid folder id filter")
	ErrInvalidTagFilter      = errors.New("invalid tag id filter")
	ErrInvalidCollectionKey  = errors.New("invalid collection key")
)

const (
	ragContextChunkLimit   = 8
	noContextAnswer        = "Bu konuda belgelerinizde yeterli bilgi bulamadım."
	contentPreviewMaxRunes = 240
)

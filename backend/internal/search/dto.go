package search

import (
	"time"

	"github.com/google/uuid"
)

type GlobalSearchResponse struct {
	Documents       []GlobalDocumentResult       `json:"documents"`
	Folders         []GlobalFolderResult         `json:"folders"`
	Tags            []GlobalTagResult            `json:"tags"`
	Collections     []GlobalCollectionResult     `json:"collections"`
	SharedDocuments []GlobalSharedDocumentResult `json:"sharedDocuments"`
	Integrations    []GlobalIntegrationResult    `json:"integrations"`
	Users           []GlobalUserResult           `json:"users"`
	Actions         []GlobalActionResult         `json:"actions"`
	Settings        []GlobalSettingResult        `json:"settings"`
}

type GlobalDocumentResult struct {
	DocumentID       uuid.UUID  `json:"documentId"`
	Title            string     `json:"title"`
	OriginalFileName string     `json:"originalFileName"`
	FileExtension    string     `json:"fileExtension"`
	MatchedSnippet   string     `json:"matchedSnippet"`
	FolderID         *uuid.UUID `json:"folderId,omitempty"`
	FolderName       string     `json:"folderName,omitempty"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type GlobalFolderResult struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	DocumentCount int       `json:"documentCount"`
}

type GlobalTagResult struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	UsageCount int       `json:"usageCount"`
}

type GlobalCollectionResult struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type GlobalSharedDocumentResult struct {
	ShareID          string    `json:"shareId"`
	DocumentID       string    `json:"documentId"`
	DocumentName     string    `json:"documentName"`
	SharedByFullName string    `json:"sharedByFullName"`
	CreatedAt        time.Time `json:"createdAt"`
}

type GlobalIntegrationResult struct {
	ID          uuid.UUID `json:"id"`
	Provider    string    `json:"provider"`
	DisplayName string    `json:"displayName"`
	Status      string    `json:"status"`
}

type GlobalUserResult struct {
	UserID   uuid.UUID `json:"userId"`
	FullName string    `json:"fullName"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
}

type GlobalActionResult struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path,omitempty"`
	Kind  string `json:"kind"`
}

type GlobalSettingResult struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

func newEmptyGlobalSearchResponse() GlobalSearchResponse {
	return GlobalSearchResponse{
		Documents:       []GlobalDocumentResult{},
		Folders:         []GlobalFolderResult{},
		Tags:            []GlobalTagResult{},
		Collections:     []GlobalCollectionResult{},
		SharedDocuments: []GlobalSharedDocumentResult{},
		Integrations:    []GlobalIntegrationResult{},
		Users:           []GlobalUserResult{},
		Actions:         []GlobalActionResult{},
		Settings:        []GlobalSettingResult{},
	}
}

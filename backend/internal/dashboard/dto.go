package dashboard

import (
	"time"

	"github.com/google/uuid"
)

type SummaryResponse struct {
	TotalDocuments             int64                `json:"totalDocuments"`
	TotalFolders               int64                `json:"totalFolders"`
	FavoriteDocuments          int64                `json:"favoriteDocuments"`
	TrashedDocuments           int64                `json:"trashedDocuments"`
	TotalStorageBytes          int64                `json:"totalStorageBytes"`
	DocumentsUploadedThisMonth int64                `json:"documentsUploadedThisMonth"`
	RecentDocuments            []RecentDocument     `json:"recentDocuments"`
	RecentActivities           []RecentActivity     `json:"recentActivities"`
	DocumentTypeStats          []DocumentTypeStat   `json:"documentTypeStats"`
	TopFolders                 []TopFolder          `json:"topFolders"`
	UpcomingImportantDates     int64                `json:"upcomingImportantDates"`
	OverdueImportantDates      int64                `json:"overdueImportantDates"`
	AssignedImportantDates     int64                `json:"assignedImportantDates"`
	UpcomingImportantDateItems []ImportantDateItem  `json:"upcomingImportantDateItems"`
}

type RecentDocument struct {
	ID               uuid.UUID `json:"id"`
	OriginalFileName string    `json:"originalFileName"`
	FileExtension    string    `json:"fileExtension"`
	FileSize         int64     `json:"fileSize"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	FolderName       *string   `json:"folderName"`
	IsFavorite       bool      `json:"isFavorite"`
}

type RecentActivity struct {
	ID          uuid.UUID `json:"id"`
	Action      string    `json:"action"`
	EntityType  string    `json:"entityType"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type DocumentTypeStat struct {
	Extension string `json:"extension"`
	Count     int64  `json:"count"`
}

type TopFolder struct {
	FolderID      uuid.UUID `json:"folderId"`
	FolderName    string    `json:"folderName"`
	DocumentCount int64     `json:"documentCount"`
}

type ImportantDateItem struct {
	ID               uuid.UUID  `json:"id"`
	DocumentID       uuid.UUID  `json:"documentId"`
	Title            string     `json:"title"`
	Date             string     `json:"date"`
	Type             *string    `json:"type"`
	IsCompleted      bool       `json:"isCompleted"`
	DocumentName     string     `json:"documentName"`
	AssignedToUserID *uuid.UUID `json:"assignedToUserId"`
}

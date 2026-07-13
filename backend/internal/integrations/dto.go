package integrations

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateIntegrationRequest struct {
	Provider     string                         `json:"provider"`
	DisplayName  string                         `json:"displayName"`
	EmailAddress *string                        `json:"emailAddress"`
	Settings     map[string]any                 `json:"settings"`
	Credentials  *IntegrationCredentialsRequest `json:"credentials"`
}

type UpdateIntegrationRequest struct {
	DisplayName *string                        `json:"displayName"`
	Settings    map[string]any                 `json:"settings"`
	IsActive    *bool                          `json:"isActive"`
	Credentials *IntegrationCredentialsRequest `json:"credentials"`
}

type IntegrationCredentialsRequest struct {
	Password string `json:"password"`
}

type IntegrationResponse struct {
	ID           uuid.UUID      `json:"id"`
	Provider     string         `json:"provider"`
	DisplayName  string         `json:"displayName"`
	EmailAddress *string        `json:"emailAddress"`
	Status       string         `json:"status"`
	Settings     map[string]any `json:"settings"`
	LastSyncAt   *time.Time     `json:"lastSyncAt"`
	LastError    *string        `json:"lastError"`
	IsActive          bool           `json:"isActive"`
	ImportedItemCount int64          `json:"importedItemCount"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

type SyncQueuedResponse struct {
	Message string `json:"message"`
}

type GmailConnectResponse struct {
	AuthURL string `json:"authUrl"`
}

type OutlookConnectResponse struct {
	AuthURL string `json:"authUrl"`
}

type GoogleDriveConnectResponse struct {
	AuthURL string `json:"authUrl"`
}

type SyncLogResponse struct {
	ID             uuid.UUID      `json:"id"`
	SyncType       string         `json:"syncType"`
	Status         string         `json:"status"`
	StartedAt      *time.Time     `json:"startedAt"`
	FinishedAt     *time.Time     `json:"finishedAt"`
	TotalItems     int            `json:"totalItems"`
	ImportedItems  int            `json:"importedItems"`
	SkippedItems   int            `json:"skippedItems"`
	FailedItems    int            `json:"failedItems"`
	ErrorMessage   *string        `json:"errorMessage"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      time.Time      `json:"createdAt"`
}

func toIntegrationResponse(account IntegrationAccount) IntegrationResponse {
	return IntegrationResponse{
		ID:           account.ID,
		Provider:     account.Provider,
		DisplayName:  account.DisplayName,
		EmailAddress: account.EmailAddress,
		Status:       account.Status,
		Settings:     decodeJSONMap(account.Settings),
		LastSyncAt:   account.LastSyncAt,
		LastError:    account.LastError,
		IsActive:     account.IsActive,
		CreatedAt:    account.CreatedAt,
		UpdatedAt:    account.UpdatedAt,
	}
}

func toIntegrationResponseWithCount(account IntegrationAccount, importedItemCount int64) IntegrationResponse {
	response := toIntegrationResponse(account)
	response.ImportedItemCount = importedItemCount
	return response
}

func toSyncLogResponse(log IntegrationSyncLog) SyncLogResponse {
	return SyncLogResponse{
		ID:            log.ID,
		SyncType:      log.SyncType,
		Status:        log.Status,
		StartedAt:     log.StartedAt,
		FinishedAt:    log.FinishedAt,
		TotalItems:    log.TotalItems,
		ImportedItems: log.ImportedItems,
		SkippedItems:  log.SkippedItems,
		FailedItems:   log.FailedItems,
		ErrorMessage:  log.ErrorMessage,
		Metadata:      decodeJSONMap(log.Metadata),
		CreatedAt:     log.CreatedAt,
	}
}

func encodeJSONMap(value map[string]any) ([]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return json.Marshal(value)
}

func decodeJSONMap(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	return result
}

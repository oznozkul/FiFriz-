package navigation

import "time"

type CreatePinRequest struct {
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Label      string `json:"label"`
	Path       string `json:"path"`
}

type PinResponse struct {
	ID         string    `json:"id"`
	EntityType string    `json:"entityType"`
	EntityID   string    `json:"entityId"`
	Label      string    `json:"label"`
	Path       string    `json:"path"`
	CreatedAt  time.Time `json:"createdAt"`
}

type RecentResponse struct {
	Documents        []RecentItem `json:"documents"`
	Folders          []RecentItem `json:"folders"`
	SmartCollections []RecentItem `json:"smartCollections"`
	Integrations     []RecentItem `json:"integrations"`
	SettingsPages    []RecentItem `json:"settingsPages"`
}

type TrackVisitRequest struct {
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Label      string `json:"label"`
	Path       string `json:"path"`
}

package sharing

import "time"

type CreateInternalShareRequest struct {
	SharedWithUserID string                  `json:"sharedWithUserId"`
	Permission       InternalSharePermission `json:"permission"`
}

type UpdateInternalShareRequest struct {
	Permission InternalSharePermission `json:"permission"`
}

type InternalShareResponse struct {
	ID                 string                  `json:"id"`
	DocumentID         string                  `json:"documentId"`
	SharedWithUserID   string                  `json:"sharedWithUserId"`
	SharedWithFullName string                  `json:"sharedWithFullName"`
	SharedWithEmail    string                  `json:"sharedWithEmail"`
	SharedByUserID     string                  `json:"sharedByUserId"`
	SharedByFullName   string                  `json:"sharedByFullName"`
	Permission         InternalSharePermission `json:"permission"`
	CreatedAt          time.Time               `json:"createdAt"`
	UpdatedAt          time.Time               `json:"updatedAt"`
}

type SharedWithMeResponse struct {
	ShareID          string                  `json:"shareId"`
	DocumentID       string                  `json:"documentId"`
	DocumentName     string                  `json:"documentName"`
	SharedByUserID   string                  `json:"sharedByUserId"`
	SharedByFullName string                  `json:"sharedByFullName"`
	Permission       InternalSharePermission `json:"permission"`
	CreatedAt        time.Time               `json:"createdAt"`
}

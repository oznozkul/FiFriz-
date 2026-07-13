package documents

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RelationshipTypeRelated     = "related"
	RelationshipTypeAttachment  = "attachment"
	RelationshipTypeParent      = "parent"
	RelationshipTypeChild       = "child"
	RelationshipTypeReplacement = "replacement"
	RelationshipTypeReference   = "reference"
	RelationshipTypeInvoiceOf   = "invoice_of"
	RelationshipTypeContractOf  = "contract_of"
	RelationshipTypeIdentityOf  = "identity_of"
	RelationshipTypeCustom      = "custom"
)

var AllowedRelationshipTypes = map[string]struct{}{
	RelationshipTypeRelated:     {},
	RelationshipTypeAttachment:  {},
	RelationshipTypeParent:      {},
	RelationshipTypeChild:       {},
	RelationshipTypeReplacement: {},
	RelationshipTypeReference:   {},
	RelationshipTypeInvoiceOf:   {},
	RelationshipTypeContractOf:  {},
	RelationshipTypeIdentityOf:  {},
	RelationshipTypeCustom:      {},
}

type DocumentRelationship struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID         uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	SourceDocumentID uuid.UUID `gorm:"type:uuid;not null;index" json:"source_document_id"`
	TargetDocumentID uuid.UUID `gorm:"type:uuid;not null;index" json:"target_document_id"`
	RelationshipType string    `gorm:"type:text;not null" json:"relationship_type"`
	CreatedByUserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"created_by_user_id"`
	CreatedAt        time.Time `json:"created_at"`
}

func (r *DocumentRelationship) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (DocumentRelationship) TableName() string {
	return "document_relationships"
}

func IsAllowedRelationshipType(value string) bool {
	_, ok := AllowedRelationshipTypes[value]
	return ok
}

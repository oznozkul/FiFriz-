package folders

import (
	"context"

	"github.com/google/uuid"
)

var personalDefaultCategories = []string{
	"Kira & Ev",
	"Faturalar",
	"Garanti Belgeleri",
	"Araç Belgeleri",
	"Sağlık Belgeleri",
	"Sigorta",
	"Banka & Finans",
	"Kimlik & Pasaport",
	"Eğitim & Sertifikalar",
	"Hukuki Belgeler",
	"Diğer",
}

func (s *Service) BootstrapPersonalTenant(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.ensurePersonalDefaults(ctx, tenantID, userID)
}

func (s *Service) ensurePersonalDefaults(ctx context.Context, tenantID, userID uuid.UUID) error {
	if s == nil || s.repository == nil {
		return nil
	}

	seeded, err := s.repository.HasSystemFolders(ctx, tenantID)
	if err != nil {
		return err
	}
	if seeded {
		return nil
	}

	folders := make([]Folder, 0, len(personalDefaultCategories))
	for _, name := range personalDefaultCategories {
		folders = append(folders, Folder{
			TenantID:        tenantID,
			ParentID:        nil,
			Name:            name,
			CreatedByUserID: userID,
			IsSystem:        true,
		})
	}

	return s.repository.CreateBatch(ctx, folders)
}

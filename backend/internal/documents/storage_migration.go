package documents

import (
	"context"
	"errors"

	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/permissions"
	"fifriz/internal/response"
	"fifriz/internal/storage"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StorageMigrationService struct {
	storage *storage.Manager
	repo    *Repository
	audit   audit.Recorder
	appEnv  string
}

func NewStorageMigrationService(storageManager *storage.Manager, repository *Repository, auditRecorder audit.Recorder, appEnv string) *StorageMigrationService {
	return &StorageMigrationService{
		storage: storageManager,
		repo:    repository,
		audit:   auditRecorder,
		appEnv:  appEnv,
	}
}

type StorageMigrationReport struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

func (s *StorageMigrationService) MigrateLocalToMinIO(ctx context.Context) (*StorageMigrationReport, error) {
	if s.storage == nil || s.storage.MinIOProvider() == nil {
		return nil, errors.New("minio storage is not configured")
	}

	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !s.canRunMigration(authContext) {
		return nil, audit.ErrForbidden
	}

	items, err := s.repo.ListLocalStorageDocuments(ctx, authContext.TenantID)
	if err != nil {
		return nil, err
	}

	report := &StorageMigrationReport{Total: len(items)}
	for _, document := range items {
		location, err := documentStorageLocation(s.storage, &document)
		if err != nil {
			report.Failed++
			report.Errors = append(report.Errors, document.ID.String()+": "+err.Error())
			s.logMigrationFailure(ctx, authContext, document.ID, err)
			continue
		}

		target, err := s.storage.MigrateDocumentToMinIO(ctx, storage.MigrateDocumentInput{
			TenantID:             document.TenantID,
			DocumentID:           document.ID,
			StoredFileName:       document.StoredFileName,
			CurrentLocation:      location,
			IntegrationAccountID: document.SourceIntegrationAccountID,
		})
		if err != nil {
			report.Failed++
			report.Errors = append(report.Errors, document.ID.String()+": "+err.Error())
			s.logMigrationFailure(ctx, authContext, document.ID, err)
			continue
		}

		document.StorageProvider = storage.ProviderMinIO
		document.StorageKey = target.Key
		document.BucketName = target.BucketName
		document.FilePath = ""
		if err := s.repo.Update(ctx, &document); err != nil {
			report.Failed++
			report.Errors = append(report.Errors, document.ID.String()+": "+err.Error())
			s.logMigrationFailure(ctx, authContext, document.ID, err)
			continue
		}

		report.Success++
		s.logMigrationSuccess(ctx, authContext, document.ID, target.Key)
	}

	return report, nil
}

func (s *StorageMigrationService) canRunMigration(authContext auth.Context) bool {
	if permissions.CanViewAuditLogs(permissions.Role(authContext.Role)) {
		return true
	}
	return s.appEnv == "development"
}

func (s *StorageMigrationService) logMigrationSuccess(ctx context.Context, authContext auth.Context, documentID uuid.UUID, storageKey string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.CreateLog(ctx, audit.CreateLogInput{
		TenantID:   authContext.TenantID,
		UserID:     &authContext.UserID,
		Action:     audit.ActionDocumentStorageMigrated,
		EntityType: audit.EntityTypeDocument,
		EntityID:   &documentID,
		Metadata: map[string]any{
			"storageKey": storageKey,
			"provider":   storage.ProviderMinIO,
		},
	})
}

func (s *StorageMigrationService) logMigrationFailure(ctx context.Context, authContext auth.Context, documentID uuid.UUID, migrateErr error) {
	if s.audit == nil {
		return
	}
	_ = s.audit.CreateLog(ctx, audit.CreateLogInput{
		TenantID:   authContext.TenantID,
		UserID:     &authContext.UserID,
		Action:     audit.ActionStorageMigrationFailed,
		EntityType: audit.EntityTypeDocument,
		EntityID:   &documentID,
		Metadata: map[string]any{
			"error": migrateErr.Error(),
		},
	})
}

type StorageMigrationHandler struct {
	service *StorageMigrationService
}

func NewStorageMigrationHandler(service *StorageMigrationService) *StorageMigrationHandler {
	return &StorageMigrationHandler{service: service}
}

func (h *StorageMigrationHandler) MigrateLocalToMinIO(c *fiber.Ctx) error {
	report, err := h.service.MigrateLocalToMinIO(c.UserContext())
	if err != nil {
		if errors.Is(err, audit.ErrForbidden) {
			return response.WriteForbidden(c, "forbidden", "you do not have permission for this action")
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Fail("storage_migration_failed", "could not migrate local documents to minio"))
	}
	return c.JSON(response.OK(report))
}

func RegisterStorageAdminRoutes(router fiber.Router, handler *StorageMigrationHandler) {
	router.Post("/admin/storage/migrate-local-to-minio", handler.MigrateLocalToMinIO)
}

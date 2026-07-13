package folders

import (
	"context"
	"errors"
	"strings"
	"time"

	"fifriz/internal/auth"
	"fifriz/internal/cache"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrFolderNotFound        = errors.New("folder not found")
	ErrInvalidFolderID       = errors.New("invalid folder id")
	ErrInvalidFolderData     = errors.New("invalid folder data")
	ErrFolderNameConflict    = errors.New("folder name conflict")
	ErrForbidden             = errors.New("forbidden")
	ErrSystemFolderProtected = errors.New("system folder cannot be modified")
)

type TenantTypeReader interface {
	TenantType(ctx context.Context, tenantID uuid.UUID) (auth.TenantType, error)
}

type Service struct {
	repository *Repository
	publisher  events.Publisher
	tenants    TenantTypeReader
	cache      *cache.Service
}

func NewService(repository *Repository, publisher events.Publisher, tenants TenantTypeReader, cacheService *cache.Service) *Service {
	return &Service{
		repository: repository,
		publisher:  publisher,
		tenants:    tenants,
		cache:      cacheService,
	}
}

func (s *Service) List(ctx context.Context) ([]FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	isAdmin := permissions.CanManageFolder(permissions.Role(authContext.Role))

	return cache.GetOrSetJSON(ctx, s.cache, cache.FolderTreeKey(authContext.TenantID), cache.TTLFolderTree, func() ([]FolderResponse, error) {
		if err := s.ensurePersonalDefaultsForTenant(ctx, authContext.TenantID, authContext.UserID); err != nil {
			return nil, err
		}

		items, err := s.repository.List(ctx, authContext.TenantID)
		if err != nil {
			return nil, err
		}

		return toFolderTree(items, isAdmin), nil
	})
}

func (s *Service) GetByID(ctx context.Context, id string) (*FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	folder, err := s.repository.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	if !IsFolderVisible(folder, permissions.CanManageFolder(permissions.Role(authContext.Role))) {
		return nil, ErrFolderNotFound
	}

	touchActivity(folder)
	_ = s.repository.Update(ctx, folder)

	response := toFolderResponse(*folder)
	s.publisher.Publish(ctx, domain.NewFolderOpened(
		authContext.TenantID,
		authContext.UserID,
		folder.ID,
		folder.Name,
		"/folders",
	))
	return &response, nil
}

func (s *Service) TrackOpened(ctx context.Context, id string) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return
	}
	folderID, err := uuid.Parse(id)
	if err != nil {
		return
	}
	folder, err := s.repository.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		return
	}
	touchActivity(folder)
	_ = s.repository.Update(ctx, folder)
	s.publisher.Publish(ctx, domain.NewFolderOpened(
		authContext.TenantID,
		authContext.UserID,
		folder.ID,
		folder.Name,
		"/folders",
	))
}

func (s *Service) Create(ctx context.Context, input CreateFolderRequest) (*FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrInvalidFolderData
	}

	parentID, err := s.parseParentID(ctx, authContext.TenantID, input.ParentID)
	if err != nil {
		return nil, err
	}
	if parentID != nil {
		parent, parentErr := s.repository.GetByID(ctx, authContext.TenantID, *parentID)
		if parentErr != nil {
			return nil, ErrFolderNotFound
		}
		if err := s.requireFolderUnlocked(parent); err != nil {
			return nil, err
		}
	}

	exists, err := s.repository.ExistsByName(ctx, authContext.TenantID, parentID, name, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrFolderNameConflict
	}

	now := folderNow()
	folder := &Folder{
		TenantID:             authContext.TenantID,
		ParentID:             parentID,
		Name:                 name,
		Description:          strings.TrimSpace(input.Description),
		Status:               StatusActive,
		InheritPermissions:   true,
		Icon:                 input.Icon,
		Color:                input.Color,
		CreatedByUserID:      authContext.UserID,
		LastModifiedByUserID: &authContext.UserID,
		LastActivityAt:       &now,
	}

	if err := s.repository.Create(ctx, folder); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewFolderCreated(authContext.TenantID, authContext.UserID, folder.ID, folder.Name))

	response := toFolderResponse(*folder)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateFolderRequest) (*FolderResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	if !permissions.CanManageFolder(permissions.Role(authContext.Role)) {
		return nil, ErrForbidden
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	folder, err := s.repository.GetByID(ctx, authContext.TenantID, folderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	if folder.IsSystem {
		return nil, ErrSystemFolderProtected
	}
	if err := s.requireFolderUnlocked(folder); err != nil {
		return nil, err
	}

	oldName := folder.Name
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrInvalidFolderData
	}

	exists, err := s.repository.ExistsByName(ctx, authContext.TenantID, folder.ParentID, name, &folder.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrFolderNameConflict
	}

	folder.Name = name
	folder.Description = strings.TrimSpace(input.Description)
	if strings.TrimSpace(input.Status) != "" {
		folder.Status = NormalizeStatus(input.Status)
	}
	if input.Icon != nil {
		folder.Icon = input.Icon
	}
	if input.Color != nil {
		folder.Color = input.Color
	}
	folder.LastModifiedByUserID = &authContext.UserID
	touchActivity(folder)

	if err := s.repository.Update(ctx, folder); err != nil {
		return nil, err
	}

	if oldName != folder.Name {
		s.publisher.PublishAfterCommit(ctx, domain.NewFolderRenamed(authContext.TenantID, authContext.UserID, folderID, oldName, folder.Name))
	} else {
		s.publisher.PublishAfterCommit(ctx, domain.NewFolderUpdated(authContext.TenantID, authContext.UserID, folderID, folder.Name))
	}

	response := toFolderResponse(*folder)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}
	if !permissions.CanManageFolder(permissions.Role(authContext.Role)) {
		return ErrForbidden
	}

	folderID, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidFolderID
	}

	allFolders, err := s.repository.List(ctx, authContext.TenantID)
	if err != nil {
		return err
	}

	var folder *Folder
	for index := range allFolders {
		if allFolders[index].ID == folderID {
			folder = &allFolders[index]
			break
		}
	}
	if folder == nil {
		return ErrFolderNotFound
	}
	if folder.IsSystem {
		return ErrSystemFolderProtected
	}
	if err := s.requireFolderUnlocked(folder); err != nil {
		return err
	}

	deleteIDs := CollectFolderAndDescendantIDs(allFolders, folderID)
	if err := s.repository.SoftDeleteIDs(ctx, authContext.TenantID, deleteIDs); err != nil {
		return err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewFolderDeleted(authContext.TenantID, authContext.UserID, folder.ID, folder.Name))
	return nil
}

func (s *Service) ensurePersonalDefaultsForTenant(ctx context.Context, tenantID, userID uuid.UUID) error {
	if s.tenants == nil {
		return nil
	}
	tenantType, err := s.tenants.TenantType(ctx, tenantID)
	if err != nil {
		return err
	}
	if !ShouldSeedDefaultFolders(tenantType) {
		return nil
	}
	return s.ensurePersonalDefaults(ctx, tenantID, userID)
}

func (s *Service) parseParentID(ctx context.Context, tenantID uuid.UUID, raw string) (*uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parentID, err := uuid.Parse(raw)
	if err != nil {
		return nil, ErrInvalidFolderID
	}

	parent, err := s.repository.GetByID(ctx, tenantID, parentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}
	if parent.TenantID != tenantID {
		return nil, ErrFolderMoveCrossTenant
	}

	return &parentID, nil
}

func folderNow() time.Time {
	return time.Now()
}

package sharing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"fifriz/internal/audit"
	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/events"
	"fifriz/internal/notifications"
	"fifriz/internal/permissions"
	"fifriz/internal/storage"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrShareLinkNotFound      = errors.New("share link not found")
	ErrInvalidShareLinkID     = errors.New("invalid share link id")
	ErrInvalidDocumentID      = errors.New("invalid document id")
	ErrDocumentNotFound       = errors.New("document not found")
	ErrForbidden              = errors.New("forbidden")
	ErrInvalidSharePermission = errors.New("invalid share permission")
	ErrInvalidShareChannel    = errors.New("invalid share channel")
	ErrShareLinkExpired       = errors.New("share link expired")
	ErrShareLinkInactive      = errors.New("share link inactive")
	ErrMaxViewsReached        = errors.New("max view count reached")
	ErrPasswordRequired       = errors.New("password required")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrDownloadNotAllowed     = errors.New("download not allowed")
	ErrInvalidShareToken      = errors.New("invalid share token")
	ErrInvalidShareType       = errors.New("invalid share type")
	ErrShareAccessDenied      = errors.New("share access denied")
)

type Service struct {
	repository         *Repository
	documentRepository *documents.Repository
	authRepository     *auth.Repository
	access             DocumentShareAccess
	storage            *storage.Manager
	publicAppURL       string
	publicAPIURL       string
	jwtSecret          string
	publisher          events.Publisher
	notifications      notifications.Creator
	auditRecorder      audit.Recorder
	downloadMaxRetries int
	downloadTimeout    time.Duration
}

func NewService(
	repository *Repository,
	documentRepository *documents.Repository,
	authRepository *auth.Repository,
	access DocumentShareAccess,
	storageManager *storage.Manager,
	publicAppURL string,
	publicAPIURL string,
	jwtSecret string,
	publisher events.Publisher,
	notificationCreator notifications.Creator,
) *Service {
	return &Service{
		repository:         repository,
		documentRepository: documentRepository,
		authRepository:     authRepository,
		access:             access,
		storage:            storageManager,
		publicAppURL:       strings.TrimRight(publicAppURL, "/"),
		publicAPIURL:       strings.TrimRight(publicAPIURL, "/"),
		jwtSecret:          jwtSecret,
		publisher:          publisher,
		notifications:      notificationCreator,
		downloadMaxRetries: 3,
		downloadTimeout:    30 * time.Second,
	}
}

func (s *Service) SetAuditRecorder(recorder audit.Recorder) {
	s.auditRecorder = recorder
}

func (s *Service) SetDownloadConfig(maxRetries int, timeout time.Duration) {
	if maxRetries > 0 {
		s.downloadMaxRetries = maxRetries
	}
	if timeout > 0 {
		s.downloadTimeout = timeout
	}
}

func (s *Service) CreateShareLink(ctx context.Context, documentID string, input CreateShareLinkRequest) (*ShareLinkResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if allowed, err := s.canManageShareLink(ctx, authContext, document); err != nil {
		return nil, err
	} else if !allowed {
		return nil, ErrForbidden
	}

	permission := SharePermission(strings.TrimSpace(string(input.Permission)))
	if permission != SharePermissionView && permission != SharePermissionDownload {
		return nil, ErrInvalidSharePermission
	}

	allowDownload := true
	if input.AllowDownload != nil {
		allowDownload = *input.AllowDownload
	}
	if permission == SharePermissionView && !allowDownload {
		allowDownload = false
	}
	if permission == SharePermissionDownload {
		allowDownload = true
	}

	var passwordHash *string
	hasPassword := false
	if input.Password != nil && strings.TrimSpace(*input.Password) != "" {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(*input.Password)), bcrypt.DefaultCost)
		if hashErr != nil {
			return nil, hashErr
		}
		value := string(hash)
		passwordHash = &value
		hasPassword = true
	}

	token, err := generateShareToken()
	if err != nil {
		return nil, err
	}

	link := &DocumentShareLink{
		TenantID:        authContext.TenantID,
		DocumentID:      document.ID,
		CreatedByUserID: authContext.UserID,
		Token:           token,
		Permission:      permission,
		PasswordHash:    passwordHash,
		MaxViewCount:    input.MaxViewCount,
		AllowDownload:   allowDownload,
		IsActive:        true,
	}

	ownerName := authContext.Email
	if creator, creatorErr := s.authRepository.GetUserByID(ctx, authContext.UserID); creatorErr == nil {
		ownerName = creator.FullName
	}
	if err := s.applyCreateShareLinkFields(ownerName, link, input); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, link); err != nil {
		return nil, err
	}

	response := toShareLinkResponse(link, s.publicAppURL)
	response.HasPassword = hasPassword
	s.logShareAudit(ctx, authContext.TenantID, document.ID, link.ID, audit.ActionShareCreated, "Paylaşım linki oluşturuldu", map[string]interface{}{
		"shareType":  link.ShareType,
		"permission": link.Permission,
	})
	return &response, nil
}

func (s *Service) ListShareLinks(ctx context.Context, documentID string) ([]ShareLinkResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedDocumentID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}

	if allowed, err := s.canManageShareLink(ctx, authContext, document); err != nil {
		return nil, err
	} else if !allowed {
		return nil, ErrForbidden
	}

	links, err := s.repository.ListActiveByDocument(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		return nil, err
	}

	responses := make([]ShareLinkResponse, 0, len(links))
	for _, link := range links {
		item := toShareLinkResponse(&link, s.publicAppURL)
		item.HasPassword = link.PasswordHash != nil && *link.PasswordHash != ""
		responses = append(responses, item)
	}

	return responses, nil
}

func (s *Service) DeleteShareLink(ctx context.Context, documentID, shareLinkID string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	parsedDocumentID, err := uuid.Parse(documentID)
	if err != nil {
		return ErrInvalidDocumentID
	}

	parsedShareLinkID, err := uuid.Parse(shareLinkID)
	if err != nil {
		return ErrInvalidShareLinkID
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, parsedDocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}

	if allowed, err := s.canManageShareLink(ctx, authContext, document); err != nil {
		return err
	} else if !allowed {
		return ErrForbidden
	}

	link, err := s.repository.GetByID(ctx, authContext.TenantID, parsedShareLinkID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShareLinkNotFound
		}
		return err
	}

	if link.DocumentID != parsedDocumentID {
		return ErrShareLinkNotFound
	}

	if err := s.repository.Revoke(ctx, authContext.TenantID, parsedShareLinkID); err != nil {
		return err
	}

	if err := s.repository.SoftDelete(ctx, authContext.TenantID, parsedShareLinkID); err != nil {
		return err
	}

	s.logShareAudit(ctx, authContext.TenantID, link.DocumentID, link.ID, audit.ActionShareRevoked, "Paylaşım linki iptal edildi", nil)
	return nil
}

func (s *Service) LogShareChannel(ctx context.Context, shareLinkID string, channel ShareChannel) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	parsedShareLinkID, err := uuid.Parse(shareLinkID)
	if err != nil {
		return ErrInvalidShareLinkID
	}

	if !isValidShareChannel(channel) {
		return ErrInvalidShareChannel
	}

	link, err := s.repository.GetByID(ctx, authContext.TenantID, parsedShareLinkID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrShareLinkNotFound
		}
		return err
	}

	if !link.IsActive {
		return ErrShareLinkInactive
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, link.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}

	if allowed, err := s.canManageShareLink(ctx, authContext, document); err != nil {
		return err
	} else if !allowed {
		return ErrForbidden
	}

	s.logShareAudit(ctx, authContext.TenantID, link.DocumentID, link.ID, audit.ActionShareLinkSharedViaChannel, "Paylaşım kanalı kullanıldı", map[string]interface{}{
		"channel": string(channel),
	})
	return nil
}

func (s *Service) GetPublicShare(ctx context.Context, token, accessToken string, meta PublicAccessMeta) (*PublicShareResponse, error) {
	link, document, state, err := s.resolveShareLink(ctx, token)
	if err != nil {
		return nil, err
	}

	status := EffectiveShareStatus(link)
	response := &PublicShareResponse{
		DocumentName:     document.OriginalFileName,
		FileExtension:    document.FileExtension,
		FileSize:         document.FileSize,
		AllowDownload:    link.AllowDownload,
		AllowPrint:       link.AllowPrint,
		AllowCopyText:    link.AllowCopyText,
		WatermarkEnabled: link.WatermarkEnabled,
		Status:           status,
		DomainRestricted: len(DecodeAllowedDomains(link.AllowedDomains)) > 0,
		Expired:          state.expired,
		MaxViewsReached:  state.maxViewsReached,
	}

	if state.expired {
		response.CanAccess = false
		response.PasswordRequired = link.PasswordHash != nil && *link.PasswordHash != ""
		_ = s.repository.MarkExpired(ctx, link.ID)
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareExpired, "Paylaşım linki süresi doldu", nil)
		return response, nil
	}

	if state.maxViewsReached {
		response.CanAccess = false
		response.PasswordRequired = link.PasswordHash != nil && *link.PasswordHash != ""
		_ = s.repository.MarkAccessLimitReached(ctx, link.ID)
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareLimitReached, "Paylaşım linki erişim limitine ulaştı", nil)
		return response, nil
	}

	if !link.IsActive || status == ShareStatusRevoked || status == ShareStatusDisabled {
		return nil, ErrShareLinkInactive
	}

	if err := s.checkDomainAccess(link, meta.VisitorEmail); err != nil {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareAccessDenied, "Paylaşım erişimi reddedildi", map[string]interface{}{
			"reason": "domain_restriction",
		})
		return nil, err
	}

	passwordRequired := link.PasswordHash != nil && *link.PasswordHash != ""
	response.PasswordRequired = passwordRequired

	accessGranted := !passwordRequired
	if passwordRequired {
		accessGranted = parseShareAccessToken(s.jwtSecret, accessToken, link.Token) == nil
	}

	response.CanAccess = accessGranted
	if accessGranted {
		response.PreviewURL = fmt.Sprintf("%s/api/v1/public/share/%s/download", s.publicAPIURL, link.Token)
		if err := s.repository.RecordShareOpen(ctx, link.ID, meta.IPAddress, meta.UserAgent); err != nil {
			return nil, err
		}
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareOpened, "Paylaşım linki açıldı", map[string]interface{}{
			"ipAddress": meta.IPAddress,
			"userAgent": meta.UserAgent,
		})
	}

	return response, nil
}

func (s *Service) VerifyPassword(ctx context.Context, token, password string, meta PublicAccessMeta) (*VerifyPasswordResponse, error) {
	link, document, state, err := s.resolveShareLink(ctx, token)
	if err != nil {
		return nil, err
	}

	if state.expired {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareExpired, "Paylaşım linki süresi doldu", nil)
		return nil, ErrShareLinkExpired
	}
	if state.maxViewsReached {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareLimitReached, "Paylaşım linki erişim limitine ulaştı", nil)
		return nil, ErrMaxViewsReached
	}
	if !link.IsActive {
		return nil, ErrShareLinkInactive
	}
	if err := s.checkDomainAccess(link, meta.VisitorEmail); err != nil {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareAccessDenied, "Paylaşım erişimi reddedildi", map[string]interface{}{
			"reason": "domain_restriction",
		})
		return nil, err
	}
	if link.PasswordHash == nil || *link.PasswordHash == "" {
		return nil, ErrPasswordRequired
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*link.PasswordHash), []byte(password)); err != nil {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionSharePasswordFailed, "Paylaşım şifresi başarısız", nil)
		return nil, ErrInvalidPassword
	}

	accessToken, expiresIn, err := issueShareAccessToken(s.jwtSecret, link.ID.String(), link.Token)
	if err != nil {
		return nil, err
	}

	s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareLinkPasswordVerified, "Paylaşım şifresi doğrulandı", nil)
	return &VerifyPasswordResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

type shareDocumentFile struct {
	Reader           io.ReadCloser
	OriginalFileName string
	MimeType         string
	Size             int64
	AllowDownload    bool
}

func (s *Service) GetPublicDownload(ctx context.Context, token, accessToken, accessMode string, meta PublicAccessMeta) (*shareDocumentFile, error) {
	link, document, state, err := s.resolveShareLink(ctx, token)
	if err != nil {
		return nil, err
	}

	if state.expired {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareExpired, "Paylaşım linki süresi doldu", nil)
		return nil, ErrShareLinkExpired
	}
	if state.maxViewsReached {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareLimitReached, "Paylaşım linki erişim limitine ulaştı", nil)
		return nil, ErrMaxViewsReached
	}
	if !link.IsActive {
		return nil, ErrShareLinkInactive
	}
	if err := s.checkDomainAccess(link, meta.VisitorEmail); err != nil {
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareAccessDenied, "Paylaşım erişimi reddedildi", map[string]interface{}{
			"reason": "domain_restriction",
		})
		return nil, err
	}

	if link.PasswordHash != nil && *link.PasswordHash != "" {
		if err := parseShareAccessToken(s.jwtSecret, accessToken, link.Token); err != nil {
			return nil, ErrPasswordRequired
		}
	}

	mode := strings.TrimSpace(strings.ToLower(accessMode))
	if mode == "download" && !link.AllowDownload {
		return nil, ErrDownloadNotAllowed
	}
	if mode == "print" && !link.AllowPrint {
		return nil, ErrShareAccessDenied
	}

	location, err := s.storage.ResolveLocation(
		document.TenantID,
		document.StorageProvider,
		document.StorageKey,
		document.FilePath,
		document.BucketName,
		document.StoredFileName,
	)
	if err != nil {
		return nil, err
	}
	if document.TenantID != uuid.Nil && !storage.ValidateTenantKey(document.TenantID, location.Key) && location.Provider == storage.ProviderMinIO {
		return nil, ErrForbidden
	}

	reader, err := s.storage.DownloadWithRetry(ctx, location, s.downloadMaxRetries, s.downloadTimeout)
	if err != nil {
		return nil, err
	}

	s.logPublicShareFileAccess(ctx, link, document, accessMode, meta)

	return &shareDocumentFile{
		Reader:           reader,
		OriginalFileName: document.OriginalFileName,
		MimeType:         document.MimeType,
		Size:             document.FileSize,
		AllowDownload:    link.AllowDownload,
	}, nil
}

type shareLinkState struct {
	expired         bool
	maxViewsReached bool
}

func (s *Service) resolveShareLink(ctx context.Context, token string) (*DocumentShareLink, *documents.Document, shareLinkState, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil, shareLinkState{}, ErrInvalidShareToken
	}

	link, err := s.repository.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, shareLinkState{}, ErrShareLinkNotFound
		}
		return nil, nil, shareLinkState{}, err
	}

	state := shareLinkState{}
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		state.expired = true
	}
	if link.MaxViewCount != nil && link.ViewCount >= *link.MaxViewCount {
		state.maxViewsReached = true
	}

	document, err := s.documentRepository.GetByID(ctx, link.TenantID, link.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, shareLinkState{}, ErrDocumentNotFound
		}
		return nil, nil, shareLinkState{}, err
	}

	return link, document, state, nil
}

func (s *Service) canManageShareLink(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error) {
	sharedPermission, err := s.documentRepository.GetInternalSharePermission(ctx, authContext.TenantID, document.ID, authContext.UserID)
	if err != nil {
		return false, err
	}
	return permissions.CanShareDocument(
		permissions.SubjectFromAuth(authContext),
		permissions.DocumentFromFields(document.TenantID, document.UploadedByUserID, sharedPermission),
	), nil
}

func isValidShareChannel(channel ShareChannel) bool {
	switch channel {
	case ShareChannelWhatsApp, ShareChannelTelegram, ShareChannelEmail, ShareChannelCopy, ShareChannelNative:
		return true
	default:
		return false
	}
}

func generateShareToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func toShareLinkResponse(link *DocumentShareLink, publicAppURL string) ShareLinkResponse {
	return ShareLinkResponse{
		ID:               link.ID.String(),
		DocumentID:       link.DocumentID.String(),
		ShareType:        link.ShareType,
		Status:           EffectiveShareStatus(link),
		ShareURL:         fmt.Sprintf("%s/share/%s", publicAppURL, link.Token),
		Token:            link.Token,
		Permission:       link.Permission,
		ExpirationPreset: link.ExpirationPreset,
		ExpiresAt:        link.ExpiresAt,
		MaxViewCount:     link.MaxViewCount,
		ViewCount:        link.ViewCount,
		DownloadCount:    link.DownloadCount,
		AllowDownload:    link.AllowDownload,
		AllowPrint:       link.AllowPrint,
		AllowCopyText:    link.AllowCopyText,
		AllowedDomains:   DecodeAllowedDomains(link.AllowedDomains),
		WatermarkEnabled: link.WatermarkEnabled,
		IsActive:         link.IsActive,
		InvitedEmail:     link.InvitedEmail,
		CreatedAt:        link.CreatedAt,
	}
}

func toShareLinkAnalyticsResponse(analytics *ShareLinkAnalytics) ShareLinkAnalyticsResponse {
	return ShareLinkAnalyticsResponse{
		ShareLinkID:      analytics.ShareLinkID.String(),
		DocumentID:       analytics.DocumentID.String(),
		OpenCount:        analytics.OpenCount,
		DownloadCount:    analytics.DownloadCount,
		LastOpenedAt:     analytics.LastOpenedAt,
		LastDownloadedAt: analytics.LastDownloadedAt,
		LastIPAddress:    analytics.LastIPAddress,
		LastUserAgent:    analytics.LastUserAgent,
		CreatedByUserID:  analytics.CreatedByUserID.String(),
		CreatedByName:    analytics.CreatedByName,
		Status:           analytics.Status,
	}
}

func (s *Service) logPublicShareFileAccess(ctx context.Context, link *DocumentShareLink, document *documents.Document, accessMode string, meta PublicAccessMeta) {
	if s.auditRecorder == nil || link == nil || document == nil {
		return
	}

	mode := strings.TrimSpace(strings.ToLower(accessMode))
	if mode == "download" {
		_ = s.repository.RecordShareDownload(ctx, link.ID, meta.IPAddress, meta.UserAgent)
		s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareDownloaded, "Paylaşım linkinden indirildi", map[string]interface{}{
			"accessMode": mode,
			"ipAddress":  meta.IPAddress,
			"userAgent":  meta.UserAgent,
		})
		return
	}

	s.logShareAudit(ctx, link.TenantID, document.ID, link.ID, audit.ActionShareOpened, "Paylaşım linki önizlendi", map[string]interface{}{
		"accessMode": mode,
		"ipAddress":  meta.IPAddress,
		"userAgent":  meta.UserAgent,
	})
}

package sharing

import (
	"context"
	"errors"
	"strings"

	"fifriz/internal/audit"
	"fifriz/internal/auth"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PublicAccessMeta struct {
	IPAddress    string
	UserAgent    string
	VisitorEmail string
}

func (s *Service) BulkCreateShareLinks(ctx context.Context, input BulkCreateShareLinksRequest) ([]ShareLinkResponse, error) {
	if len(input.DocumentIDs) == 0 {
		return nil, ErrInvalidDocumentID
	}

	responses := make([]ShareLinkResponse, 0, len(input.DocumentIDs))
	for _, documentID := range input.DocumentIDs {
		result, err := s.CreateShareLink(ctx, documentID, input.CreateShareLinkRequest)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *result)
	}
	return responses, nil
}

func (s *Service) GetShareLinkAnalytics(ctx context.Context, shareLinkID string) (*ShareLinkAnalyticsResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	parsedShareLinkID, err := uuid.Parse(shareLinkID)
	if err != nil {
		return nil, ErrInvalidShareLinkID
	}

	link, err := s.repository.GetByID(ctx, authContext.TenantID, parsedShareLinkID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShareLinkNotFound
		}
		return nil, err
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, link.DocumentID)
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

	analytics, err := s.repository.AnalyticsForLink(ctx, authContext.TenantID, parsedShareLinkID)
	if err != nil {
		return nil, err
	}

	creator, err := s.authRepository.GetUserByID(ctx, link.CreatedByUserID)
	if err == nil {
		analytics.CreatedByName = creator.FullName
	}

	response := toShareLinkAnalyticsResponse(analytics)
	return &response, nil
}

func (s *Service) RevokeShareLink(ctx context.Context, documentID, shareLinkID string) error {
	return s.DeleteShareLink(ctx, documentID, shareLinkID)
}

func (s *Service) applyCreateShareLinkFields(
	ownerName string,
	link *DocumentShareLink,
	input CreateShareLinkRequest,
) error {
	shareType := strings.TrimSpace(input.ShareType)
	if shareType == "" {
		shareType = ShareTypePublicLink
	}
	switch shareType {
	case ShareTypePublicLink, ShareTypeEmailInvite:
		link.ShareType = shareType
	default:
		return ErrInvalidShareType
	}

	preset := strings.TrimSpace(input.ExpirationPreset)
	if preset == "" && input.ExpiresAt != nil {
		preset = ExpirationCustom
	}
	link.ExpirationPreset = preset
	link.ExpiresAt = ResolveExpiration(preset, input.ExpiresAt)

	if input.AllowPrint != nil {
		link.AllowPrint = *input.AllowPrint
	}
	if input.AllowCopyText != nil {
		link.AllowCopyText = *input.AllowCopyText
	}
	if len(input.AllowedDomains) > 0 {
		link.AllowedDomains = EncodeAllowedDomains(input.AllowedDomains)
	}
	if input.WatermarkEnabled != nil {
		link.WatermarkEnabled = *input.WatermarkEnabled
	}
	if input.WatermarkEnabled != nil && *input.WatermarkEnabled {
		template := DefaultWatermarkTemplate(ownerName)
		if input.WatermarkTemplate != nil && strings.TrimSpace(*input.WatermarkTemplate) != "" {
			template = strings.TrimSpace(*input.WatermarkTemplate)
		}
		link.WatermarkTemplate = &template
	}
	if input.InvitedEmail != nil && strings.TrimSpace(*input.InvitedEmail) != "" {
		value := strings.TrimSpace(*input.InvitedEmail)
		link.InvitedEmail = &value
	}
	link.Status = ShareStatusActive
	return nil
}

func (s *Service) checkDomainAccess(link *DocumentShareLink, visitorEmail string) error {
	domains := DecodeAllowedDomains(link.AllowedDomains)
	if len(domains) == 0 {
		return nil
	}
	if !IsEmailDomainAllowed(visitorEmail, domains) {
		return ErrShareAccessDenied
	}
	return nil
}

func (s *Service) logShareAudit(
	ctx context.Context,
	tenantID uuid.UUID,
	documentID uuid.UUID,
	shareLinkID uuid.UUID,
	action string,
	description string,
	meta map[string]interface{},
) {
	if s.auditRecorder == nil {
		return
	}
	if meta == nil {
		meta = map[string]interface{}{}
	}
	meta["shareLinkId"] = shareLinkID.String()
	audit.LogDocumentAction(ctx, s.auditRecorder, tenantID, nil, action, documentID, description, meta)
}

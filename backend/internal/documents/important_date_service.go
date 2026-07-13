package documents

import (
	"context"
	"errors"
	"strings"
	"time"

	"fifriz/internal/applog"
	"fifriz/internal/auth"
	domain "fifriz/internal/events/events"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrDocumentImportantDateNotFound = errors.New("document important date not found")
	ErrInvalidImportantDateID        = errors.New("invalid important date id")
	ErrEmptyImportantDateTitle       = errors.New("empty important date title")
	ErrInvalidImportantDateValue     = errors.New("invalid important date value")
	ErrInvalidImportantDateType      = errors.New("invalid important date type")
	ErrInvalidAssignedUser           = errors.New("assigned user must belong to the tenant")
)

func (s *Service) ListImportantDates(ctx context.Context, documentIDRaw string) ([]DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListImportantDatesByDocument(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}
	return toImportantDateResponses(rows), nil
}

func (s *Service) ListImportantDatesGlobal(ctx context.Context, query ImportantDateListQuery) ([]DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	tenantWide := HasTenantWideDocumentAccess(authContext.Role)
	rows, err := s.repository.ListImportantDates(ctx, authContext.TenantID, authContext.UserID, tenantWide, query)
	if err != nil {
		return nil, err
	}
	return toImportantDateResponses(rows), nil
}

func (s *Service) GetDueReminders(ctx context.Context, reminderDate time.Time) ([]ImportantDateListRow, error) {
	return s.repository.GetDueReminders(ctx, reminderDate)
}

func (s *Service) CreateImportantDate(ctx context.Context, documentIDRaw string, input CreateDocumentImportantDateRequest) (*DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrEmptyImportantDateTitle
	}

	parsedDate, err := parseImportantDate(input.Date)
	if err != nil {
		return nil, ErrInvalidImportantDateValue
	}
	if err := validateImportantDateType(input.Type); err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireEditAccess(ctx, authContext, document); err != nil {
		return nil, err
	}
	if err := s.validateAssignedUser(ctx, authContext.TenantID, input.AssignedToUserID); err != nil {
		return nil, err
	}

	item := &DocumentImportantDate{
		TenantID:           authContext.TenantID,
		DocumentID:         documentID,
		CreatedByUserID:    authContext.UserID,
		AssignedToUserID:   input.AssignedToUserID,
		Title:              title,
		Date:               parsedDate,
		ReminderDaysBefore: input.ReminderDaysBefore,
		Note:               normalizeOptionalText(input.Note),
		Type:               normalizeImportantDateType(input.Type),
		Source:             ImportantDateSourceManual,
		Status:             ImportantDateStatusActive,
	}
	if err := s.repository.CreateImportantDate(ctx, item); err != nil {
		return nil, err
	}

	row, err := s.repository.GetImportantDateWithDetails(ctx, authContext.TenantID, documentID, item.ID)
	if err != nil {
		response := toDocumentImportantDateResponse(ImportantDateListRow{DocumentImportantDate: *item, DocumentName: document.OriginalFileName})
		s.publishDocumentUpdated(ctx, authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName)
		return &response, nil
	}

	s.publishDocumentUpdated(ctx, authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName)
	response := toDocumentImportantDateResponse(*row)
	return &response, nil
}

func (s *Service) UpdateImportantDate(ctx context.Context, documentIDRaw, dateIDRaw string, input UpdateDocumentImportantDateRequest) (*DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrEmptyImportantDateTitle
	}

	parsedDate, err := parseImportantDate(input.Date)
	if err != nil {
		return nil, ErrInvalidImportantDateValue
	}
	if err := validateImportantDateType(input.Type); err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}
	dateID, err := uuid.Parse(dateIDRaw)
	if err != nil {
		return nil, ErrInvalidImportantDateID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	item, err := s.repository.GetImportantDateByID(ctx, authContext.TenantID, documentID, dateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentImportantDateNotFound
		}
		return nil, err
	}

	canEditDocument, err := s.shareAccess.CanEdit(ctx, authContext, document)
	if err != nil {
		return nil, err
	}
	subject := permissions.SubjectFromAuth(authContext)
	if !permissions.CanUpdateDocumentImportantDate(subject, item.CreatedByUserID, canEditDocument) {
		applog.PermissionDenied("update_important_date", item.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return nil, ErrForbidden
	}
	if err := s.validateAssignedUser(ctx, authContext.TenantID, input.AssignedToUserID); err != nil {
		return nil, err
	}

	item.Title = title
	item.Date = parsedDate
	item.ReminderDaysBefore = input.ReminderDaysBefore
	item.Note = normalizeOptionalText(input.Note)
	item.Type = normalizeImportantDateType(input.Type)
	item.AssignedToUserID = input.AssignedToUserID
	if err := s.repository.UpdateImportantDate(ctx, item); err != nil {
		return nil, err
	}

	row, err := s.repository.GetImportantDateWithDetails(ctx, authContext.TenantID, documentID, item.ID)
	if err != nil {
		response := toDocumentImportantDateResponse(ImportantDateListRow{DocumentImportantDate: *item, DocumentName: document.OriginalFileName})
		s.publishDocumentUpdated(ctx, authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName)
		return &response, nil
	}

	s.publishDocumentUpdated(ctx, authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName)
	response := toDocumentImportantDateResponse(*row)
	return &response, nil
}

func (s *Service) DeleteImportantDate(ctx context.Context, documentIDRaw, dateIDRaw string) error {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return ErrInvalidDocumentID
	}
	dateID, err := uuid.Parse(dateIDRaw)
	if err != nil {
		return ErrInvalidImportantDateID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return err
	}

	item, err := s.repository.GetImportantDateByID(ctx, authContext.TenantID, documentID, dateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDocumentImportantDateNotFound
		}
		return err
	}
	if err := s.requireDeleteImportantDateAccess(authContext, item); err != nil {
		return err
	}

	if err := s.repository.SoftDeleteImportantDate(ctx, authContext.TenantID, documentID, dateID); err != nil {
		return err
	}

	s.publishDocumentUpdated(ctx, authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName)
	return nil
}

func (s *Service) CompleteImportantDate(ctx context.Context, documentIDRaw, dateIDRaw string) (*DocumentImportantDateResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}
	dateID, err := uuid.Parse(dateIDRaw)
	if err != nil {
		return nil, ErrInvalidImportantDateID
	}

	document, err := s.repository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	item, err := s.repository.GetImportantDateByID(ctx, authContext.TenantID, documentID, dateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentImportantDateNotFound
		}
		return nil, err
	}
	if err := s.requireCompleteImportantDateAccess(authContext, item); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item.IsCompleted = true
	item.CompletedAt = &now
	if err := s.repository.UpdateImportantDate(ctx, item); err != nil {
		return nil, err
	}

	row, err := s.repository.GetImportantDateWithDetails(ctx, authContext.TenantID, documentID, item.ID)
	if err != nil {
		response := toDocumentImportantDateResponse(ImportantDateListRow{DocumentImportantDate: *item, DocumentName: document.OriginalFileName})
		s.publishDocumentUpdated(ctx, authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName)
		return &response, nil
	}

	s.publishDocumentUpdated(ctx, authContext.TenantID, authContext.UserID, document.ID, document.OriginalFileName)
	response := toDocumentImportantDateResponse(*row)
	return &response, nil
}

func (s *Service) publishDocumentUpdated(ctx context.Context, tenantID, userID, documentID uuid.UUID, fileName string) {
	s.publisher.PublishAfterCommit(ctx, domain.NewDocumentUpdated(tenantID, userID, documentID, fileName))
}

func (s *Service) requireDeleteImportantDateAccess(authContext auth.Context, item *DocumentImportantDate) error {
	subject := permissions.SubjectFromAuth(authContext)
	if !permissions.CanDeleteDocumentImportantDate(subject, item.CreatedByUserID) {
		applog.PermissionDenied("delete_important_date", item.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireCompleteImportantDateAccess(authContext auth.Context, item *DocumentImportantDate) error {
	subject := permissions.SubjectFromAuth(authContext)
	if !permissions.CanCompleteDocumentImportantDate(subject, item.CreatedByUserID, item.AssignedToUserID) {
		applog.PermissionDenied("complete_important_date", item.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}

func (s *Service) validateAssignedUser(ctx context.Context, tenantID uuid.UUID, assignedToUserID *uuid.UUID) error {
	if assignedToUserID == nil || *assignedToUserID == uuid.Nil {
		return nil
	}
	exists, err := s.repository.TenantMemberExists(ctx, tenantID, *assignedToUserID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidAssignedUser
	}
	return nil
}

func parseImportantDate(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, ErrInvalidImportantDateValue
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, ErrInvalidImportantDateValue
	}
	return parsed.UTC(), nil
}

func validateImportantDateType(value *string) error {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	switch strings.TrimSpace(*value) {
	case ImportantDateTypeExpiration, ImportantDateTypeRenewal, ImportantDateTypePaymentDue, ImportantDateTypeReview, ImportantDateTypeCustom:
		return nil
	default:
		return ErrInvalidImportantDateType
	}
}

func normalizeImportantDateType(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func parseImportantDateListQuery(c QueryParser) ImportantDateListQuery {
	query := ImportantDateListQuery{
		UpcomingOnly: parseBoolQuery(c.Query("upcomingOnly")),
		AssignedToMe: parseBoolQuery(c.Query("assignedToMe")),
	}
	if completedRaw := strings.TrimSpace(c.Query("completed")); completedRaw != "" {
		value := parseBoolQuery(completedRaw)
		query.Completed = &value
	}
	if typeRaw := strings.TrimSpace(c.Query("type")); typeRaw != "" {
		query.Type = &typeRaw
	}
	if fromRaw := strings.TrimSpace(c.Query("from")); fromRaw != "" {
		if parsed, err := time.Parse("2006-01-02", fromRaw); err == nil {
			query.From = &parsed
		}
	}
	if toRaw := strings.TrimSpace(c.Query("to")); toRaw != "" {
		if parsed, err := time.Parse("2006-01-02", toRaw); err == nil {
			query.To = &parsed
		}
	}
	return query
}

type QueryParser interface {
	Query(key string, defaultValue ...string) string
}

func parseBoolQuery(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

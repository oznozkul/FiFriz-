package approvals

import (
	"context"
	"errors"
	"sort"
	"strings"

	"fifriz/internal/applog"
	"fifriz/internal/auth"
	"fifriz/internal/documents"
	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/permissions"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrForbidden              = errors.New("forbidden")
	ErrWorkflowNotFound       = errors.New("workflow not found")
	ErrInvalidWorkflowID      = errors.New("invalid workflow id")
	ErrInvalidDocumentID      = errors.New("invalid document id")
	ErrDocumentNotFound       = errors.New("document not found")
	ErrInvalidWorkflowInput   = errors.New("invalid workflow input")
	ErrWorkflowNotPending     = errors.New("workflow is not pending")
	ErrNotCurrentApprover     = errors.New("not current step approver")
	ErrInvalidStepOrder       = errors.New("invalid step order")
	ErrApproverNotInTenant    = errors.New("approver not in tenant")
)

type DocumentAccess interface {
	CanView(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
	CanEdit(ctx context.Context, authContext auth.Context, document *documents.Document) (bool, error)
}

type Service struct {
	repository         *Repository
	documentRepository *documents.Repository
	authRepository     *auth.Repository
	access             DocumentAccess
	publisher          events.Publisher
}

func NewService(
	repository *Repository,
	documentRepository *documents.Repository,
	authRepository *auth.Repository,
	access DocumentAccess,
	publisher events.Publisher,
) *Service {
	return &Service{
		repository:         repository,
		documentRepository: documentRepository,
		authRepository:     authRepository,
		access:             access,
		publisher:          publisher,
	}
}

func (s *Service) CreateForDocument(ctx context.Context, documentIDRaw string, input CreateWorkflowRequest) (*WorkflowResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	title := strings.TrimSpace(input.Title)
	if title == "" || len(input.Steps) == 0 {
		return nil, ErrInvalidWorkflowInput
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireCreateAccess(ctx, authContext, document); err != nil {
		return nil, err
	}

	steps, err := s.validateAndBuildSteps(ctx, authContext.TenantID, input.Steps)
	if err != nil {
		return nil, err
	}

	var description *string
	if input.Description != nil {
		if trimmed := strings.TrimSpace(*input.Description); trimmed != "" {
			description = &trimmed
		}
	}

	workflow := &ApprovalWorkflow{
		TenantID:         authContext.TenantID,
		DocumentID:       document.ID,
		CreatedByUserID:  authContext.UserID,
		Title:            title,
		Description:      description,
		Status:           WorkflowStatusPending,
		CurrentStepOrder: 1,
	}
	for i := range steps {
		steps[i].TenantID = authContext.TenantID
		steps[i].Status = StepStatusPending
	}

	if err := s.repository.CreateWorkflowWithSteps(ctx, workflow, steps, "pending"); err != nil {
		if errors.Is(err, ErrActiveWorkflowExists) {
			return nil, ErrActiveWorkflowExists
		}
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewWorkflowCreated(
		authContext.TenantID,
		authContext.UserID,
		workflow.ID,
		document.ID,
		document.UploadedByUserID,
		document.OriginalFileName,
		workflow.Title,
	))

	return s.buildWorkflowResponse(ctx, authContext.TenantID, workflow.ID)
}

func (s *Service) ListByDocument(ctx context.Context, documentIDRaw string) ([]WorkflowResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	documentID, err := uuid.Parse(documentIDRaw)
	if err != nil {
		return nil, ErrInvalidDocumentID
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, documentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document, uuid.Nil); err != nil {
		return nil, err
	}

	rows, err := s.repository.ListByDocument(ctx, authContext.TenantID, documentID)
	if err != nil {
		return nil, err
	}

	responses := make([]WorkflowResponse, 0, len(rows))
	for _, row := range rows {
		item, err := s.buildWorkflowResponse(ctx, authContext.TenantID, row.ID)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *item)
	}
	return responses, nil
}

func (s *Service) GetByID(ctx context.Context, workflowIDRaw string) (*WorkflowResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return nil, ErrInvalidWorkflowID
	}

	workflow, err := s.repository.GetWorkflowByID(ctx, authContext.TenantID, workflowID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, workflow.DocumentID)
	if err != nil {
		return nil, err
	}
	if err := s.requireViewAccess(ctx, authContext, document, workflow.ID); err != nil {
		return nil, err
	}

	return s.buildWorkflowResponse(ctx, authContext.TenantID, workflow.ID)
}

func (s *Service) ListPendingForMe(ctx context.Context) ([]PendingApprovalResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repository.ListPendingForApprover(ctx, authContext.TenantID, authContext.UserID)
}

func (s *Service) Approve(ctx context.Context, workflowIDRaw string, input DecideWorkflowRequest) (*WorkflowResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return nil, ErrInvalidWorkflowID
	}

	workflow, document, step, err := s.loadPendingCurrentStep(ctx, authContext, workflowID)
	if err != nil {
		return nil, err
	}

	var note *string
	if input.Note != nil {
		if trimmed := strings.TrimSpace(*input.Note); trimmed != "" {
			note = &trimmed
		}
	}

	maxStep, err := s.repository.GetMaxStepOrder(ctx, authContext.TenantID, workflow.ID)
	if err != nil {
		return nil, err
	}

	completed := workflow.CurrentStepOrder >= maxStep
	docApprovalStatus := "pending"
	if completed {
		docApprovalStatus = "approved"
	}

	if err := s.repository.ApproveStepAdvance(ctx, workflow, step, note, docApprovalStatus, completed); err != nil {
		return nil, err
	}

	if completed {
		s.publisher.PublishAfterCommit(ctx, domain.NewWorkflowApproved(
			authContext.TenantID,
			authContext.UserID,
			workflow.ID,
			document.ID,
			workflow.CreatedByUserID,
			document.OriginalFileName,
			workflow.Title,
		))
		if document.UploadedByUserID != workflow.CreatedByUserID {
			s.publisher.PublishAfterCommit(ctx, domain.NewWorkflowApproved(
				authContext.TenantID,
				authContext.UserID,
				workflow.ID,
				document.ID,
				document.UploadedByUserID,
				document.OriginalFileName,
				workflow.Title,
			))
		}
	}

	return s.buildWorkflowResponse(ctx, authContext.TenantID, workflow.ID)
}

func (s *Service) Reject(ctx context.Context, workflowIDRaw string, input DecideWorkflowRequest) (*WorkflowResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return nil, ErrInvalidWorkflowID
	}

	workflow, document, step, err := s.loadPendingCurrentStep(ctx, authContext, workflowID)
	if err != nil {
		return nil, err
	}

	var note *string
	if input.Note != nil {
		if trimmed := strings.TrimSpace(*input.Note); trimmed != "" {
			note = &trimmed
		}
	}

	if err := s.repository.RejectWorkflow(ctx, workflow, step, note); err != nil {
		return nil, err
	}

	s.publisher.PublishAfterCommit(ctx, domain.NewWorkflowRejected(
		authContext.TenantID,
		authContext.UserID,
		workflow.ID,
		document.ID,
		workflow.CreatedByUserID,
		document.OriginalFileName,
		workflow.Title,
	))
	if document.UploadedByUserID != workflow.CreatedByUserID {
		s.publisher.PublishAfterCommit(ctx, domain.NewWorkflowRejected(
			authContext.TenantID,
			authContext.UserID,
			workflow.ID,
			document.ID,
			document.UploadedByUserID,
			document.OriginalFileName,
			workflow.Title,
		))
	}

	return s.buildWorkflowResponse(ctx, authContext.TenantID, workflow.ID)
}

func (s *Service) Cancel(ctx context.Context, workflowIDRaw string) (*WorkflowResponse, error) {
	authContext, err := auth.RequireContext(ctx)
	if err != nil {
		return nil, err
	}

	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return nil, ErrInvalidWorkflowID
	}

	workflow, err := s.repository.GetWorkflowByID(ctx, authContext.TenantID, workflowID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWorkflowNotFound
		}
		return nil, err
	}
	if workflow.Status != WorkflowStatusPending {
		return nil, ErrWorkflowNotPending
	}
	if err := s.requireCancelAccess(ctx, authContext, workflow); err != nil {
		return nil, err
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, workflow.DocumentID)
	if err != nil {
		return nil, err
	}
	_ = document

	if err := s.repository.CancelWorkflow(ctx, workflow); err != nil {
		return nil, err
	}

	return s.buildWorkflowResponse(ctx, authContext.TenantID, workflow.ID)
}

func (s *Service) loadPendingCurrentStep(ctx context.Context, authContext auth.Context, workflowID uuid.UUID) (*ApprovalWorkflow, *documents.Document, *ApprovalStep, error) {
	workflow, err := s.repository.GetWorkflowByID(ctx, authContext.TenantID, workflowID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, ErrWorkflowNotFound
		}
		return nil, nil, nil, err
	}
	if workflow.Status != WorkflowStatusPending {
		return nil, nil, nil, ErrWorkflowNotPending
	}

	step, err := s.repository.GetCurrentStep(ctx, authContext.TenantID, workflow.ID, workflow.CurrentStepOrder)
	if err != nil {
		return nil, nil, nil, err
	}
	if step.ApproverUserID != authContext.UserID {
		applog.PermissionDenied("approve_workflow_step", workflow.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return nil, nil, nil, ErrNotCurrentApprover
	}

	document, err := s.documentRepository.GetByID(ctx, authContext.TenantID, workflow.DocumentID)
	if err != nil {
		return nil, nil, nil, err
	}

	return workflow, document, step, nil
}

func (s *Service) validateAndBuildSteps(ctx context.Context, tenantID uuid.UUID, inputs []CreateWorkflowStepInput) ([]ApprovalStep, error) {
	if len(inputs) == 0 {
		return nil, ErrInvalidWorkflowInput
	}

	sorted := make([]CreateWorkflowStepInput, len(inputs))
	copy(sorted, inputs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].StepOrder < sorted[j].StepOrder })

	steps := make([]ApprovalStep, 0, len(sorted))
	for i, item := range sorted {
		expected := i + 1
		if item.StepOrder != expected {
			return nil, ErrInvalidStepOrder
		}
		if item.ApproverUserID == uuid.Nil {
			return nil, ErrInvalidWorkflowInput
		}
		if _, err := s.authRepository.GetMembership(ctx, item.ApproverUserID, tenantID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrApproverNotInTenant
			}
			return nil, err
		}
		steps = append(steps, ApprovalStep{
			StepOrder:      item.StepOrder,
			ApproverUserID: item.ApproverUserID,
			Status:         StepStatusPending,
		})
	}
	return steps, nil
}

func (s *Service) requireCreateAccess(ctx context.Context, authContext auth.Context, document *documents.Document) error {
	if s.access == nil {
		return nil
	}
	allowed, err := s.access.CanEdit(ctx, authContext, document)
	if err != nil {
		return err
	}
	if !allowed {
		applog.PermissionDenied("create_workflow", document.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
		return ErrForbidden
	}
	return nil
}

func (s *Service) requireViewAccess(ctx context.Context, authContext auth.Context, document *documents.Document, workflowID uuid.UUID) error {
	subject := permissions.SubjectFromAuth(authContext)
	docResource := permissions.DocumentFromFields(document.TenantID, document.UploadedByUserID, "")
	if permissions.IsPrivilegedDocumentUser(subject, docResource) {
		return nil
	}
	if document.UploadedByUserID == authContext.UserID {
		return nil
	}
	if workflowID != uuid.Nil {
		workflow, err := s.repository.GetWorkflowByID(ctx, authContext.TenantID, workflowID)
		if err == nil && workflow.CreatedByUserID == authContext.UserID {
			return nil
		}
		isApprover, err := s.repository.IsWorkflowApprover(ctx, authContext.TenantID, workflowID, authContext.UserID)
		if err == nil && isApprover {
			return nil
		}
	}
	if s.access != nil {
		allowed, err := s.access.CanView(ctx, authContext, document)
		if err != nil {
			return err
		}
		if allowed {
			return nil
		}
	}
	applog.PermissionDenied("view_workflow", document.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
	return ErrForbidden
}

func (s *Service) requireCancelAccess(ctx context.Context, authContext auth.Context, workflow *ApprovalWorkflow) error {
	subject := permissions.SubjectFromAuth(authContext)
	if permissions.HasTenantWideAccess(subject.Role) {
		return nil
	}
	if workflow.CreatedByUserID == authContext.UserID {
		return nil
	}
	applog.PermissionDenied("cancel_workflow", workflow.ID.String(), authContext.UserID.String(), authContext.TenantID.String())
	return ErrForbidden
}

func (s *Service) buildWorkflowResponse(ctx context.Context, tenantID, workflowID uuid.UUID) (*WorkflowResponse, error) {
	workflow, err := s.repository.GetWorkflowByID(ctx, tenantID, workflowID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repository.ListByDocument(ctx, tenantID, workflow.DocumentID)
	if err != nil {
		return nil, err
	}
	var meta WorkflowListRow
	for _, row := range rows {
		if row.ID == workflowID {
			meta = row
			break
		}
	}
	stepRows, err := s.repository.ListStepsByWorkflow(ctx, tenantID, workflowID)
	if err != nil {
		return nil, err
	}

	documentName := meta.DocumentName
	createdByName := meta.CreatedByFullName
	if documentName == "" {
		if document, docErr := s.documentRepository.GetByID(ctx, tenantID, workflow.DocumentID); docErr == nil {
			documentName = document.OriginalFileName
		}
	}
	if createdByName == "" {
		if user, userErr := s.authRepository.GetUserByID(ctx, workflow.CreatedByUserID); userErr == nil {
			createdByName = user.FullName
		}
	}

	steps := make([]WorkflowStepResponse, 0, len(stepRows))
	for _, row := range stepRows {
		steps = append(steps, WorkflowStepResponse{
			ID:               row.ID,
			StepOrder:        row.StepOrder,
			ApproverUserID:   row.ApproverUserID,
			ApproverFullName: row.ApproverFullName,
			Status:           row.Status,
			Note:             row.Note,
			DecidedAt:        row.DecidedAt,
			IsCurrent:        workflow.Status == WorkflowStatusPending && row.StepOrder == workflow.CurrentStepOrder,
			CreatedAt:        row.CreatedAt,
		})
	}
	return &WorkflowResponse{
		ID:                workflow.ID,
		DocumentID:        workflow.DocumentID,
		DocumentName:      documentName,
		CreatedByUserID:   workflow.CreatedByUserID,
		CreatedByFullName: createdByName,
		Title:             workflow.Title,
		Description:       workflow.Description,
		Status:            workflow.Status,
		CurrentStepOrder:  workflow.CurrentStepOrder,
		Steps:             steps,
		CreatedAt:         workflow.CreatedAt,
		UpdatedAt:         workflow.UpdatedAt,
		CompletedAt:       workflow.CompletedAt,
	}, nil
}

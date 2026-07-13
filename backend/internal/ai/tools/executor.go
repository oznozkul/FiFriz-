package tools

import (
	"context"
	"fmt"

	"fifriz/internal/audit"
	"fifriz/internal/auth"

	"github.com/google/uuid"
)

type Executor struct {
	registry *Registry
	audit    audit.Recorder
}

func NewExecutor(registry *Registry, auditRecorder audit.Recorder) *Executor {
	return &Executor{registry: registry, audit: auditRecorder}
}

type ExecuteOptions struct {
	SuggestionID *uuid.UUID
	DocumentID   *uuid.UUID
}

func (e *Executor) Execute(
	ctx context.Context,
	user auth.Context,
	toolName string,
	input map[string]interface{},
	opts ExecuteOptions,
) (*ExecutionResult, error) {
	tool, ok := e.registry.Get(toolName)
	if !ok {
		return nil, ErrToolNotFound
	}

	if err := tool.CanExecute(ctx, user, input); err != nil {
		e.logFailure(ctx, user, toolName, input, opts, err.Error())
		return nil, err
	}

	output, err := tool.Execute(ctx, user, input)
	if err != nil {
		e.logFailure(ctx, user, toolName, input, opts, err.Error())
		return &ExecutionResult{ToolName: toolName, Success: false, Error: err.Error()}, fmt.Errorf("%w: %v", ErrToolExecutionFailed, err)
	}

	e.logSuccess(ctx, user, toolName, input, opts, output)
	return &ExecutionResult{ToolName: toolName, Success: true, Output: output}, nil
}

func (e *Executor) logSuccess(ctx context.Context, user auth.Context, toolName string, input map[string]interface{}, opts ExecuteOptions, output map[string]interface{}) {
	userID := user.UserID
	metadata := auditMetadata(toolName, input, opts)
	metadata["output"] = output
	if opts.DocumentID != nil {
		audit.LogDocumentAction(ctx, e.audit, user.TenantID, &userID, audit.ActionAIToolExecuted, *opts.DocumentID, "ai tool executed", metadata)
		return
	}
	audit.LogWithActor(ctx, e.audit, user.TenantID, &userID, audit.CreateLogInput{
		TenantID:    user.TenantID,
		Action:      audit.ActionAIToolExecuted,
		EntityType:  audit.EntityTypeDocument,
		Description: "ai tool executed",
		Metadata:    metadata,
	})
}

func (e *Executor) logFailure(ctx context.Context, user auth.Context, toolName string, input map[string]interface{}, opts ExecuteOptions, message string) {
	userID := user.UserID
	metadata := auditMetadata(toolName, input, opts)
	metadata["error"] = message
	if opts.DocumentID != nil {
		audit.LogDocumentAction(ctx, e.audit, user.TenantID, &userID, audit.ActionAIToolExecutionFailed, *opts.DocumentID, "ai tool execution failed", metadata)
		return
	}
	audit.LogWithActor(ctx, e.audit, user.TenantID, &userID, audit.CreateLogInput{
		TenantID:    user.TenantID,
		Action:      audit.ActionAIToolExecutionFailed,
		EntityType:  audit.EntityTypeDocument,
		Description: "ai tool execution failed",
		Metadata:    metadata,
	})
}

func auditMetadata(toolName string, input map[string]interface{}, opts ExecuteOptions) map[string]interface{} {
	metadata := map[string]interface{}{
		"toolName": toolName,
		"input":    input,
	}
	if opts.SuggestionID != nil {
		metadata["suggestionId"] = opts.SuggestionID.String()
	}
	if opts.DocumentID != nil {
		metadata["documentId"] = opts.DocumentID.String()
	}
	return metadata
}

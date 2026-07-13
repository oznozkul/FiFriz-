package tools

import (
	"context"
	"errors"

	"fifriz/internal/auth"
)

var (
	ErrToolNotFound      = errors.New("tool not found")
	ErrToolNotAllowed    = errors.New("tool execution not allowed")
	ErrInvalidToolInput  = errors.New("invalid tool input")
	ErrToolExecutionFailed = errors.New("tool execution failed")
)

type Tool interface {
	Name() string
	Description() string
	Schema() map[string]interface{}
	CanExecute(ctx context.Context, user auth.Context, input map[string]interface{}) error
	Execute(ctx context.Context, user auth.Context, input map[string]interface{}) (map[string]interface{}, error)
}

type ExecutionResult struct {
	ToolName string                 `json:"toolName"`
	Success  bool                   `json:"success"`
	Output   map[string]interface{} `json:"output,omitempty"`
	Error    string                 `json:"error,omitempty"`
}

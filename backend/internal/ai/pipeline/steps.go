package pipeline

import "context"

// SkipReason notes why a step was not executed.
const (
	SkipReasonUnsupported = "unsupported"
)

// noopStep is a placeholder for future AI features.
type noopStep struct {
	name string
}

func (s *noopStep) Name() string { return s.name }

func (s *noopStep) Execute(ctx context.Context, job *PipelineJob) error {
	_ = ctx
	_ = job
	return nil
}

func (s *noopStep) CanRun(ctx context.Context, job *PipelineJob) bool {
	_ = ctx
	_ = job
	return true
}

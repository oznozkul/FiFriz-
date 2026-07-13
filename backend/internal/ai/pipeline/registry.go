package pipeline

import (
	"context"
	"fmt"
	"sort"
)

// Registry holds registered pipeline steps in execution order.
type Registry struct {
	ordered []PipelineStep
	byName  map[string]PipelineStep
}

func NewRegistry() *Registry {
	return &Registry{
		ordered: make([]PipelineStep, 0),
		byName:  make(map[string]PipelineStep),
	}
}

func (r *Registry) Register(step PipelineStep) error {
	if step == nil {
		return fmt.Errorf("pipeline step is nil")
	}
	name := step.Name()
	if name == "" {
		return fmt.Errorf("pipeline step name is empty")
	}
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("pipeline step already registered: %s", name)
	}
	r.byName[name] = step
	r.ordered = append(r.ordered, step)
	return nil
}

func (r *Registry) Get(name string) (PipelineStep, bool) {
	step, ok := r.byName[name]
	return step, ok
}

func (r *Registry) Ordered() []PipelineStep {
	out := make([]PipelineStep, len(r.ordered))
	copy(out, r.ordered)
	return out
}

func (r *Registry) StepNames() []string {
	names := make([]string, 0, len(r.ordered))
	for _, step := range r.ordered {
		names = append(names, step.Name())
	}
	return names
}

func (r *Registry) NextStep(current string) (string, bool) {
	names := r.StepNames()
	for index, name := range names {
		if name == current && index+1 < len(names) {
			return names[index+1], true
		}
	}
	return "", false
}

// RegisterDefaultSteps wires the standard pipeline using the provided executor.
func RegisterDefaultSteps(registry *Registry, executor DocumentExecutor) error {
	steps := []PipelineStep{
		newExecutorStep(StepExtractText, executor, func(ctx context.Context, job *PipelineJob) error {
			return executor.ExtractText(ctx, job)
		}, func(ctx context.Context, job *PipelineJob) bool {
			supported, err := executor.IsTextExtractionSupported(job.DocumentID, job.TenantID)
			return err == nil && supported
		}),
		newExecutorStep(StepChunkDocument, executor, executor.GenerateChunks, alwaysCanRun),
		newExecutorStep(StepGenerateEmbedding, executor, executor.GenerateEmbeddings, alwaysCanRun),
		newExecutorStep(StepGenerateSummary, executor, executor.GenerateSummary, alwaysCanRun),
		newExecutorStep(StepClassifyDocument, executor, executor.ClassifyDocument, alwaysCanRun),
		newExecutorStep(StepExtractMetadata, executor, executor.ExtractMetadata, alwaysCanRun),
		newExecutorStep(StepExtractDates, executor, executor.ExtractImportantDates, alwaysCanRun),
		newExecutorStep(StepGenerateTags, executor, executor.SuggestTags, alwaysCanRun),
		newExecutorStep(StepExtractEntities, executor, executor.ExtractEntities, alwaysCanRun),
		newExecutorStep(StepLinkRelatedDocuments, executor, executor.LinkRelatedDocuments, alwaysCanRun),
		newExecutorStep(StepCompleted, executor, executor.FinalizePipeline, alwaysCanRun),
	}

	for _, step := range steps {
		if err := registry.Register(step); err != nil {
			return err
		}
	}
	return nil
}

func alwaysCanRun(context.Context, *PipelineJob) bool { return true }

type executorStep struct {
	name    string
	execute func(context.Context, *PipelineJob) error
	canRun  func(context.Context, *PipelineJob) bool
}

func newExecutorStep(
	name string,
	_ DocumentExecutor,
	execute func(context.Context, *PipelineJob) error,
	canRun func(context.Context, *PipelineJob) bool,
) PipelineStep {
	if canRun == nil {
		canRun = alwaysCanRun
	}
	return &executorStep{name: name, execute: execute, canRun: canRun}
}

func (s *executorStep) Name() string { return s.name }

func (s *executorStep) Execute(ctx context.Context, job *PipelineJob) error {
	return s.execute(ctx, job)
}

func (s *executorStep) CanRun(ctx context.Context, job *PipelineJob) bool {
	return s.canRun(ctx, job)
}

// SortStepNames returns names sorted by default pipeline order.
func SortStepNames(names []string) []string {
	orderIndex := make(map[string]int, len(DefaultStepOrder))
	for index, name := range DefaultStepOrder {
		orderIndex[name] = index
	}
	sorted := append([]string(nil), names...)
	sort.Slice(sorted, func(i, j int) bool {
		return orderIndex[sorted[i]] < orderIndex[sorted[j]]
	})
	return sorted
}

package pipeline

import "testing"

func TestRegistryNextStep(t *testing.T) {
	registry := NewRegistry()
	_ = registry.Register(&noopStep{name: StepExtractText})
	_ = registry.Register(&noopStep{name: StepChunkDocument})

	next, ok := registry.NextStep(StepExtractText)
	if !ok || next != StepChunkDocument {
		t.Fatalf("NextStep() = %q, %v; want chunk_document, true", next, ok)
	}
}

func TestValidateJobRequiresIDs(t *testing.T) {
	if err := ValidateJob(&PipelineJob{}); err == nil {
		t.Fatal("expected validation error for empty job")
	}
}

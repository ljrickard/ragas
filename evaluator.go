package ragas

import (
	"context"
)

// LLMClient is the interface any provider must satisfy
type LLMClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

// Evaluator holds the injected client
type Evaluator struct {
	client LLMClient
}

// NewEvaluator strictly takes only the interface
func NewEvaluator(client LLMClient) *Evaluator {
	return &Evaluator{
		client: client,
	}
}

// EvalResult standardizes the JSON output from the judge models
type EvalResult struct {
	Score     float64 `json:"score"`
	Reasoning string  `json:"reasoning"`
}

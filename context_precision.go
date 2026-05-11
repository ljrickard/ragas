package ragas

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const precisionPrompt = `You are an impartial judge evaluating a RAG AI system.
Given the following Question, Ground Truth, and a list of retrieved Context chunks, determine which contexts are relevant to the Ground Truth.

Question: %s
Ground Truth: %s
Contexts:
%s

Instructions:
1. For each context chunk, determine if it is "Relevant" or "Not Relevant" to answering the Question based on the Ground Truth.
2. Provide a relevance score for each chunk (1 for Relevant, 0 for Not Relevant).
3. Calculate the Precision@K: (Sum of relevance scores) / (Total number of chunks).
4. Respond ONLY with a JSON object in this format: {"score": 0.80, "reasoning": "Chunks 1 and 2 were highly relevant, but chunk 3 was about a different episode."}`

/*
EvaluateContextPrecision
------------------------------------------------
What it measures: Did the vector database put the most relevant information at the top?
What it compares: Retrieved Contexts vs. Ground Truth.
*/
func (e *Evaluator) EvaluateContextPrecision(ctx context.Context, question string, contexts []string, groundTruth string) (float64, string, error) {
	// Format the contexts into a numbered list for the judge
	var contextList strings.Builder
	for i, c := range contexts {
		contextList.WriteString(fmt.Sprintf("[%d] %s\n", i+1, c))
	}

	prompt := fmt.Sprintf(precisionPrompt, question, groundTruth, contextList.String())

	responseJSON, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return 0, "", fmt.Errorf("llm evaluation failed: %w", err)
	}

	// Standard Ragas JSON cleaning logic
	cleanJSON := strings.TrimSpace(responseJSON)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json\n")
	cleanJSON = strings.TrimSuffix(cleanJSON, "\n```")
	cleanJSON = strings.TrimSpace(cleanJSON)

	var result EvalResult
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return 0, "", fmt.Errorf("failed to parse judge response: %w, raw output: %s", err, cleanJSON)
	}

	return result.Score, result.Reasoning, nil
}

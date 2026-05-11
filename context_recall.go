package ragas

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const recallPrompt = `You are an impartial judge evaluating a RAG AI system.
Your goal is to measure Context Recall: Did the retrieved Contexts contain all the necessary information to reconstruct the Ground Truth answer?

Question: %s
Ground Truth: %s

Retrieved Contexts:
%s

Instructions:
1. Break the Ground Truth answer down into individual, distinct factual statements.
2. For each factual statement, thoroughly scan the Retrieved Contexts to see if that fact is present, supported, or can be directly inferred.
3. Calculate the Context Recall score: (Number of supported statements) / (Total number of statements in the Ground Truth).
4. Respond ONLY with a JSON object in this format: {"score": 0.66, "reasoning": "The contexts supported the fact that Frasier was married to Lilith, but failed to contain any mention of Nanette Guzman."}`

/*
EvaluateContextRecall
------------------------------------------------
What it measures: Did the search retrieve ALL the necessary facts?
What it compares: Ground Truth vs. Retrieved Contexts.
*/
func (e *Evaluator) EvaluateContextRecall(ctx context.Context, question string, contexts []string, groundTruth string) (float64, string, error) {
	// Format the contexts into a numbered list for the judge
	var contextList strings.Builder
	for i, c := range contexts {
		contextList.WriteString(fmt.Sprintf("[%d] %s\n", i+1, c))
	}

	prompt := fmt.Sprintf(recallPrompt, question, groundTruth, contextList.String())

	responseJSON, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return 0, "", fmt.Errorf("llm evaluation failed: %w", err)
	}

	// Clean Markdown formatting if the LLM includes it
	cleanJSON := strings.TrimSpace(responseJSON)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json\n")
	cleanJSON = strings.TrimSuffix(cleanJSON, "\n```")
	cleanJSON = strings.TrimSpace(cleanJSON)

	// Assuming you have this struct defined in evaluator.go or similar
	var result struct {
		Score     float64 `json:"score"`
		Reasoning string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return 0, "", fmt.Errorf("failed to parse judge response: %w, raw output: %s", err, cleanJSON)
	}

	return result.Score, result.Reasoning, nil
}

package ragas

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const correctnessPrompt = `You are an impartial judge evaluating a RAG AI system.
Your goal is to measure Answer Correctness: Does the generated Answer completely and accurately align with the Ground Truth?

Question: %s
Ground Truth: %s
Generated Answer: %s

Instructions:
1. Break down the Ground Truth into core factual statements.
2. Check if the Generated Answer contains those exact facts without contradicting them.
3. Penalize the score if the Generated Answer misses critical facts from the Ground Truth or introduces contradictory information. (Additional, non-contradictory context is fine).
4. Calculate a Correctness score from 0.0 (completely wrong/unrelated) to 1.0 (perfectly correct and complete).
5. Respond ONLY with a JSON object in this format: {"score": 0.85, "reasoning": "The answer correctly identified Lilith, but missed Nanette Guzman."}`

/*
EvaluateAnswerCorrectness
------------------------------------------------
What it measures: Did the final generated answer match the factual reality of the Ground Truth?
What it compares: Generated Answer vs. Ground Truth.
*/
func (e *Evaluator) EvaluateAnswerCorrectness(ctx context.Context, question string, generatedAnswer string, groundTruth string) (float64, string, error) {
	prompt := fmt.Sprintf(correctnessPrompt, question, groundTruth, generatedAnswer)

	responseJSON, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return 0, "", fmt.Errorf("llm evaluation failed: %w", err)
	}

	// Clean Markdown formatting if the LLM includes it
	cleanJSON := strings.TrimSpace(responseJSON)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json\n")
	cleanJSON = strings.TrimSuffix(cleanJSON, "\n```")
	cleanJSON = strings.TrimSpace(cleanJSON)

	var result struct {
		Score     float64 `json:"score"`
		Reasoning string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return 0, "", fmt.Errorf("failed to parse judge response: %w, raw output: %s", err, cleanJSON)
	}

	return result.Score, result.Reasoning, nil
}

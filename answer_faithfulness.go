package ragas

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const faithfulnessPrompt = `You are an impartial judge evaluating a RAG AI system.
Given the following Question, retrieved Context, and the AI's Answer, calculate the Faithfulness score.

Question: %s
Context: %s
Answer: %s

Instructions:
1. Extract all factual statements made in the Answer.
2. For each statement, determine if it can be directly inferred from the Context (Yes/No).
3. Calculate the score: (Number of 'Yes' statements) / (Total statements).
4. Respond ONLY with a JSON object in this format: {"score": 0.85, "reasoning": "statement 1 was supported, statement 2 was not."}`

/*
EvaluateFaithfulness (Checks for Hallucinations)
------------------------------------------------
What it measures: Is the generated answer factually grounded in the retrieved context?
What it compares: Answer vs. Retrieved Context.

How it works: It penalizes the LLM if it makes up facts (hallucinates) or uses outside
knowledge not present in your database. A score of 1.0 means every single factual statement
in the answer can be directly traced back to the retrieved chunks.
*/
func (e *Evaluator) EvaluateFaithfulness(ctx context.Context, question string, contexts []string, answer string) (float64, string, error) {
	joinedContext := strings.Join(contexts, "\n---\n")
	prompt := fmt.Sprintf(faithfulnessPrompt, question, joinedContext, answer)

	responseJSON, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return 0, "", fmt.Errorf("llm evaluation failed: %w", err)
	}

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

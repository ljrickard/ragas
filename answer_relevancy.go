package ragas

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const relevancyPrompt = `You are an impartial judge evaluating a RAG AI system.
Given the following Question and the AI's Answer, calculate the Answer Relevancy score.

Question: %s
Answer: %s

Instructions:
1. Assess how directly and comprehensively the Answer addresses the Question.
2. Penalize the score for incomplete answers or redundant, tangential information.
3. Calculate the score on a scale from 0.0 (completely irrelevant) to 1.0 (perfectly relevant and concise).
4. Respond ONLY with a JSON object in this format: {"score": 0.85, "reasoning": "The answer directly addressed the question but included an unnecessary paragraph about a different topic."}`

/*
EvaluateAnswerRelevancy
------------------------------------------------
What it measures: How directly and concisely the answer addresses the initial question.
What it compares: Answer vs. Question.

How it works: It penalizes the LLM if it provides a factually correct but tangential
answer, or if it writes a massive paragraph to answer a simple yes/no question.
A score of 1.0 means the answer perfectly and concisely addressed the user's prompt.
*/
func (e *Evaluator) EvaluateAnswerRelevancy(ctx context.Context, question string, answer string) (float64, string, error) {
	prompt := fmt.Sprintf(relevancyPrompt, question, answer)

	responseJSON, err := e.client.GenerateContent(ctx, prompt)
	if err != nil {
		return 0, "", fmt.Errorf("llm evaluation failed: %w", err)
	}

	// Clean the response (LLMs love to wrap JSON in markdown blocks)
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

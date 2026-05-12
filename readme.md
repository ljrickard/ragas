# ragas-go 🎯

A lightweight, fully decoupled Go library for evaluating Retrieval-Augmented Generation (RAG) pipelines using the "LLM-as-a-judge" pattern. 

Inspired by the popular Python `ragas` framework, this package brings quantitative RAG evaluation natively to Go backend services, allowing you to integrate AI quality checks directly into your CI/CD pipelines or Go test suites.

## Overview
Testing non-deterministic AI systems is notoriously difficult. Traditional unit tests fail when outputs are dynamically generated. `ragas-go` solves this by using a stronger LLM to grade your pipeline's outputs against a known ground truth, scoring them mathematically from `0.0` to `1.0`.

This library isolates the **evaluation logic** from the **LLM provider**, allowing you to plug in any model (OpenAI, Gemini, Anthropic, or local models via Ollama) to act as the judge.

## 📊 Supported Evaluation Metrics

This package breaks RAG evaluation into two distinct phases: **Retrieval** (Vector DB performance) and **Generation** (LLM performance).

### Retrieval Metrics
* **Context Precision:** Did the vector database put the most highly relevant chunks at the very top of the results? (Penalizes burying the answer in noise).
* **Context Recall:** Did the vector database manage to retrieve *all* the necessary facts required to reconstruct the ground truth?

### Generation Metrics
* **Answer Faithfulness:** Is the generated answer completely backed by the retrieved context? (A pure hallucination check).
* **Answer Relevancy:** How concisely and directly did the LLM answer the user's prompt? (Penalizes rambling or tangential answers).
* **Answer Correctness:** Does the final generated answer factually align with the known ground truth?

## 🚀 Installation

```bash
go get github.com/ljrickard/ragas@v0.1.2
```

## 🛠️ Usage & Architecture

### 1. The Interface Design
To maintain a clean architecture, `ragas-go` does not hardcode any specific LLM SDKs. Instead, it relies on a simple interface. You just need to pass it a client that knows how to turn a prompt string into a response string:

```go
// The only interface you need to satisfy
type LLMClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}
```

### 2. Basic Example
Here is how you might use it in a Go test suite to grade a query:

```go
package main

import (
	"context"
	"fmt"
	"github.com/ljrickard/ragas"
)

// 1. Wrap your preferred LLM (e.g., Gemini, OpenAI) to satisfy the interface
type MyLLMJudge struct {}

func (m *MyLLMJudge) GenerateContent(ctx context.Context, prompt string) (string, error) {
	// Call your LLM API here and return the raw text
	return `{"score": 1.0, "reasoning": "The answer was perfect."}`, nil
}

func main() {
	ctx := context.Background()
	
	// 2. Instantiate the evaluator
	evaluator := ragas.NewEvaluator(&MyLLMJudge{})

	// 3. Define your test data
	question := "Who was Frasier married to?"
	groundTruth := "Frasier Crane was married to Lilith Sternin and Nanette Guzman."
	answer := "Frasier was married to Lilith and Nanette."
	contexts := []string{
        "Frasier mentions his ex-wife Lilith.",
        "Nanette Guzman, also known as Nanny G, was Frasier's first wife.",
    }

	// 4. Run your metrics!
	score, reasoning, err := evaluator.EvaluateAnswerCorrectness(ctx, question, answer, groundTruth)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Correctness Score: %.2f\nReasoning: %s\n", score, reasoning)
}
```

## 🏗️ Design Philosophy
This library is designed with standard Go idioms in mind:
* **Zero Dependencies:** Beyond standard library JSON parsing and context management, the package is exceptionally lightweight.
* **Thread-Safe:** The `Evaluator` struct holds no state other than the injected client, making it safe for parallel test execution (`t.Parallel()`).
* **Clean JSON Parsing:** LLMs often wrap JSON outputs in markdown code blocks. The internal parsing logic aggressively cleans and sanitizes the output before unmarshaling, preventing pipeline crashes from overly chatty judge models.

## License
MIT
// Package service contains the Explainer which calls a local Ollama instance
// running Llama 3.2 to generate plain-English code explanations.
//
// Ollama runs at http://localhost:11434 by default.
// The streaming endpoint is POST /api/chat
//
// NewExplainer(ollamaURL string) *Explainer
//   - ollamaURL: base URL of the Ollama instance, e.g. "http://localhost:11434"
//   - No API key needed — Ollama is unauthenticated locally
//
// ExplainStream(
//   ctx      context.Context,
//   code     string,
//   language string,
//   mode     string,
//   onChunk  func(chunk string) error,
// ) error
//
// Request body to POST /api/chat:
//   {
//     "model": "llama3.2",
//     "stream": true,
//     "messages": [
//       {
//         "role": "system",
//         "content": "You are a code explanation assistant. You only explain
//                     what code does. You never modify, fix, or generate code.
//                     Be concise and use plain English."
//       },
//       {
//         "role": "user",
//         "content": ""
//       }
//     ]
//   }
//
// User message content rules:
//   - If mode == "summary":
//       "Explain in 2-4 sentences what this {language} code does:\n\n{code}"
//   - If mode == "line-by-line":
//       "Explain this {language} code line by line or block by block:\n\n{code}"
//
// Streaming response:
//   Ollama streams newline-delimited JSON objects. Each line looks like:
//   {"model":"llama3.2","message":{"role":"assistant","content":"chunk"},"done":false}
//
//   Parse each line as JSON. Extract message.content as the text chunk.
//   Call onChunk(content) for each line where done == false.
//   Stop when done == true.
//   If onChunk returns an error, stop and return it.
//   Honour ctx cancellation throughout.
//
// Use encoding/json to marshal the request and unmarshal each response line.
// Use bufio.NewReader to read the streaming response line by line.

package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Explainer struct {
	ollamaURL string
	client    *http.Client
	model     string
}

func NewExplainer(ollamaURL string) *Explainer {
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2"
	}
	return &Explainer{
		ollamaURL: strings.TrimRight(ollamaURL, "/"),
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		model: model,
	}
}
func (e *Explainer) ExplainStream(ctx context.Context, code, language, mode string, onChunk func(string) error) error {
	if mode != "summary" && mode != "line-by-line" {
		return errors.New("invalid mode: must be 'summary' or 'line-by-line'")
	}
	prompt := fmt.Sprintf("Explain in %s what this %s code does:\n\n%s",
		func() string {
			if mode == "summary" {
				return "2-4 sentences"
			}
			return "line by line or block by block"
		}(),
		language,
		code,
	)
	reqBody := map[string]interface{}{
		"model":  e.model,
		"stream": true,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a code explanation assistant. You only explain what code does. You never modify, fix, or generate code. Be concise and use plain English.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", e.ollamaURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("request to Ollama failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Ollama non-200 response: %d - %s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var respObj struct {
			Model   string `json:"model"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &respObj); err != nil {
			return fmt.Errorf("failed to parse response line as JSON: %w", err)
		}
		if !respObj.Done {
			if err := onChunk(respObj.Message.Content); err != nil {
				return fmt.Errorf("onChunk callback returned an error: %w", err)
			}
		} else {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}
	return nil
}

package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// GeminiProvider implements LLM provider using Google Gemini API
type GeminiProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	model   string
}

func NewGeminiProvider() (*GeminiProvider, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	return &GeminiProvider{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "gemini-2.5-flash",
	}, nil
}

func (g *GeminiProvider) ExtractJSON(input string) ([]byte, error) {
	// Build the prompt for structured extraction
	prompt := buildGeminiExtractionPrompt(input)

	// Prepare the request body
	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]any{
			"temperature":     0.1,
			"maxOutputTokens": 500,
			"responseMimeType": "application/json",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the URL with API key
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", g.baseURL, g.model, g.apiKey)

	// Make the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Extract JSON from response
	content := geminiResp.Candidates[0].Content.Parts[0].Text
	
	// Remove markdown code blocks if present
	contentBytes := []byte(content)
	contentBytes = bytes.TrimSpace(contentBytes)
	if bytes.HasPrefix(contentBytes, []byte("```json")) {
		contentBytes = bytes.TrimPrefix(contentBytes, []byte("```json"))
		contentBytes = bytes.TrimSuffix(contentBytes, []byte("```"))
	} else if bytes.HasPrefix(contentBytes, []byte("```")) {
		contentBytes = bytes.TrimPrefix(contentBytes, []byte("```"))
		contentBytes = bytes.TrimSuffix(contentBytes, []byte("```"))
	}
	contentBytes = bytes.TrimSpace(contentBytes)
	
	// Try to extract JSON if wrapped in text
	contentStr := string(contentBytes)
	if idx := strings.Index(contentStr, "{"); idx >= 0 {
		if endIdx := strings.LastIndex(contentStr, "}"); endIdx > idx {
			contentBytes = []byte(contentStr[idx : endIdx+1])
		}
	}
	
	return contentBytes, nil
}

func buildGeminiExtractionPrompt(input string) string {
	return fmt.Sprintf(`You are a structured data extraction assistant for a personal task management system. Extract information from the user's natural language input and return ONLY valid JSON.

Input: "%s"

Return a JSON object with this exact structure:
{
  "type": "log" | "progress" | "commitment" | "update" | "event" | "correction",
  "confidence": 0.0-1.0,
  "data": {
    // Fields depend on type:
    // - commitment: { "person": string, "project": string (optional), "expectation": { "description": string, "deadline": "YYYY-MM-DD", "hardness": "hard"|"soft" } }
    // - progress: { "project": string, "status": string (optional), "notes": string (optional) }
    // - update: { "commitment_id": string (optional), "person": string (optional), "project": string (optional), "status": string }
    // - event: { "time": "YYYY-MM-DD HH:MM" or "YYYY-MM-DD", "person": string (optional), "project": string (optional), "title": string }
    // - log: { "text": string }
    // - correction: { "text": string }
  },
  "questions": [] // Array of questions if information is missing/ambiguous
}

Guidelines:
- If it's a commitment (told someone, promised, will do, said I'll), use type "commitment"
- If it's progress update (done, finished, completed, made progress), use type "progress" or "update"
- If it's scheduling (meet, call, event, appointment), use type "event"
- If it's a correction (that's wrong, actually, correction), use type "correction"
- Otherwise, use type "log"
- Extract dates relative to today if needed (tomorrow = today + 1 day, next week = today + 7 days)
- Be confident (>= 0.7) if you're sure, lower if uncertain
- Include questions array if key info is missing (e.g., missing person, deadline, project)
- For commitments, infer project from context if mentioned (caresad, farmer, kaifu, soldecoder, etc.)

Return ONLY the JSON object, no other text.`, input)
}

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

func (g *GeminiProvider) ExtractJSON(input string, now time.Time) ([]byte, error) {
	// Build the prompt for structured extraction
	prompt := buildGeminiExtractionPrompt(input, now)

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
			"maxOutputTokens": 2000,
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
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Check if response was truncated
	if geminiResp.Candidates[0].FinishReason == "MAX_TOKENS" {
		return nil, fmt.Errorf("response truncated due to token limit (increase maxOutputTokens)")
	}

	// Extract JSON from response
	content := geminiResp.Candidates[0].Content.Parts[0].Text
	rawContent := content // Keep original for error messages
	
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
	// Use balanced brace matching for more robust extraction
	contentStr := string(contentBytes)
	startIdx := strings.Index(contentStr, "{")
	if startIdx >= 0 {
		// Find matching closing brace by counting braces
		braceCount := 0
		endIdx := -1
		for i := startIdx; i < len(contentStr); i++ {
			if contentStr[i] == '{' {
				braceCount++
			} else if contentStr[i] == '}' {
				braceCount--
				if braceCount == 0 {
					endIdx = i
					break
				}
			}
		}
		if endIdx > startIdx {
			contentBytes = []byte(contentStr[startIdx : endIdx+1])
		} else {
			// JSON appears incomplete - return error with full context
			rawContentStr := rawContent
			if len(rawContentStr) > 1000 {
				rawContentStr = rawContentStr[:1000] + "... (truncated for display)"
			}
			return nil, fmt.Errorf("incomplete JSON response (unmatched braces, found %d open)\n  extracted so far: %q\n  full response: %q", braceCount, string(contentBytes[startIdx:]), rawContentStr)
		}
	}
	
	// Validate that we have some content
	if len(contentBytes) == 0 {
		rawContentStr := rawContent
		if len(rawContentStr) > 1000 {
			rawContentStr = rawContentStr[:1000] + "... (truncated for display)"
		}
		return nil, fmt.Errorf("no JSON content extracted from LLM response\n  raw response: %q", rawContentStr)
	}
	
	// Validate JSON is complete by attempting to parse it
	var testJSON map[string]any
	if err := json.Unmarshal(contentBytes, &testJSON); err != nil {
		rawContentStr := rawContent
		if len(rawContentStr) > 1000 {
			rawContentStr = rawContentStr[:1000] + "... (truncated for display)"
		}
		return nil, fmt.Errorf("extracted JSON is invalid: %w\n  raw response: %q", err, rawContentStr)
	}
	
	return contentBytes, nil
}

func buildGeminiExtractionPrompt(input string, now time.Time) string {
	today := now.Format("2006-01-02")
	dayOfWeek := now.Format("Monday")
	currentTime := now.Format("15:04")
	dateReadable := now.Format("January 2, 2006")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	nextWeek := now.AddDate(0, 0, 7).Format("2006-01-02")
	
	return fmt.Sprintf(`You are a structured data extraction assistant for a personal task management system. Extract information from the user's natural language input and return ONLY valid JSON.

Current context:
- Date: %s (%s)
- Day of week: %s
- Current time: %s
- Date (readable): %s

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
- Extract dates relative to today (%s, %s) if needed:
  * "today" = %s
  * "tomorrow" = %s
  * "next week" = %s
  * "next few hours" or "later today" = %s (same day, use current time %s to determine if still possible)
  * "this week" = calculate based on current day (%s) - week runs Monday-Sunday
  * "next Monday/Tuesday/etc" = next occurrence of that day
  * Always use YYYY-MM-DD format for dates
- Be confident (>= 0.7) if you're sure, lower if uncertain
- Include questions array if key info is missing (e.g., missing person, deadline, project)
- For commitments, infer project from context if mentioned (kaifu, octomod, etc.)

Return ONLY the JSON object, no other text.`, today, dayOfWeek, dayOfWeek, currentTime, dateReadable, input, today, dayOfWeek, today, tomorrow, nextWeek, today, currentTime, dayOfWeek)
}

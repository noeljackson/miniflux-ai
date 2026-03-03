package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type CurationResult struct {
	Summary   string   `json:"summary"`
	Tags      []string `json:"tags"`
	Relevance int      `json:"relevance"`
	Reason    string   `json:"reason"`
}

type Curator interface {
	Curate(title, content, url string) (*CurationResult, error)
}

type ClaudeCurator struct {
	client *anthropic.Client
}

func NewClaudeCurator(apiKey string) *ClaudeCurator {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &ClaudeCurator{client: &client}
}

func (c *ClaudeCurator) Curate(title, content, url string) (*CurationResult, error) {
	cleanContent := truncate(content, 4000)

	prompt := fmt.Sprintf(`Analyze this RSS feed entry and return a JSON object with these fields:
- "summary": 1-2 sentence summary of the key points
- "tags": array of 2-5 topic tags (lowercase)
- "reason": 1 sentence explaining why this article might be interesting

Title: %s
URL: %s
Content: %s

Return ONLY valid JSON, no markdown fences or other text.`, title, url, cleanContent)

	message, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 512,
		System: []anthropic.TextBlockParam{
			{Text: "You are an RSS feed curator. Return only valid JSON."},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}

	var responseText string
	for _, block := range message.Content {
		if b, ok := block.AsAny().(anthropic.TextBlock); ok {
			responseText = b.Text
			break
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("empty response from Claude")
	}

	return parseCurationResult(responseText)
}

func parseCurationResult(text string) (*CurationResult, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result CurationResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse curation JSON: %w (raw: %s)", err, text)
	}

	if result.Relevance < 0 {
		result.Relevance = 0
	}
	if result.Relevance > 100 {
		result.Relevance = 100
	}

	return &result, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

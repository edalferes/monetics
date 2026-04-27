package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/edalferes/monetics/internal/config"
	"github.com/edalferes/monetics/pkg/logger"
)

// Pricing per 1M tokens (USD). Used only for an estimated cost in responses.
// Values are approximations and intentionally conservative.
var modelPricing = map[string]struct {
	InputUSDPerM  float64
	OutputUSDPerM float64
}{
	"gpt-4o-mini": {InputUSDPerM: 0.150, OutputUSDPerM: 0.600},
	"gpt-4o":      {InputUSDPerM: 2.500, OutputUSDPerM: 10.000},
}

// OpenAIClient implements Client against the OpenAI Chat Completions API.
type OpenAIClient struct {
	httpClient *http.Client
	cfg        config.AIConfig
	logger     logger.Logger
}

// NewOpenAIClient constructs an OpenAI-backed Client.
func NewOpenAIClient(cfg config.AIConfig, log logger.Logger) *OpenAIClient {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &OpenAIClient{
		httpClient: &http.Client{Timeout: timeout},
		cfg:        cfg,
		logger:     log.With().Str("component", "ai.openai").Logger(),
	}
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIRequest struct {
	Model          string               `json:"model"`
	Messages       []openAIMessage      `json:"messages"`
	Temperature    float64              `json:"temperature"`
	ResponseFormat openAIResponseFormat `json:"response_format"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type suggestionsPayload struct {
	Suggestions []CategorySuggestion `json:"suggestions"`
}

// Suggest sends a chat completion request and parses the structured JSON reply.
func (c *OpenAIClient) Suggest(ctx context.Context, req SuggestRequest) (SuggestResponse, error) {
	userPrompt, err := buildUserPrompt(req)
	if err != nil {
		return SuggestResponse{}, err
	}

	payload := openAIRequest{
		Model: c.cfg.Model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature:    0,
		ResponseFormat: openAIResponseFormat{Type: "json_object"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return SuggestResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return SuggestResponse{}, fmt.Errorf("build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return SuggestResponse{}, fmt.Errorf("call openai: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return SuggestResponse{}, fmt.Errorf("read openai body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Avoid logging the API key; safe to log status + truncated body.
		preview := string(respBytes)
		if len(preview) > 300 {
			preview = preview[:300]
		}
		c.logger.Error().Int("status", resp.StatusCode).Str("body", preview).Msg("openai returned non-2xx")
		return SuggestResponse{}, fmt.Errorf("openai returned status %d", resp.StatusCode)
	}

	var oai openAIResponse
	if err := json.Unmarshal(respBytes, &oai); err != nil {
		return SuggestResponse{}, fmt.Errorf("decode openai response: %w", err)
	}
	if len(oai.Choices) == 0 {
		return SuggestResponse{}, fmt.Errorf("openai returned no choices")
	}

	var parsed suggestionsPayload
	if err := json.Unmarshal([]byte(oai.Choices[0].Message.Content), &parsed); err != nil {
		return SuggestResponse{}, fmt.Errorf("decode suggestions content: %w", err)
	}

	return SuggestResponse{
		Suggestions: parsed.Suggestions,
		Usage: Usage{
			PromptTokens:     oai.Usage.PromptTokens,
			CompletionTokens: oai.Usage.CompletionTokens,
			EstimatedCostUSD: estimateCostUSD(c.cfg.Model, oai.Usage.PromptTokens, oai.Usage.CompletionTokens),
		},
	}, nil
}

// estimateCostUSD returns an approximation; unknown models return 0.
func estimateCostUSD(model string, prompt, completion int) float64 {
	p, ok := modelPricing[model]
	if !ok {
		return 0
	}
	return (float64(prompt)*p.InputUSDPerM + float64(completion)*p.OutputUSDPerM) / 1_000_000.0
}

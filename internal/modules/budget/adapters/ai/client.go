// Package ai provides clients for AI-assisted features (e.g., smart categorization).
package ai

import (
	"context"
)

// CategorySuggestion is the LLM's proposal for a single transaction.
type CategorySuggestion struct {
	RowIndex   int     `json:"row_index"`
	CategoryID uint    `json:"category_id"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason,omitempty"`
}

// Usage represents token usage and estimated cost from the provider.
type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
}

// Client defines the interface for smart categorization through LLMs.
type Client interface {
	Suggest(ctx context.Context, req SuggestRequest) (SuggestResponse, error)
}

// SuggestRequest contains all data needed by the LLM for categorization.
type SuggestRequest struct {
	Items      []TransactionRef `json:"items"`
	Categories []CategoryRef    `json:"categories"`
	History    []HistoryRef     `json:"history,omitempty"`
}

// SuggestResponse aggregates results from the AI provider.
type SuggestResponse struct {
	Suggestions []CategorySuggestion `json:"suggestions"`
	Usage       Usage                `json:"usage"`
}

// TransactionRef is a simplified transaction for the prompt.
type TransactionRef struct {
	RowIndex    int     `json:"row_index"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Date        string  `json:"date,omitempty"`
}

// CategoryRef is a simplified category for the prompt.
type CategoryRef struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// HistoryRef is a past example of categorization.
type HistoryRef struct {
	Description string `json:"description"`
	CategoryID  uint   `json:"category_id"`
}

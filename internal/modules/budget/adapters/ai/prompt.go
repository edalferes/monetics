package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// systemPrompt is fixed and provider-agnostic.
const systemPrompt = `You are a financial transaction categorizer. ` +
	`Given a list of bank/credit card transactions and a user's category catalog, ` +
	`assign the most appropriate category_id to each transaction. ` +
	`Use the provided history examples as preference signals. ` +
	`Match category type (expense/income/transfer) to transaction type. ` +
	`If unsure, return category_id=0 and a low confidence value. ` +
	`Output strictly the JSON schema requested. Do not invent category IDs.`

// buildUserPrompt assembles the user-facing prompt content.
// Descriptions are sanitized to avoid prompt injection / control chars.
func buildUserPrompt(req SuggestRequest) (string, error) {
	var b strings.Builder

	b.WriteString("CATEGORIES:\n")
	for _, c := range req.Categories {
		fmt.Fprintf(&b, "- id=%d name=%q type=%s\n", c.ID, sanitize(c.Name), c.Type)
	}

	if len(req.History) > 0 {
		b.WriteString("\nRECENT EXAMPLES (description -> category_id):\n")
		for _, h := range req.History {
			fmt.Fprintf(&b, "- %q -> %d\n", sanitize(h.Description), h.CategoryID)
		}
	}

	items := make([]TransactionRef, len(req.Items))
	for i, it := range req.Items {
		it.Description = sanitize(it.Description)
		items[i] = it
	}
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("marshal items: %w", err)
	}

	b.WriteString("\nTRANSACTIONS TO CATEGORIZE:\n")
	b.Write(itemsJSON)
	b.WriteString("\n\nReturn JSON: {\"suggestions\":[{\"row_index\":int,\"category_id\":int,\"confidence\":number,\"reason\":string}]}")

	return b.String(), nil
}

// sanitize strips control characters and trims whitespace.
func sanitize(s string) string {
	s = strings.TrimSpace(s)
	var out strings.Builder
	out.Grow(len(s))
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			out.WriteRune(' ')
			continue
		}
		if r < 32 || r == 127 {
			continue
		}
		out.WriteRune(r)
	}
	// collapse repeated spaces
	return strings.Join(strings.Fields(out.String()), " ")
}

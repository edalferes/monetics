package transaction

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/edalferes/monetics/internal/config"
	"github.com/edalferes/monetics/internal/modules/budget/adapters/ai"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/errors"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
	"github.com/edalferes/monetics/pkg/timex"
)

// SuggestCategoriesInput is the use case input.
type SuggestCategoriesInput struct {
	UserID    uint
	AccountID uint
	Items     []ImportItem
}

// SuggestCategoriesOutput is the use case output.
type SuggestCategoriesOutput struct {
	Suggestions []ai.CategorySuggestion `json:"suggestions"`
	Usage       ai.Usage                `json:"usage"`
}

// SuggestCategoriesUseCase asks an AI provider for category suggestions
// without persisting any transactions.
type SuggestCategoriesUseCase struct {
	aiClient        ai.Client
	accountRepo     interfaces.AccountRepository
	categoryRepo    interfaces.CategoryRepository
	transactionRepo interfaces.TransactionRepository
	logger          logger.Logger
	cfg             config.AIConfig
}

// NewSuggestCategoriesUseCase creates a new SuggestCategoriesUseCase.
func NewSuggestCategoriesUseCase(
	aiClient ai.Client,
	accountRepo interfaces.AccountRepository,
	categoryRepo interfaces.CategoryRepository,
	transactionRepo interfaces.TransactionRepository,
	cfg config.AIConfig,
	log logger.Logger,
) *SuggestCategoriesUseCase {
	return &SuggestCategoriesUseCase{
		aiClient:        aiClient,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
		cfg:             cfg,
		logger:          log.With().Str("usecase", "transaction.suggest_categories").Logger(),
	}
}

// Execute returns category suggestions for the provided rows.
func (uc *SuggestCategoriesUseCase) Execute(ctx context.Context, in SuggestCategoriesInput) (SuggestCategoriesOutput, error) {
	if in.UserID == 0 {
		return SuggestCategoriesOutput{}, errors.ErrInvalidUserID
	}
	if len(in.Items) == 0 {
		return SuggestCategoriesOutput{}, errors.ErrInvalidData
	}
	if uc.cfg.MaxItemsPerRequest > 0 && len(in.Items) > uc.cfg.MaxItemsPerRequest {
		return SuggestCategoriesOutput{}, fmt.Errorf("too many items: %d > %d", len(in.Items), uc.cfg.MaxItemsPerRequest)
	}

	// Account belongs to user.
	acc, err := uc.accountRepo.GetByID(ctx, in.AccountID)
	if err != nil {
		return SuggestCategoriesOutput{}, errors.ErrAccountNotFound
	}
	if acc.UserID != in.UserID {
		return SuggestCategoriesOutput{}, errors.ErrUnauthorizedAccess
	}

	// Load user categories.
	cats, err := uc.categoryRepo.GetByUserID(ctx, in.UserID)
	if err != nil {
		return SuggestCategoriesOutput{}, fmt.Errorf("load categories: %w", err)
	}
	if len(cats) == 0 {
		return SuggestCategoriesOutput{}, fmt.Errorf("user has no categories")
	}

	validByID := make(map[uint]domain.Category, len(cats))
	catRefs := make([]ai.CategoryRef, 0, len(cats))
	fallbackByType := make(map[string]uint, 2)
	for _, c := range cats {
		if !c.IsActive {
			continue
		}
		validByID[c.ID] = c
		catRefs = append(catRefs, ai.CategoryRef{
			ID:   c.ID,
			Name: c.Name,
			Type: string(c.Type),
		})
		if c.Name == "Outros" {
			fallbackByType[string(c.Type)] = c.ID
		}
	}

	// Few-shot history: most recent N completed transactions in lookback window.
	history := uc.buildHistory(ctx, in.UserID)

	// Map items to AI request.
	itemRefs := make([]ai.TransactionRef, 0, len(in.Items))
	for i, it := range in.Items {
		txType := it.Type
		if txType == "" {
			txType = "expense"
		}
		dateStr := it.Date
		if d, err := timex.ParseFlexibleDate(it.Date); err == nil {
			dateStr = d.Format("2006-01-02")
		}
		itemRefs = append(itemRefs, ai.TransactionRef{
			RowIndex:    i,
			Description: it.Description,
			Amount:      it.Amount,
			Type:        txType,
			Date:        dateStr,
		})
	}

	resp, err := uc.aiClient.Suggest(ctx, ai.SuggestRequest{
		Categories: catRefs,
		History:    history,
		Items:      itemRefs,
	})
	if err != nil {
		uc.logger.Error().Err(err).Msg("ai suggest failed")
		return SuggestCategoriesOutput{}, err
	}

	// Index suggestions by row.
	byRow := make(map[int]ai.CategorySuggestion, len(resp.Suggestions))
	for _, s := range resp.Suggestions {
		byRow[s.RowIndex] = s
	}

	final := make([]ai.CategorySuggestion, len(in.Items))
	for i, it := range in.Items {
		s, ok := byRow[i]
		if !ok {
			s = ai.CategorySuggestion{RowIndex: i}
		} else {
			s.RowIndex = i
			s = uc.validateSuggestion(s, it, validByID)
		}
		if s.CategoryID == 0 {
			type_ := it.Type
			if type_ == "" || type_ == "transfer" {
				type_ = "expense"
			}
			if fb, ok := fallbackByType[type_]; ok {
				s.CategoryID = fb
			}
		}
		final[i] = s
	}

	uc.logger.Info().
		Int("items", len(in.Items)).
		Int("prompt_tokens", resp.Usage.PromptTokens).
		Int("completion_tokens", resp.Usage.CompletionTokens).
		Msg("ai suggest completed")

	return SuggestCategoriesOutput{
		Suggestions: final,
		Usage:       resp.Usage,
	}, nil
}

// validateSuggestion enforces server-side rules: id must belong to user,
// type must match transaction type (try sibling otherwise), confidence
// threshold respected.
func (uc *SuggestCategoriesUseCase) validateSuggestion(
	s ai.CategorySuggestion,
	item ImportItem,
	valid map[uint]domain.Category,
) ai.CategorySuggestion {
	if s.CategoryID == 0 {
		return s
	}
	cat, ok := valid[s.CategoryID]
	if !ok {
		return ai.CategorySuggestion{RowIndex: s.RowIndex}
	}
	itemType := item.Type
	if itemType == "" {
		itemType = "expense"
	}
	// Transfers can use any category; income/expense must match.
	if itemType != "transfer" && string(cat.Type) != itemType {
		return ai.CategorySuggestion{RowIndex: s.RowIndex}
	}
	if s.Confidence < uc.cfg.MinConfidence {
		s.CategoryID = 0
	}
	return s
}

// buildHistory loads recent categorized transactions to use as few-shot examples.
func (uc *SuggestCategoriesUseCase) buildHistory(ctx context.Context, userID uint) []ai.HistoryRef {
	max := uc.cfg.HistoryMaxExamples
	if max <= 0 {
		return nil
	}
	lookbackDays := uc.cfg.HistoryLookbackDays
	if lookbackDays <= 0 {
		lookbackDays = 90
	}
	end := time.Now()
	start := end.AddDate(0, 0, -lookbackDays)

	txs, err := uc.transactionRepo.GetByDateRange(ctx, userID, start, end)
	if err != nil || len(txs) == 0 {
		return nil
	}

	// Most recent first.
	sort.SliceStable(txs, func(i, j int) bool { return txs[i].Date.After(txs[j].Date) })

	out := make([]ai.HistoryRef, 0, max)
	for _, t := range txs {
		if t.CategoryID == 0 || t.Description == "" {
			continue
		}
		out = append(out, ai.HistoryRef{
			Description: t.Description,
			CategoryID:  t.CategoryID,
		})
		if len(out) >= max {
			break
		}
	}
	return out
}

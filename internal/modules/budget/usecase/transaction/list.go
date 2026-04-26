package transaction

import (
	"context"
	"time"

	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/pkg/logger"
	"github.com/edalferes/monetics/pkg/timex"
)

type ListUseCase struct {
	transactionRepo interfaces.TransactionRepository
	logger          logger.Logger
}

func NewListUseCase(transactionRepo interfaces.TransactionRepository, log logger.Logger) *ListUseCase {
	return &ListUseCase{
		transactionRepo: transactionRepo,
		logger:          log.With().Str("usecase", "transaction.list").Logger(),
	}
}

type ListInput struct {
	UserID     uint
	Page       int
	PageSize   int
	Type       *domain.TransactionType
	AccountID  *uint
	CategoryID *uint
	StartDate  *string
	EndDate    *string
}

type ListOutput struct {
	Transactions []domain.Transaction
	Total        int64
	Page         int
	PageSize     int
	TotalPages   int
}

func (uc *ListUseCase) Execute(ctx context.Context, input ListInput) (ListOutput, error) {
	logEvent := uc.logger.Debug().
		Uint("user_id", input.UserID).
		Int("page", input.Page).
		Int("page_size", input.PageSize)
	if input.StartDate != nil {
		logEvent = logEvent.Str("start_date", *input.StartDate)
	}
	if input.EndDate != nil {
		logEvent = logEvent.Str("end_date", *input.EndDate)
	}
	logEvent.Msg("listing transactions")

	// Default pagination values
	if input.Page <= 0 {
		input.Page = 1
	}
	if input.PageSize <= 0 {
		input.PageSize = 20 // Default 20 items per page
	}
	if input.PageSize > 100 {
		input.PageSize = 100 // Max 100 items per page
	}

	// Calculate offset
	offset := (input.Page - 1) * input.PageSize

	// Parse date filters if provided
	var startDate, endDate *time.Time
	if input.StartDate != nil {
		parsed, err := timex.ParseFlexibleDate(*input.StartDate)
		if err != nil {
			uc.logger.Error().Err(err).Str("start_date", *input.StartDate).Msg("failed to parse start_date")
			return ListOutput{}, err
		}
		startDate = &parsed
	}
	if input.EndDate != nil {
		parsed, err := timex.ParseFlexibleDate(*input.EndDate)
		if err != nil {
			uc.logger.Error().Err(err).Str("end_date", *input.EndDate).Msg("failed to parse end_date")
			return ListOutput{}, err
		}
		// Set end date to end of day (23:59:59)
		endOfDay := parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		endDate = &endOfDay
	}

	// Get paginated transactions with all filters
	var transactions []domain.Transaction
	var total int64
	var err error

	// Build query with filters
	transactions, err = uc.transactionRepo.GetByUserIDPaginatedWithAllFilters(
		ctx,
		input.UserID,
		input.PageSize,
		offset,
		input.Type,
		input.AccountID,
		input.CategoryID,
		startDate,
		endDate,
	)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", input.UserID).Msg("failed to get paginated transactions")
		return ListOutput{}, err
	}

	total, err = uc.transactionRepo.CountByUserIDWithAllFilters(
		ctx,
		input.UserID,
		input.Type,
		input.AccountID,
		input.CategoryID,
		startDate,
		endDate,
	)
	if err != nil {
		uc.logger.Error().Err(err).Uint("user_id", input.UserID).Msg("failed to count transactions")
		return ListOutput{}, err
	}

	// Calculate total pages
	totalPages := int(total) / input.PageSize
	if int(total)%input.PageSize > 0 {
		totalPages++
	}

	uc.logger.Info().
		Uint("user_id", input.UserID).
		Int("count", len(transactions)).
		Int("total", int(total)).
		Msg("transactions listed successfully")

	return ListOutput{
		Transactions: transactions,
		Total:        total,
		Page:         input.Page,
		PageSize:     input.PageSize,
		TotalPages:   totalPages,
	}, nil
}

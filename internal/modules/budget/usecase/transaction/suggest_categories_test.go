package transaction_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/edalferes/monetics/internal/config"
	"github.com/edalferes/monetics/internal/modules/budget/adapters/ai"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	budgeterrors "github.com/edalferes/monetics/internal/modules/budget/errors"
	transactionuc "github.com/edalferes/monetics/internal/modules/budget/usecase/transaction"
	"github.com/edalferes/monetics/pkg/logger"
)

// mockAIClient implements ai.Client.
type mockAIClient struct {
	mock.Mock
}

func (m *mockAIClient) Suggest(ctx context.Context, req ai.SuggestRequest) (ai.SuggestResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(ai.SuggestResponse), args.Error(1)
}

func newSuggestUseCase(
	t *testing.T,
	aiClient ai.Client,
	accRepo *MockAccountRepository,
	catRepo *MockCategoryRepository,
	txRepo *MockTransactionRepository,
) *transactionuc.SuggestCategoriesUseCase {
	t.Helper()
	cfg := config.AIConfig{
		Enabled:             true,
		Provider:            "openai",
		Model:               "gpt-4o-mini",
		MaxItemsPerRequest:  500,
		MinConfidence:       0.4,
		HistoryLookbackDays: 90,
		HistoryMaxExamples:  5,
	}
	log := logger.New(logger.Config{Level: "error", Format: "console"})
	return transactionuc.NewSuggestCategoriesUseCase(aiClient, accRepo, catRepo, txRepo, cfg, log)
}

func TestSuggestCategories_Success(t *testing.T) {
	accRepo := new(MockAccountRepository)
	catRepo := new(MockCategoryRepository)
	txRepo := new(MockTransactionRepository)
	aiClient := new(mockAIClient)

	accRepo.On("GetByID", mock.Anything, uint(1)).
		Return(domain.Account{ID: 1, UserID: 10}, nil)
	catRepo.On("GetByUserID", mock.Anything, uint(10)).
		Return([]domain.Category{
			{ID: 100, UserID: 10, Name: "Food", Type: domain.CategoryTypeExpense, IsActive: true},
			{ID: 101, UserID: 10, Name: "Salary", Type: domain.CategoryTypeIncome, IsActive: true},
		}, nil)
	txRepo.On("GetByDateRange", mock.Anything, uint(10), mock.Anything, mock.Anything).
		Return([]domain.Transaction{}, nil)

	aiClient.On("Suggest", mock.Anything, mock.Anything).
		Return(ai.SuggestResponse{
			Suggestions: []ai.CategorySuggestion{
				{RowIndex: 0, CategoryID: 100, Confidence: 0.9},
				{RowIndex: 1, CategoryID: 101, Confidence: 0.85},
			},
			Usage: ai.Usage{PromptTokens: 100, CompletionTokens: 50},
		}, nil)

	uc := newSuggestUseCase(t, aiClient, accRepo, catRepo, txRepo)
	out, err := uc.Execute(context.Background(), transactionuc.SuggestCategoriesInput{
		UserID:    10,
		AccountID: 1,
		Items: []transactionuc.ImportItem{
			{Date: "2026-03-01", Description: "IFOOD", Amount: 47.10, Type: "expense"},
			{Date: "2026-03-02", Description: "Payroll", Amount: 5000, Type: "income"},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, out.Suggestions, 2)
	assert.Equal(t, uint(100), out.Suggestions[0].CategoryID)
	assert.Equal(t, uint(101), out.Suggestions[1].CategoryID)
}

func TestSuggestCategories_InvalidIDBecomesZero(t *testing.T) {
	accRepo := new(MockAccountRepository)
	catRepo := new(MockCategoryRepository)
	txRepo := new(MockTransactionRepository)
	aiClient := new(mockAIClient)

	accRepo.On("GetByID", mock.Anything, uint(1)).
		Return(domain.Account{ID: 1, UserID: 10}, nil)
	catRepo.On("GetByUserID", mock.Anything, uint(10)).
		Return([]domain.Category{
			{ID: 100, UserID: 10, Name: "Food", Type: domain.CategoryTypeExpense, IsActive: true},
		}, nil)
	txRepo.On("GetByDateRange", mock.Anything, uint(10), mock.Anything, mock.Anything).
		Return([]domain.Transaction{}, nil)

	aiClient.On("Suggest", mock.Anything, mock.Anything).
		Return(ai.SuggestResponse{
			Suggestions: []ai.CategorySuggestion{
				{RowIndex: 0, CategoryID: 999, Confidence: 0.95}, // not in user's categories
			},
		}, nil)

	uc := newSuggestUseCase(t, aiClient, accRepo, catRepo, txRepo)
	out, err := uc.Execute(context.Background(), transactionuc.SuggestCategoriesInput{
		UserID:    10,
		AccountID: 1,
		Items: []transactionuc.ImportItem{
			{Date: "2026-03-01", Description: "x", Amount: 10, Type: "expense"},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, uint(0), out.Suggestions[0].CategoryID)
}

func TestSuggestCategories_LowConfidenceBecomesZero(t *testing.T) {
	accRepo := new(MockAccountRepository)
	catRepo := new(MockCategoryRepository)
	txRepo := new(MockTransactionRepository)
	aiClient := new(mockAIClient)

	accRepo.On("GetByID", mock.Anything, uint(1)).
		Return(domain.Account{ID: 1, UserID: 10}, nil)
	catRepo.On("GetByUserID", mock.Anything, uint(10)).
		Return([]domain.Category{
			{ID: 100, UserID: 10, Name: "Food", Type: domain.CategoryTypeExpense, IsActive: true},
		}, nil)
	txRepo.On("GetByDateRange", mock.Anything, uint(10), mock.Anything, mock.Anything).
		Return([]domain.Transaction{}, nil)

	aiClient.On("Suggest", mock.Anything, mock.Anything).
		Return(ai.SuggestResponse{
			Suggestions: []ai.CategorySuggestion{
				{RowIndex: 0, CategoryID: 100, Confidence: 0.2},
			},
		}, nil)

	uc := newSuggestUseCase(t, aiClient, accRepo, catRepo, txRepo)
	out, err := uc.Execute(context.Background(), transactionuc.SuggestCategoriesInput{
		UserID:    10,
		AccountID: 1,
		Items: []transactionuc.ImportItem{
			{Date: "2026-03-01", Description: "x", Amount: 10, Type: "expense"},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, uint(0), out.Suggestions[0].CategoryID)
}

func TestSuggestCategories_TypeMismatchBecomesZero(t *testing.T) {
	accRepo := new(MockAccountRepository)
	catRepo := new(MockCategoryRepository)
	txRepo := new(MockTransactionRepository)
	aiClient := new(mockAIClient)

	accRepo.On("GetByID", mock.Anything, uint(1)).
		Return(domain.Account{ID: 1, UserID: 10}, nil)
	catRepo.On("GetByUserID", mock.Anything, uint(10)).
		Return([]domain.Category{
			{ID: 100, UserID: 10, Name: "Salary", Type: domain.CategoryTypeIncome, IsActive: true},
		}, nil)
	txRepo.On("GetByDateRange", mock.Anything, uint(10), mock.Anything, mock.Anything).
		Return([]domain.Transaction{}, nil)

	aiClient.On("Suggest", mock.Anything, mock.Anything).
		Return(ai.SuggestResponse{
			Suggestions: []ai.CategorySuggestion{
				{RowIndex: 0, CategoryID: 100, Confidence: 0.9}, // income cat for expense tx
			},
		}, nil)

	uc := newSuggestUseCase(t, aiClient, accRepo, catRepo, txRepo)
	out, err := uc.Execute(context.Background(), transactionuc.SuggestCategoriesInput{
		UserID:    10,
		AccountID: 1,
		Items: []transactionuc.ImportItem{
			{Date: "2026-03-01", Description: "x", Amount: 10, Type: "expense"},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, uint(0), out.Suggestions[0].CategoryID)
}

func TestSuggestCategories_MissingRowFilledWithZero(t *testing.T) {
	accRepo := new(MockAccountRepository)
	catRepo := new(MockCategoryRepository)
	txRepo := new(MockTransactionRepository)
	aiClient := new(mockAIClient)

	accRepo.On("GetByID", mock.Anything, uint(1)).
		Return(domain.Account{ID: 1, UserID: 10}, nil)
	catRepo.On("GetByUserID", mock.Anything, uint(10)).
		Return([]domain.Category{
			{ID: 100, UserID: 10, Name: "Food", Type: domain.CategoryTypeExpense, IsActive: true},
		}, nil)
	txRepo.On("GetByDateRange", mock.Anything, uint(10), mock.Anything, mock.Anything).
		Return([]domain.Transaction{}, nil)

	aiClient.On("Suggest", mock.Anything, mock.Anything).
		Return(ai.SuggestResponse{
			Suggestions: []ai.CategorySuggestion{
				{RowIndex: 0, CategoryID: 100, Confidence: 0.9},
				// row 1 missing
			},
		}, nil)

	uc := newSuggestUseCase(t, aiClient, accRepo, catRepo, txRepo)
	out, err := uc.Execute(context.Background(), transactionuc.SuggestCategoriesInput{
		UserID:    10,
		AccountID: 1,
		Items: []transactionuc.ImportItem{
			{Date: "2026-03-01", Description: "a", Amount: 10, Type: "expense"},
			{Date: "2026-03-02", Description: "b", Amount: 20, Type: "expense"},
		},
	})

	assert.NoError(t, err)
	assert.Len(t, out.Suggestions, 2)
	assert.Equal(t, uint(100), out.Suggestions[0].CategoryID)
	assert.Equal(t, uint(0), out.Suggestions[1].CategoryID)
}

func TestSuggestCategories_AccountNotOwnedByUser(t *testing.T) {
	accRepo := new(MockAccountRepository)
	catRepo := new(MockCategoryRepository)
	txRepo := new(MockTransactionRepository)
	aiClient := new(mockAIClient)

	accRepo.On("GetByID", mock.Anything, uint(1)).
		Return(domain.Account{ID: 1, UserID: 99}, nil)

	uc := newSuggestUseCase(t, aiClient, accRepo, catRepo, txRepo)
	_, err := uc.Execute(context.Background(), transactionuc.SuggestCategoriesInput{
		UserID:    10,
		AccountID: 1,
		Items: []transactionuc.ImportItem{
			{Date: "2026-03-01", Description: "a", Amount: 10, Type: "expense"},
		},
	})

	assert.ErrorIs(t, err, budgeterrors.ErrUnauthorizedAccess)
}

func TestSuggestCategories_AIError(t *testing.T) {
	accRepo := new(MockAccountRepository)
	catRepo := new(MockCategoryRepository)
	txRepo := new(MockTransactionRepository)
	aiClient := new(mockAIClient)

	accRepo.On("GetByID", mock.Anything, uint(1)).
		Return(domain.Account{ID: 1, UserID: 10}, nil)
	catRepo.On("GetByUserID", mock.Anything, uint(10)).
		Return([]domain.Category{
			{ID: 100, UserID: 10, Name: "Food", Type: domain.CategoryTypeExpense, IsActive: true},
		}, nil)
	txRepo.On("GetByDateRange", mock.Anything, uint(10), mock.Anything, mock.Anything).
		Return([]domain.Transaction{}, nil)

	aiClient.On("Suggest", mock.Anything, mock.Anything).
		Return(ai.SuggestResponse{}, errors.New("upstream timeout"))

	uc := newSuggestUseCase(t, aiClient, accRepo, catRepo, txRepo)
	_, err := uc.Execute(context.Background(), transactionuc.SuggestCategoriesInput{
		UserID:    10,
		AccountID: 1,
		Items: []transactionuc.ImportItem{
			{Date: "2026-03-01", Description: "x", Amount: 10, Type: "expense"},
		},
	})

	assert.Error(t, err)
}

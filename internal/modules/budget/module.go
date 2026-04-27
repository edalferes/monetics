package budget

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/edalferes/monetics/internal/config"
	"github.com/edalferes/monetics/internal/modules/auth"
	"github.com/edalferes/monetics/internal/modules/budget/adapters/ai"
	"github.com/edalferes/monetics/internal/modules/budget/adapters/http/handlers"
	"github.com/edalferes/monetics/internal/modules/budget/adapters/repository"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/account"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/budget"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/category"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/interfaces"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/report"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/transaction"
	"github.com/edalferes/monetics/pkg/logger"
)

// Module represents the budget module
type Module struct {
	db  *gorm.DB
	cfg *config.Config

	// Repositories
	accountRepo     interfaces.AccountRepository
	categoryRepo    interfaces.CategoryRepository
	transactionRepo interfaces.TransactionRepository
	budgetRepo      interfaces.BudgetRepository

	// Use cases - Account
	createAccountUseCase  *account.CreateUseCase
	listAccountsUseCase   *account.ListUseCase
	getAccountByIDUseCase *account.GetByIDUseCase
	updateAccountUseCase  *account.UpdateUseCase
	deleteAccountUseCase  *account.DeleteUseCase

	// Use cases - Category
	createCategoryUseCase  *category.CreateUseCase
	listCategoriesUseCase  *category.ListUseCase
	getCategoryByIDUseCase *category.GetByIDUseCase
	updateCategoryUseCase  *category.UpdateUseCase
	deleteCategoryUseCase  *category.DeleteUseCase

	// Use cases - Transaction
	createTransactionUseCase  *transaction.CreateUseCase
	listTransactionsUseCase   *transaction.ListUseCase
	getTransactionByIDUseCase *transaction.GetByIDUseCase
	updateTransactionUseCase  *transaction.UpdateUseCase
	deleteTransactionUseCase  *transaction.DeleteUseCase
	importCSVUseCase          *transaction.ImportCSVUseCase
	suggestCategoriesUseCase  *transaction.SuggestCategoriesUseCase

	// Use cases - Budget
	createBudgetUseCase  *budget.CreateUseCase
	listBudgetsUseCase   *budget.ListUseCase
	getBudgetByIDUseCase *budget.GetByIDUseCase
	updateBudgetUseCase  *budget.UpdateUseCase
	deleteBudgetUseCase  *budget.DeleteUseCase

	// Use cases - Report
	getAccountBalanceUseCase *report.GetAccountBalanceUseCase
	getMonthlyReportUseCase  *report.GetMonthlyReportUseCase

	// Handlers
	accountHandler     *handlers.AccountHandler
	categoryHandler    *handlers.CategoryHandler
	transactionHandler *handlers.TransactionHandler
	budgetHandler      *handlers.BudgetHandler
	reportHandler      *handlers.ReportHandler
	featuresHandler    *handlers.FeaturesHandler
}

// NewModule creates a new budget module instance using the standard
// (db, cfg, log) constructor signature.
func NewModule(db *gorm.DB, cfg *config.Config, log logger.Logger) *Module {
	module := &Module{
		db:  db,
		cfg: cfg,
	}

	// Initialize repositories
	module.accountRepo = repository.NewAccountRepository(db)
	module.categoryRepo = repository.NewCategoryRepository(db)
	module.transactionRepo = repository.NewTransactionRepository(db)
	module.budgetRepo = repository.NewBudgetRepository(db)

	// Initialize use cases - Account
	module.createAccountUseCase = account.NewCreateUseCase(module.accountRepo, log)
	module.listAccountsUseCase = account.NewListUseCase(module.accountRepo, module.transactionRepo, log)
	module.getAccountByIDUseCase = account.NewGetByIDUseCase(module.accountRepo, log)
	module.updateAccountUseCase = account.NewUpdateUseCase(module.accountRepo, log)
	module.deleteAccountUseCase = account.NewDeleteUseCase(module.accountRepo, log)

	// Initialize use cases - Category
	module.createCategoryUseCase = category.NewCreateUseCase(module.categoryRepo, log)
	module.listCategoriesUseCase = category.NewListUseCase(module.categoryRepo, log)
	module.getCategoryByIDUseCase = category.NewGetByIDUseCase(module.categoryRepo, log)
	module.updateCategoryUseCase = category.NewUpdateUseCase(module.categoryRepo, log)
	module.deleteCategoryUseCase = category.NewDeleteUseCase(module.categoryRepo, module.transactionRepo, module.budgetRepo, log)

	// Initialize use cases - Transaction
	module.createTransactionUseCase = transaction.NewCreateUseCase(
		module.transactionRepo,
		module.accountRepo,
		module.categoryRepo,
		module.budgetRepo,
		log,
	)
	module.listTransactionsUseCase = transaction.NewListUseCase(module.transactionRepo, log)
	module.getTransactionByIDUseCase = transaction.NewGetByIDUseCase(module.transactionRepo, log)
	module.updateTransactionUseCase = transaction.NewUpdateUseCase(
		module.transactionRepo,
		module.accountRepo,
		module.categoryRepo,
		module.budgetRepo,
		log,
	)
	module.deleteTransactionUseCase = transaction.NewDeleteUseCase(module.transactionRepo, module.budgetRepo, log)
	module.importCSVUseCase = transaction.NewImportCSVUseCase(
		module.transactionRepo,
		module.accountRepo,
		module.categoryRepo,
		module.budgetRepo,
		log,
	)

	// AI-assisted suggestion (optional, gated by cfg.AI.Enabled)
	if cfg.AI.Enabled {
		var aiClient ai.Client = ai.NewOpenAIClient(cfg.AI, log)
		module.suggestCategoriesUseCase = transaction.NewSuggestCategoriesUseCase(
			aiClient,
			module.accountRepo,
			module.categoryRepo,
			module.transactionRepo,
			cfg.AI,
			log,
		)
	}

	// Initialize use cases - Budget
	module.createBudgetUseCase = budget.NewCreateUseCase(
		module.budgetRepo,
		module.categoryRepo,
		log,
	)
	module.listBudgetsUseCase = budget.NewListUseCase(module.budgetRepo, module.transactionRepo, log)
	module.getBudgetByIDUseCase = budget.NewGetByIDUseCase(module.budgetRepo, log)
	module.updateBudgetUseCase = budget.NewUpdateUseCase(module.budgetRepo, log)
	module.deleteBudgetUseCase = budget.NewDeleteUseCase(module.budgetRepo, log)

	// Initialize use cases - Report
	module.getAccountBalanceUseCase = report.NewGetAccountBalanceUseCase(
		module.accountRepo,
		module.transactionRepo,
	)
	module.getMonthlyReportUseCase = report.NewGetMonthlyReportUseCase(
		module.transactionRepo,
		module.budgetRepo,
		module.categoryRepo,
	)

	// Initialize handlers
	module.accountHandler = handlers.NewAccountHandler(
		module.createAccountUseCase,
		module.listAccountsUseCase,
		module.getAccountByIDUseCase,
		module.updateAccountUseCase,
		module.deleteAccountUseCase,
		module.getAccountBalanceUseCase,
	)
	module.categoryHandler = handlers.NewCategoryHandler(
		module.createCategoryUseCase,
		module.listCategoriesUseCase,
		module.getCategoryByIDUseCase,
		module.updateCategoryUseCase,
		module.deleteCategoryUseCase,
	)
	module.transactionHandler = handlers.NewTransactionHandler(
		module.createTransactionUseCase,
		module.listTransactionsUseCase,
		module.getTransactionByIDUseCase,
		module.updateTransactionUseCase,
		module.deleteTransactionUseCase,
		module.importCSVUseCase,
		module.suggestCategoriesUseCase,
	)
	module.budgetHandler = handlers.NewBudgetHandler(
		module.createBudgetUseCase,
		module.listBudgetsUseCase,
		module.getBudgetByIDUseCase,
		module.updateBudgetUseCase,
		module.deleteBudgetUseCase,
	)
	module.reportHandler = handlers.NewReportHandler(
		module.getMonthlyReportUseCase,
	)
	module.featuresHandler = handlers.NewFeaturesHandler(cfg.AI.Enabled)

	return module
}

// RegisterRoutes registers all budget module routes
func (m *Module) RegisterRoutes(api *echo.Group, authMiddleware echo.MiddlewareFunc) {
	budget := api.Group("/budget")
	budget.Use(authMiddleware)

	// Account routes
	accounts := budget.Group("/accounts")
	accounts.POST("", m.accountHandler.CreateAccount)
	accounts.GET("", m.accountHandler.ListAccounts)
	accounts.GET("/:id", m.accountHandler.GetAccount)
	accounts.GET("/:id/detail", m.accountHandler.GetAccountByID)
	accounts.PUT("/:id", m.accountHandler.UpdateAccount)
	accounts.DELETE("/:id", m.accountHandler.DeleteAccount)

	// Category routes
	categories := budget.Group("/categories")
	categories.POST("", m.categoryHandler.CreateCategory)
	categories.GET("", m.categoryHandler.ListCategories)
	categories.GET("/:id", m.categoryHandler.GetCategoryByID)
	categories.PUT("/:id", m.categoryHandler.UpdateCategory)
	categories.DELETE("/:id", m.categoryHandler.DeleteCategory)

	// Transaction routes
	transactions := budget.Group("/transactions")
	transactions.POST("", m.transactionHandler.CreateTransaction)
	transactions.POST("/import", m.transactionHandler.ImportCSV)
	transactions.POST("/import/ai-suggest", m.transactionHandler.SuggestCategories)
	transactions.GET("", m.transactionHandler.ListTransactions)
	transactions.GET("/:id", m.transactionHandler.GetTransactionByID)
	transactions.PUT("/:id", m.transactionHandler.UpdateTransaction)
	transactions.DELETE("/:id", m.transactionHandler.DeleteTransaction)

	// Budget routes
	budgets := budget.Group("/budgets")
	budgets.POST("", m.budgetHandler.CreateBudget)
	budgets.GET("", m.budgetHandler.ListBudgets)
	budgets.GET("/:id", m.budgetHandler.GetBudgetByID)
	budgets.PUT("/:id", m.budgetHandler.UpdateBudget)
	budgets.DELETE("/:id", m.budgetHandler.DeleteBudget)

	// Report routes
	reports := budget.Group("/reports")
	reports.GET("/monthly", m.reportHandler.GetMonthlyReport)

	// Config / feature flags (auth-protected so only logged users see it)
	configGroup := api.Group("/config")
	configGroup.Use(authMiddleware)
	configGroup.GET("/features", m.featuresHandler.GetFeatures)
}

// WireUp initializes and registers the budget module.
func WireUp(group *echo.Group, db *gorm.DB, cfg *config.Config, log logger.Logger) {
	module := NewModule(db, cfg, log)
	authMiddleware := auth.JWTMiddleware(cfg.JWT.Secret)
	module.RegisterRoutes(group, authMiddleware)

	log.Info().Msg("Budget module initialized")
}

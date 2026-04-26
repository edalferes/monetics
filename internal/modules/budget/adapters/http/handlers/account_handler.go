package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/http/dto"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/account"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/report"
)

// AccountHandler handles HTTP requests for accounts
type AccountHandler struct {
	createAccountUseCase     *account.CreateUseCase
	listAccountsUseCase      *account.ListUseCase
	getAccountByIDUseCase    *account.GetByIDUseCase
	updateAccountUseCase     *account.UpdateUseCase
	deleteAccountUseCase     *account.DeleteUseCase
	getAccountBalanceUseCase *report.GetAccountBalanceUseCase
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(
	createAccountUseCase *account.CreateUseCase,
	listAccountsUseCase *account.ListUseCase,
	getAccountByIDUseCase *account.GetByIDUseCase,
	updateAccountUseCase *account.UpdateUseCase,
	deleteAccountUseCase *account.DeleteUseCase,
	getAccountBalanceUseCase *report.GetAccountBalanceUseCase,
) *AccountHandler {
	return &AccountHandler{
		createAccountUseCase:     createAccountUseCase,
		listAccountsUseCase:      listAccountsUseCase,
		getAccountByIDUseCase:    getAccountByIDUseCase,
		updateAccountUseCase:     updateAccountUseCase,
		deleteAccountUseCase:     deleteAccountUseCase,
		getAccountBalanceUseCase: getAccountBalanceUseCase,
	}
}

// CreateAccount handles account creation
// @Summary Create a new account
// @Tags Budget - Accounts
// @Accept json
// @Produce json
// @Param request body dto.CreateAccountRequest true "Account creation request"
// @Success 201 {object} dto.AccountResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /accounts [post]
func (h *AccountHandler) CreateAccount(c echo.Context) error {
	var req dto.CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Get user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	input := account.CreateInput{
		UserID:      userID,
		Name:        req.Name,
		Type:        domain.AccountType(req.Type),
		Currency:    req.Currency,
		Description: req.Description,
	}

	acc, err := h.createAccountUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, dto.ToAccountResponse(acc))
}

// ListAccounts handles listing user accounts
// @Summary List user accounts
// @Tags Budget - Accounts
// @Produce json
// @Success 200 {array} dto.AccountResponse
// @Failure 500 {object} map[string]interface{}
// @Router /accounts [get]
func (h *AccountHandler) ListAccounts(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	accounts, err := h.listAccountsUseCase.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToAccountResponseList(accounts))
}

// GetAccount handles getting a single account
// @Summary Get account by ID with calculated balance
// @Tags Budget - Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} dto.AccountBalanceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /accounts/{id} [get]
func (h *AccountHandler) GetAccount(c echo.Context) error {
	accountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid account ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	result, err := h.getAccountBalanceUseCase.Execute(c.Request().Context(), userID, uint(accountID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
	}

	response := dto.AccountBalanceResponse{
		Account:        dto.ToAccountResponse(result.Account),
		CurrentBalance: result.CurrentBalance,
		TotalIncome:    result.TotalIncome,
		TotalExpense:   result.TotalExpense,
		TotalTransfers: result.TotalTransfers,
	}

	return c.JSON(http.StatusOK, response)
}

// GetAccountByID handles getting account by ID
// @Summary Get account by ID
// @Tags Budget - Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /accounts/{id}/detail [get]
func (h *AccountHandler) GetAccountByID(c echo.Context) error {
	accountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid account ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	acc, err := h.getAccountByIDUseCase.Execute(c.Request().Context(), userID, uint(accountID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToAccountResponse(acc))
}

// UpdateAccount handles account update
// @Summary Update an account
// @Tags Budget - Accounts
// @Accept json
// @Produce json
// @Param id path int true "Account ID"
// @Param request body dto.UpdateAccountRequest true "Account update request"
// @Success 200 {object} dto.AccountResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /accounts/{id} [put]
func (h *AccountHandler) UpdateAccount(c echo.Context) error {
	accountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid account ID",
		})
	}

	var req dto.UpdateAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	input := account.UpdateInput{
		ID:          uint(accountID),
		UserID:      userID,
		Name:        req.Name,
		Type:        dto.AccountTypePtr(req.Type),
		Currency:    req.Currency,
		Description: req.Description,
	}

	acc, err := h.updateAccountUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToAccountResponse(acc))
}

// DeleteAccount handles account deletion (soft delete)
// @Summary Delete an account
// @Tags Budget - Accounts
// @Produce json
// @Param id path int true "Account ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /accounts/{id} [delete]
func (h *AccountHandler) DeleteAccount(c echo.Context) error {
	accountID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid account ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	err = h.deleteAccountUseCase.Execute(c.Request().Context(), userID, uint(accountID))
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/http/dto"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/transaction"
)

// convertTransactionType converts string pointer to TransactionType pointer
func convertTransactionType(t *string) *domain.TransactionType {
	if t == nil {
		return nil
	}
	txType := domain.TransactionType(*t)
	return &txType
}

const maxImportBatchSize = 1000 // Max transactions per import

// TransactionHandler handles HTTP requests for transactions
type TransactionHandler struct {
	createTransactionUseCase  *transaction.CreateUseCase
	listTransactionsUseCase   *transaction.ListUseCase
	getTransactionByIDUseCase *transaction.GetByIDUseCase
	updateTransactionUseCase  *transaction.UpdateUseCase
	deleteTransactionUseCase  *transaction.DeleteUseCase
	importCSVUseCase          *transaction.ImportCSVUseCase
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(
	createTransactionUseCase *transaction.CreateUseCase,
	listTransactionsUseCase *transaction.ListUseCase,
	getTransactionByIDUseCase *transaction.GetByIDUseCase,
	updateTransactionUseCase *transaction.UpdateUseCase,
	deleteTransactionUseCase *transaction.DeleteUseCase,
	importCSVUseCase *transaction.ImportCSVUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		createTransactionUseCase:  createTransactionUseCase,
		listTransactionsUseCase:   listTransactionsUseCase,
		getTransactionByIDUseCase: getTransactionByIDUseCase,
		updateTransactionUseCase:  updateTransactionUseCase,
		deleteTransactionUseCase:  deleteTransactionUseCase,
		importCSVUseCase:          importCSVUseCase,
	}
}

// CreateTransaction handles transaction creation
// @Summary Create a new transaction
// @Tags Budget - Transactions
// @Accept json
// @Produce json
// @Param request body dto.CreateTransactionRequest true "Transaction creation request"
// @Success 201 {object} dto.TransactionResponse
// @Failure 400 {object} map[string]interface{}
// @Router /transactions [post]
func (h *TransactionHandler) CreateTransaction(c echo.Context) error {
	var req dto.CreateTransactionRequest
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

	input := transaction.CreateInput{
		UserID:               userID,
		AccountID:            req.AccountID,
		CategoryID:           req.CategoryID,
		Type:                 domain.TransactionType(req.Type),
		Amount:               req.Amount,
		Description:          req.Description,
		Date:                 req.Date.Format(time.RFC3339),
		DestinationAccountID: req.DestinationAccountID,
	}

	tx, err := h.createTransactionUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, dto.ToTransactionResponse(tx))
}

// ListTransactions handles listing user transactions with pagination
// @Summary List user transactions
// @Tags Budget - Transactions
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Items per page (default: 20, max: 100)"
// @Param type query string false "Transaction type filter (income, expense, transfer)"
// @Param account_id query int false "Account ID filter"
// @Param category_id query int false "Category ID filter"
// @Param start_date query string false "Start date filter (RFC3339 or YYYY-MM-DD format)"
// @Param end_date query string false "End date filter (RFC3339 or YYYY-MM-DD format)"
// @Success 200 {object} dto.TransactionListResponse
// @Router /transactions [get]
func (h *TransactionHandler) ListTransactions(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if p := c.QueryParam("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if ps := c.QueryParam("page_size"); ps != "" {
		if parsedSize, err := strconv.Atoi(ps); err == nil && parsedSize > 0 {
			pageSize = parsedSize
		}
	}

	// Parse filter parameters
	var transactionType *domain.TransactionType
	if t := c.QueryParam("type"); t != "" {
		txType := domain.TransactionType(t)
		transactionType = &txType
	}

	var accountID, categoryID *uint
	if aid := c.QueryParam("account_id"); aid != "" {
		if parsedID, err := strconv.ParseUint(aid, 10, 32); err == nil {
			id := uint(parsedID)
			accountID = &id
		}
	}
	if cid := c.QueryParam("category_id"); cid != "" {
		if parsedID, err := strconv.ParseUint(cid, 10, 32); err == nil {
			id := uint(parsedID)
			categoryID = &id
		}
	}

	// Parse date filters
	var startDate, endDate *string
	if sd := c.QueryParam("start_date"); sd != "" {
		startDate = &sd
	}
	if ed := c.QueryParam("end_date"); ed != "" {
		endDate = &ed
	}

	input := transaction.ListInput{
		UserID:     userID,
		Page:       page,
		PageSize:   pageSize,
		Type:       transactionType,
		AccountID:  accountID,
		CategoryID: categoryID,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	result, err := h.listTransactionsUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Convert to DTO format
	dtoResult := dto.TransactionListOutput{
		Transactions: result.Transactions,
		Total:        result.Total,
		Page:         result.Page,
		PageSize:     result.PageSize,
		TotalPages:   result.TotalPages,
	}

	return c.JSON(http.StatusOK, dto.ToTransactionListResponse(dtoResult))
}

// GetTransactionByID handles getting transaction by ID
// @Summary Get transaction by ID
// @Tags Budget - Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} dto.TransactionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetTransactionByID(c echo.Context) error {
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid transaction ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	tx, err := h.getTransactionByIDUseCase.Execute(c.Request().Context(), userID, uint(transactionID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToTransactionResponse(tx))
}

// UpdateTransaction handles transaction update
// @Summary Update a transaction
// @Tags Budget - Transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param request body dto.UpdateTransactionRequest true "Transaction update request"
// @Success 200 {object} dto.TransactionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /transactions/{id} [put]
func (h *TransactionHandler) UpdateTransaction(c echo.Context) error {
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid transaction ID",
		})
	}

	var req dto.UpdateTransactionRequest
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

	var dateStr *string
	if req.Date != nil {
		ds := req.Date.Format(time.RFC3339)
		dateStr = &ds
	}

	input := transaction.UpdateInput{
		ID:          uint(transactionID),
		UserID:      userID,
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Type:        convertTransactionType(req.Type),
		Amount:      req.Amount,
		Description: req.Description,
		Date:        dateStr,
	}

	tx, err := h.updateTransactionUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToTransactionResponse(tx))
}

// DeleteTransaction handles transaction deletion (soft delete)
// @Summary Delete a transaction
// @Tags Budget - Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(c echo.Context) error {
	transactionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid transaction ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	err = h.deleteTransactionUseCase.Execute(c.Request().Context(), userID, uint(transactionID))
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// ImportCSV handles importing transactions from a CSV file
// @Summary Import transactions from CSV
// @Tags Budget - Transactions
// @Accept json
// @Produce json
// @Param request body dto.ImportTransactionsRequest true "Bulk import request"
// @Success 200 {object} dto.ImportTransactionsResponse
// @Failure 400 {object} map[string]interface{}
// @Router /transactions/import [post]
func (h *TransactionHandler) ImportCSV(c echo.Context) error {
	var req dto.ImportTransactionsRequest
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

	items := make([]transaction.ImportItem, len(req.Transactions))
	for i, t := range req.Transactions {
		items[i] = transaction.ImportItem{
			Date:        t.Date,
			Description: t.Description,
			Amount:      t.Amount,
			CategoryID:  t.CategoryID,
		}
	}

	input := transaction.ImportInput{
		UserID:       userID,
		AccountID:    req.AccountID,
		Transactions: items,
	}

	result, err := h.importCSVUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToImportTransactionsResponse(result))
}

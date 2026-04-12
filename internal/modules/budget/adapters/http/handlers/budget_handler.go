package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/http/dto"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	budgetUseCase "github.com/edalferes/monetics/internal/modules/budget/usecase/budget"
)

// convertBudgetPeriod converts string pointer to BudgetPeriod pointer
func convertBudgetPeriod(p *string) *domain.BudgetPeriod {
	if p == nil {
		return nil
	}
	period := domain.BudgetPeriod(*p)
	return &period
}

// BudgetHandler handles HTTP requests for budgets
type BudgetHandler struct {
	createBudgetUseCase  *budgetUseCase.CreateUseCase
	listBudgetsUseCase   *budgetUseCase.ListUseCase
	getBudgetByIDUseCase *budgetUseCase.GetByIDUseCase
	updateBudgetUseCase  *budgetUseCase.UpdateUseCase
	deleteBudgetUseCase  *budgetUseCase.DeleteUseCase
}

// NewBudgetHandler creates a new budget handler
func NewBudgetHandler(
	createBudgetUseCase *budgetUseCase.CreateUseCase,
	listBudgetsUseCase *budgetUseCase.ListUseCase,
	getBudgetByIDUseCase *budgetUseCase.GetByIDUseCase,
	updateBudgetUseCase *budgetUseCase.UpdateUseCase,
	deleteBudgetUseCase *budgetUseCase.DeleteUseCase,
) *BudgetHandler {
	return &BudgetHandler{
		createBudgetUseCase:  createBudgetUseCase,
		listBudgetsUseCase:   listBudgetsUseCase,
		getBudgetByIDUseCase: getBudgetByIDUseCase,
		updateBudgetUseCase:  updateBudgetUseCase,
		deleteBudgetUseCase:  deleteBudgetUseCase,
	}
}

// CreateBudget handles budget creation
// @Summary Create a new budget
// @Tags Budget - Budgets
// @Accept json
// @Produce json
// @Param request body dto.CreateBudgetRequest true "Budget creation request"
// @Success 201 {object} dto.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Router /budgets [post]
func (h *BudgetHandler) CreateBudget(c echo.Context) error {
	var req dto.CreateBudgetRequest
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

	input := budgetUseCase.CreateInput{
		UserID:      userID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Amount:      req.Amount,
		Period:      domain.BudgetPeriod(req.Period),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		AlertAt:     req.AlertAt,
		Description: req.Description,
	}

	budgetResult, err := h.createBudgetUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, dto.ToBudgetResponse(budgetResult))
}

// ListBudgets handles listing user budgets
// @Summary List user budgets
// @Tags Budget - Budgets
// @Produce json
// @Param active_only query bool false "Filter active budgets only"
// @Success 200 {array} dto.BudgetResponse
// @Router /budgets [get]
func (h *BudgetHandler) ListBudgets(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// TODO: Implement active_only filter if needed
	budgets, err := h.listBudgetsUseCase.Execute(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToBudgetResponseList(budgets))
}

// GetBudgetByID handles getting budget by ID
// @Summary Get budget by ID
// @Tags Budget - Budgets
// @Produce json
// @Param id path int true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /budgets/{id} [get]
func (h *BudgetHandler) GetBudgetByID(c echo.Context) error {
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid budget ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	budget, err := h.getBudgetByIDUseCase.Execute(c.Request().Context(), userID, uint(budgetID))
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToBudgetResponse(budget))
}

// UpdateBudget handles budget update
// @Summary Update a budget
// @Tags Budget - Budgets
// @Accept json
// @Produce json
// @Param id path int true "Budget ID"
// @Param request body dto.UpdateBudgetRequest true "Budget update request"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /budgets/{id} [put]
func (h *BudgetHandler) UpdateBudget(c echo.Context) error {
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid budget ID",
		})
	}

	var req dto.UpdateBudgetRequest
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

	input := budgetUseCase.UpdateInput{
		ID:          uint(budgetID),
		UserID:      userID,
		Name:        req.Name,
		Amount:      req.Amount,
		Period:      convertBudgetPeriod(req.Period),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		AlertAt:     req.AlertAt,
		Description: req.Description,
	}

	budget, err := h.updateBudgetUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToBudgetResponse(budget))
}

// DeleteBudget handles budget deletion (soft delete)
// @Summary Delete a budget
// @Tags Budget - Budgets
// @Produce json
// @Param id path int true "Budget ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /budgets/{id} [delete]
func (h *BudgetHandler) DeleteBudget(c echo.Context) error {
	budgetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid budget ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	err = h.deleteBudgetUseCase.Execute(c.Request().Context(), userID, uint(budgetID))
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

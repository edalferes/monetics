package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/http/dto"
	"github.com/edalferes/monetics/internal/modules/budget/domain"
	"github.com/edalferes/monetics/internal/modules/budget/usecase/category"
)

// convertCategoryType converts string pointer to CategoryType pointer
func convertCategoryType(t *string) *domain.CategoryType {
	if t == nil {
		return nil
	}
	categoryType := domain.CategoryType(*t)
	return &categoryType
}

// CategoryHandler handles HTTP requests for categories
type CategoryHandler struct {
	createCategoryUseCase  *category.CreateUseCase
	listCategoriesUseCase  *category.ListUseCase
	getCategoryByIDUseCase *category.GetByIDUseCase
	updateCategoryUseCase  *category.UpdateUseCase
	deleteCategoryUseCase  *category.DeleteUseCase
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(
	createCategoryUseCase *category.CreateUseCase,
	listCategoriesUseCase *category.ListUseCase,
	getCategoryByIDUseCase *category.GetByIDUseCase,
	updateCategoryUseCase *category.UpdateUseCase,
	deleteCategoryUseCase *category.DeleteUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		createCategoryUseCase:  createCategoryUseCase,
		listCategoriesUseCase:  listCategoriesUseCase,
		getCategoryByIDUseCase: getCategoryByIDUseCase,
		updateCategoryUseCase:  updateCategoryUseCase,
		deleteCategoryUseCase:  deleteCategoryUseCase,
	}
}

// CreateCategory handles category creation
// @Summary Create a new category
// @Tags Budget - Categories
// @Accept json
// @Produce json
// @Param request body dto.CreateCategoryRequest true "Category creation request"
// @Success 201 {object} dto.CategoryResponse
// @Failure 400 {object} map[string]interface{}
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	var req dto.CreateCategoryRequest
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

	input := category.CreateInput{
		UserID:      userID,
		Name:        req.Name,
		Type:        domain.CategoryType(req.Type),
		Icon:        req.Icon,
		Color:       req.Color,
		Description: req.Description,
	}

	categoryResult, err := h.createCategoryUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, dto.ToCategoryResponse(categoryResult))
}

// ListCategories handles listing user categories
// @Summary List user categories
// @Tags Budget - Categories
// @Produce json
// @Param type query string false "Filter by type (income or expense)"
// @Success 200 {array} dto.CategoryResponse
// @Router /categories [get]
func (h *CategoryHandler) ListCategories(c echo.Context) error {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Parse optional type filter
	var categoryType *domain.CategoryType
	if typeFilter := c.QueryParam("type"); typeFilter != "" {
		ct := domain.CategoryType(typeFilter)
		categoryType = &ct
	}

	categories, err := h.listCategoriesUseCase.Execute(c.Request().Context(), userID, categoryType)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToCategoryResponseList(categories))
}

// GetCategoryByID handles getting category by ID
// @Summary Get category by ID
// @Tags Budget - Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c echo.Context) error {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid category ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	cat, err := h.getCategoryByIDUseCase.Execute(c.Request().Context(), userID, uint(categoryID))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToCategoryResponse(cat))
}

// UpdateCategory handles category update
// @Summary Update a category
// @Tags Budget - Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body dto.UpdateCategoryRequest true "Category update request"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid category ID",
		})
	}

	var req dto.UpdateCategoryRequest
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

	input := category.UpdateInput{
		ID:          uint(categoryID),
		UserID:      userID,
		Name:        req.Name,
		Type:        convertCategoryType(req.Type),
		Icon:        req.Icon,
		Color:       req.Color,
		Description: req.Description,
	}

	cat, err := h.updateCategoryUseCase.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.ToCategoryResponse(cat))
}

// DeleteCategory handles category deletion (soft delete)
// @Summary Delete a category
// @Tags Budget - Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid category ID",
		})
	}

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
	}

	err = h.deleteCategoryUseCase.Execute(c.Request().Context(), uint(categoryID), userID)
	if err != nil {
		return c.JSON(errorToHTTPStatus(err), map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

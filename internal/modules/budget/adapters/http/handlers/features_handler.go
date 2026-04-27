package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/edalferes/monetics/internal/modules/budget/adapters/http/dto"
)

// FeaturesHandler exposes runtime feature flags to the UI.
type FeaturesHandler struct {
	aiImportEnabled bool
}

// NewFeaturesHandler creates a new FeaturesHandler.
func NewFeaturesHandler(aiImportEnabled bool) *FeaturesHandler {
	return &FeaturesHandler{aiImportEnabled: aiImportEnabled}
}

// GetFeatures returns the enabled feature flags.
// @Summary Get enabled feature flags
// @Tags Config
// @Produce json
// @Success 200 {object} dto.FeaturesResponse
// @Router /config/features [get]
func (h *FeaturesHandler) GetFeatures(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.FeaturesResponse{
		AIImport: h.aiImportEnabled,
	})
}

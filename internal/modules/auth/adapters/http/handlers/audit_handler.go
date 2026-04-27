package handlers

import (
	"net/http"

	"github.com/edalferes/monetics/internal/modules/auth/adapters/http/dto"
	"github.com/edalferes/monetics/internal/modules/auth/usecase/interfaces"
	pkgresponses "github.com/edalferes/monetics/pkg/responses"
	"github.com/labstack/echo/v4"
)

type AuditHandler struct {
	auditLogRepo interfaces.AuditLogRepository
}

func NewAuditHandler(auditLogRepo interfaces.AuditLogRepository) *AuditHandler {
	return &AuditHandler{
		auditLogRepo: auditLogRepo,
	}
}

// ListAuditLogs godoc
// @Summary List all audit logs
// @Description Get a list of all audit logs
// @Tags Audit
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.AuditLogResponse
// @Failure 500 {object} pkgresponses.ErrorResponse
// @Router /auth/audit-logs [get]
func (h *AuditHandler) ListAuditLogs(c echo.Context) error {
	logs, err := h.auditLogRepo.ListAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, pkgresponses.ErrorResponse{
			Error: "Failed to retrieve audit logs",
		})
	}

	return pkgresponses.OK(c, dto.ToAuditLogResponseList(logs))
}

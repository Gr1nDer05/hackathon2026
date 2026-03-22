package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

func (h *Handler) CreatePsychologistByAdmin(c *gin.Context) {
	var input domain.CreatePsychologistInput
	if !bindJSON(c, &input) {
		return
	}

	if fieldErrors := validateCreatePsychologistInput(input); fieldErrors != nil {
		writeValidationError(c, fieldErrors)
		return
	}

	workspace, err := h.appService.CreatePsychologistByAdmin(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			writeError(c, http.StatusConflict, "Validation failed", map[string]string{
				"email": "Email already exists",
			})
		default:
			writeError(c, http.StatusInternalServerError, "Failed to create psychologist", nil)
		}
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func (h *Handler) ListPsychologists(c *gin.Context) {
	users, err := h.appService.ListPsychologists(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list psychologists", nil)
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetPsychologistWorkspaceByAdmin(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	workspace, err := h.appService.GetPsychologistWorkspace(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Psychologist not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to load psychologist workspace", nil)
		}
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *Handler) UpdatePsychologistAccountByAdmin(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.UpdatePsychologistAccountInput
	if !bindJSON(c, &input) {
		return
	}

	if fieldErrors := validatePsychologistAccountInput(input); fieldErrors != nil {
		writeValidationError(c, fieldErrors)
		return
	}

	user, err := h.appService.UpdatePsychologistAccount(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			writeError(c, http.StatusConflict, "Validation failed", map[string]string{
				"email": "Email already exists",
			})
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Psychologist not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update psychologist account", nil)
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) UpdatePsychologistProfileByAdmin(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.UpdatePsychologistProfileInput
	if !bindJSON(c, &input) {
		return
	}

	if fieldErrors := validatePsychologistProfileInput(input, true); fieldErrors != nil {
		writeValidationError(c, fieldErrors)
		return
	}

	profile, err := h.appService.UpdatePsychologistProfile(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Psychologist not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update psychologist profile", nil)
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdatePsychologistCardByAdmin(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.UpdatePsychologistCardInput
	if !bindJSON(c, &input) {
		return
	}

	if fieldErrors := validatePsychologistCardInput(input, true); fieldErrors != nil {
		writeValidationError(c, fieldErrors)
		return
	}

	card, err := h.appService.UpdatePsychologistCard(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Psychologist not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update psychologist card", nil)
		}
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *Handler) UpdatePsychologistAccessByAdmin(c *gin.Context) {
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.UpdatePsychologistAccessInput
	if !bindJSON(c, &input) {
		return
	}

	user, err := h.appService.UpdatePsychologistAccess(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPsychologistAccess):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"portal_access_until": "Use RFC3339 or YYYY-MM-DD, or send subscription_days",
				"blocked_until":       "Use RFC3339 or YYYY-MM-DD",
				"subscription_plan":   "Use basic or pro",
				"subscription_days":   "Use an integer from 1 to 365",
				"subscriptionDays":    "Use an integer from 1 to 365",
			})
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Psychologist not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update psychologist access", nil)
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	value := c.Param(name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
			name: "Invalid id",
		})
		return 0, false
	}

	return id, true
}

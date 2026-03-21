package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

func (h *Handler) CreateFormulaRule(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.CreateFormulaRuleInput
	if !bindJSON(c, &input) {
		return
	}

	rule, err := h.appService.CreateFormulaRule(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFormulaRuleInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"name":           "Name is required",
				"condition_type": "Invalid formula rule payload",
			})
		case errors.Is(err, service.ErrTestNotFound):
			writeError(c, http.StatusNotFound, "Test not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to create formula rule", nil)
		}
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) ListFormulaRules(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	rules, err := h.appService.ListFormulaRules(c.Request.Context(), user.ID, testID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list formula rules", nil)
		return
	}

	c.JSON(http.StatusOK, rules)
}

func (h *Handler) GetFormulaRule(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	ruleID, ok := parseIDParam(c, "ruleId")
	if !ok {
		return
	}

	rule, err := h.appService.GetFormulaRuleByID(c.Request.Context(), user.ID, testID, ruleID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFormulaRuleNotFound):
			writeError(c, http.StatusNotFound, "Formula rule not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to load formula rule", nil)
		}
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *Handler) UpdateFormulaRule(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	ruleID, ok := parseIDParam(c, "ruleId")
	if !ok {
		return
	}

	var input domain.UpdateFormulaRuleInput
	if !bindJSON(c, &input) {
		return
	}

	rule, err := h.appService.UpdateFormulaRule(c.Request.Context(), user.ID, testID, ruleID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFormulaRuleInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"name":           "Name is required",
				"condition_type": "Invalid formula rule payload",
			})
		case errors.Is(err, service.ErrFormulaRuleNotFound):
			writeError(c, http.StatusNotFound, "Formula rule not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update formula rule", nil)
		}
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *Handler) DeleteFormulaRule(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	ruleID, ok := parseIDParam(c, "ruleId")
	if !ok {
		return
	}

	err := h.appService.DeleteFormulaRule(c.Request.Context(), user.ID, testID, ruleID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrFormulaRuleNotFound):
			writeError(c, http.StatusNotFound, "Formula rule not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to delete formula rule", nil)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) CalculateFormulaPreview(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.CalculateFormulaInput
	if !bindJSON(c, &input) {
		return
	}

	result, err := h.appService.CalculateFormulaPreview(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to calculate formula preview", nil)
		return
	}

	c.JSON(http.StatusOK, result)
}

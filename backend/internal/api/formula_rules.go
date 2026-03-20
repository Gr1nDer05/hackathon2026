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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.appService.CreateFormulaRule(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFormulaRuleInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formula rule payload"})
		case errors.Is(err, service.ErrTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create formula rule"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list formula rules"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "formula rule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load formula rule"})
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule, err := h.appService.UpdateFormulaRule(c.Request.Context(), user.ID, testID, ruleID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFormulaRuleInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formula rule payload"})
		case errors.Is(err, service.ErrFormulaRuleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "formula rule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update formula rule"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "formula rule not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete formula rule"})
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.appService.CalculateFormulaPreview(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate formula preview"})
		return
	}

	c.JSON(http.StatusOK, result)
}

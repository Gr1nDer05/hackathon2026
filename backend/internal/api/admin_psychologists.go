package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

func (h *Handler) CreatePsychologistByAdmin(c *gin.Context) {
	var input domain.CreatePsychologistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.appService.CreatePsychologistByAdmin(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "psychologist with this email already exists"})
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, gin.H{"error": "email, full name and password of at least 8 characters are required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create psychologist"})
		}
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func (h *Handler) ListPsychologists(c *gin.Context) {
	users, err := h.appService.ListPsychologists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list psychologists"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load psychologist workspace"})
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.appService.UpdatePsychologistAccount(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, gin.H{"error": "email and full name are required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update psychologist account"})
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.appService.UpdatePsychologistProfile(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update psychologist profile"})
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.appService.UpdatePsychologistCard(c.Request.Context(), userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update psychologist card"})
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.appService.UpdatePsychologistAccess(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPsychologistAccess):
			c.JSON(http.StatusBadRequest, gin.H{"error": "provide at least one access field; dates must be RFC3339 or YYYY-MM-DD"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update psychologist access"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	value := c.Param(name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}

	return id, true
}

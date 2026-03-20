package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

func (h *Handler) CreatePsychologistTest(c *gin.Context) {
	user := mustPsychologist(c)

	var input domain.CreateTestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	test, err := h.appService.CreatePsychologistTest(c.Request.Context(), user.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTestInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "title is required, recommended_duration and max_participants must be >= 0, status must be draft or published"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create test"})
		}
		return
	}

	c.JSON(http.StatusCreated, test)
}

func (h *Handler) ListPsychologistTests(c *gin.Context) {
	user := mustPsychologist(c)

	tests, err := h.appService.ListPsychologistTests(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tests"})
		return
	}

	c.JSON(http.StatusOK, tests)
}

func (h *Handler) GetPsychologistTest(c *gin.Context) {
	user := mustPsychologist(c)

	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	test, err := h.appService.GetPsychologistTestByID(c.Request.Context(), user.ID, testID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get test"})
		}
		return
	}

	c.JSON(http.StatusOK, test)
}

func (h *Handler) UpdatePsychologistTest(c *gin.Context) {
	user := mustPsychologist(c)

	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.UpdateTestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	test, err := h.appService.UpdatePsychologistTest(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTestInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "title is required, recommended_duration and max_participants must be >= 0, status must be draft or published"})
		case errors.Is(err, service.ErrTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update test"})
		}
		return
	}

	c.JSON(http.StatusOK, test)
}

func (h *Handler) DeletePsychologistTest(c *gin.Context) {
	user := mustPsychologist(c)

	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.appService.DeletePsychologistTest(c.Request.Context(), user.ID, testID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete test"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

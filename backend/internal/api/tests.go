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
	if !bindJSON(c, &input) {
		return
	}

	test, err := h.appService.CreatePsychologistTest(c.Request.Context(), user.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTestInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"title":                 "Title is required",
				"recommended_duration":  "Must be greater than or equal to 0",
				"max_participants":      "Must be greater than or equal to 0 and greater than 0 when has_participant_limit is true",
				"status":                "Use draft or published",
				"has_participant_limit": "Set max_participants when the limit is enabled",
			})
		default:
			writeError(c, http.StatusInternalServerError, "Failed to create test", nil)
		}
		return
	}

	c.JSON(http.StatusCreated, test)
}

func (h *Handler) ListPsychologistTests(c *gin.Context) {
	user := mustPsychologist(c)

	tests, err := h.appService.ListPsychologistTests(c.Request.Context(), user.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list tests", nil)
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
			writeError(c, http.StatusNotFound, "Test not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to get test", nil)
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
	if !bindJSON(c, &input) {
		return
	}

	test, err := h.appService.UpdatePsychologistTest(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidTestInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"title":                 "Title is required",
				"recommended_duration":  "Must be greater than or equal to 0",
				"max_participants":      "Must be greater than or equal to 0 and greater than 0 when has_participant_limit is true",
				"status":                "Use draft or published",
				"has_participant_limit": "Set max_participants when the limit is enabled",
			})
		case errors.Is(err, service.ErrTestNotFound):
			writeError(c, http.StatusNotFound, "Test not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update test", nil)
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
			writeError(c, http.StatusNotFound, "Test not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to delete test", nil)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

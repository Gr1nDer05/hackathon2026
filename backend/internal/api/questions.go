package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

func (h *Handler) CreatePsychologistQuestion(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var input domain.CreateQuestionInput
	if !bindJSON(c, &input) {
		return
	}

	question, err := h.appService.CreatePsychologistQuestion(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidQuestionInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"text":          "Question text is required",
				"question_type": "Invalid question payload",
			})
		case errors.Is(err, service.ErrTestNotFound):
			writeError(c, http.StatusNotFound, "Test not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to create question", nil)
		}
		return
	}

	c.JSON(http.StatusCreated, question)
}

func (h *Handler) ListPsychologistQuestions(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	questions, err := h.appService.ListPsychologistQuestions(c.Request.Context(), user.ID, testID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list questions", nil)
		return
	}

	c.JSON(http.StatusOK, questions)
}

func (h *Handler) GetPsychologistQuestion(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	questionID, ok := parseIDParam(c, "questionId")
	if !ok {
		return
	}

	question, err := h.appService.GetPsychologistQuestionByID(c.Request.Context(), user.ID, testID, questionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrQuestionNotFound):
			writeError(c, http.StatusNotFound, "Question not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to load question", nil)
		}
		return
	}

	c.JSON(http.StatusOK, question)
}

func (h *Handler) UpdatePsychologistQuestion(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	questionID, ok := parseIDParam(c, "questionId")
	if !ok {
		return
	}

	var input domain.UpdateQuestionInput
	if !bindJSON(c, &input) {
		return
	}

	question, err := h.appService.UpdatePsychologistQuestion(c.Request.Context(), user.ID, testID, questionID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidQuestionInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"text":          "Question text is required",
				"question_type": "Invalid question payload",
			})
		case errors.Is(err, service.ErrQuestionNotFound):
			writeError(c, http.StatusNotFound, "Question not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update question", nil)
		}
		return
	}

	c.JSON(http.StatusOK, question)
}

func (h *Handler) DeletePsychologistQuestion(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	questionID, ok := parseIDParam(c, "questionId")
	if !ok {
		return
	}

	err := h.appService.DeletePsychologistQuestion(c.Request.Context(), user.ID, testID, questionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrQuestionNotFound):
			writeError(c, http.StatusNotFound, "Question not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to delete question", nil)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

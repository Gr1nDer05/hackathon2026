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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := h.appService.CreatePsychologistQuestion(c.Request.Context(), user.ID, testID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidQuestionInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question payload"})
		case errors.Is(err, service.ErrTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create question"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list questions"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load question"})
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
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := h.appService.UpdatePsychologistQuestion(c.Request.Context(), user.ID, testID, questionID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidQuestionInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question payload"})
		case errors.Is(err, service.ErrQuestionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update question"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete question"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

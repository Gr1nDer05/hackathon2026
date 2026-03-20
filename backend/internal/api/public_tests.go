package api

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) PublishPsychologistTest(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	response, err := h.appService.PublishPsychologistTest(c.Request.Context(), user.ID, testID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "test not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish test"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetPublicTest(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	test, err := h.appService.GetPublicTestBySlug(c.Request.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "public test not found"})
		case errors.Is(err, service.ErrPublicTestLimitReached):
			c.JSON(http.StatusForbidden, gin.H{"error": "test participant limit reached"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load public test"})
		}
		return
	}

	c.JSON(http.StatusOK, test)
}

func (h *Handler) StartPublicTest(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))

	var input domain.StartPublicTestInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.appService.StartPublicTest(c.Request.Context(), slug, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "public test not found"})
		case errors.Is(err, service.ErrPublicTestLimitReached):
			c.JSON(http.StatusForbidden, gin.H{"error": "test participant limit reached"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start public test"})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) SubmitPublicTest(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))

	var input domain.SubmitPublicTestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.appService.SubmitPublicTest(c.Request.Context(), slug, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "public test not found"})
		case errors.Is(err, service.ErrInvalidPublicTestSubmission):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public test submission"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit public test"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) ListPsychologistTestSubmissions(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	submissions, err := h.appService.ListPsychologistTestSubmissions(c.Request.Context(), user.ID, testID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list test submissions"})
		return
	}

	c.JSON(http.StatusOK, submissions)
}

func (h *Handler) GetPsychologistTestSubmission(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	sessionID, ok := parseIDParam(c, "sessionId")
	if !ok {
		return
	}

	submission, err := h.appService.GetPsychologistTestSubmissionByID(c.Request.Context(), user.ID, testID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load test submission"})
		}
		return
	}

	c.JSON(http.StatusOK, submission)
}

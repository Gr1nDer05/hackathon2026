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
			writeError(c, http.StatusNotFound, "Test not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to publish test", nil)
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetPublicTest(c *gin.Context) {
	slug := c.Param("slug")
	test, err := h.appService.GetPublicTestBySlug(c.Request.Context(), slug)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			writeError(c, http.StatusNotFound, "Public test not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to load public test", nil)
		}
		return
	}

	c.JSON(http.StatusOK, test)
}

func (h *Handler) GetPublicTestReport(c *gin.Context) {
	slug := c.Param("slug")
	accessToken := strings.TrimSpace(c.Query("access_token"))
	if accessToken == "" {
		writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
			"access_token": "Access token is required",
		})
		return
	}

	report, err := h.appService.GeneratePublicClientReport(c.Request.Context(), slug, accessToken, c.Query("format"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			writeError(c, http.StatusNotFound, "Public test report not found", nil)
		case errors.Is(err, service.ErrPublicClientReportUnavailable):
			writeError(c, http.StatusForbidden, "Client report is not available for this test", singleFieldError("report", "Client report is not available for this test"))
		case errors.Is(err, service.ErrReportNotReady):
			writeError(c, http.StatusConflict, "Report is available only for completed test sessions", singleFieldError("access_token", "Report is available only for completed test sessions"))
		case errors.Is(err, service.ErrInvalidReportFormat):
			writeError(c, http.StatusBadRequest, "Unsupported report format", map[string]string{
				"format": "Use html or docx",
			})
		default:
			writeError(c, http.StatusInternalServerError, "Failed to generate public test report", nil)
		}
		return
	}

	writeGeneratedReport(c, report)
}

func (h *Handler) StartPublicTest(c *gin.Context) {
	slug := c.Param("slug")

	var input domain.StartPublicTestInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		message, fieldErrors := describeBindingError(err, &input)
		writeError(c, http.StatusBadRequest, message, fieldErrors)
		return
	}

	response, err := h.appService.StartPublicTest(c.Request.Context(), slug, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			writeError(c, http.StatusNotFound, "Public test not found", nil)
		case errors.Is(err, service.ErrInvalidPublicTestRespondent):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"respondent_name":      "Respondent full name is required",
				"respondent_phone":     "Respondent phone is required",
				"respondent_email":     "Fill enabled personal fields",
				"respondent_age":       "Fill enabled personal fields",
				"respondent_gender":    "Fill enabled personal fields",
				"respondent_education": "Fill enabled personal fields",
			})
		case errors.Is(err, service.ErrPublicTestAlreadyTaken):
			writeError(c, http.StatusConflict, "Respondent with this phone number has already taken the test", singleFieldError("respondent_phone", "Respondent with this phone number has already taken the test"))
		case errors.Is(err, service.ErrPublicTestLimitReached):
			writeError(c, http.StatusForbidden, "Test participant limit reached", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to start public test", nil)
		}
		return
	}

	statusCode := http.StatusCreated
	if response.Resumed {
		statusCode = http.StatusOK
	}
	c.JSON(statusCode, response)
}

func (h *Handler) SavePublicTestProgress(c *gin.Context) {
	slug := c.Param("slug")

	var input domain.SubmitPublicTestInput
	if !bindJSON(c, &input) {
		return
	}

	response, err := h.appService.SavePublicTestProgress(c.Request.Context(), slug, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			writeError(c, http.StatusNotFound, "Public test not found", nil)
		case errors.Is(err, service.ErrInvalidPublicTestSubmission):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"access_token": "Invalid public test progress payload",
				"answers":      "Invalid public test progress payload",
			})
		case errors.Is(err, service.ErrPublicTestAlreadyTaken):
			writeError(c, http.StatusConflict, "Test is already completed for this respondent", singleFieldError("access_token", "Test is already completed for this respondent"))
		default:
			writeError(c, http.StatusInternalServerError, "Failed to save public test progress", nil)
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) SubmitPublicTest(c *gin.Context) {
	slug := c.Param("slug")

	var input domain.SubmitPublicTestInput
	if !bindJSON(c, &input) {
		return
	}

	response, err := h.appService.SubmitPublicTest(c.Request.Context(), slug, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPublicTestNotFound):
			writeError(c, http.StatusNotFound, "Public test not found", nil)
		case errors.Is(err, service.ErrInvalidPublicTestSubmission):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"access_token": "Invalid public test submission",
				"answers":      "Invalid public test submission",
			})
		case errors.Is(err, service.ErrPublicTestAlreadyTaken):
			writeError(c, http.StatusConflict, "Test is already completed for this respondent", singleFieldError("access_token", "Test is already completed for this respondent"))
		default:
			writeError(c, http.StatusInternalServerError, "Failed to submit public test", nil)
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetPsychologistSubmissionBySessionID(c *gin.Context) {
	user := mustPsychologist(c)
	sessionID, ok := parseIDParam(c, "sessionId")
	if !ok {
		return
	}

	submission, err := h.appService.GetPsychologistTestSubmissionBySessionID(c.Request.Context(), user.ID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTestNotFound):
			writeError(c, http.StatusNotFound, "Submission not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to load test submission", nil)
		}
		return
	}

	c.JSON(http.StatusOK, submission)
}

func (h *Handler) ListPsychologistTestSubmissions(c *gin.Context) {
	user := mustPsychologist(c)
	testID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	submissions, err := h.appService.ListPsychologistTestSubmissions(c.Request.Context(), user.ID, testID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list test submissions", nil)
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
			writeError(c, http.StatusNotFound, "Submission not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to load test submission", nil)
		}
		return
	}

	c.JSON(http.StatusOK, submission)
}

package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPsychologistReportBySessionID(c *gin.Context) {
	user := mustPsychologist(c)
	sessionID, ok := parseIDParam(c, "sessionId")
	if !ok {
		return
	}

	audience := strings.TrimSpace(strings.ToLower(c.DefaultQuery("audience", "psychologist")))

	var (
		report service.GeneratedReport
		err    error
	)
	switch audience {
	case "", "psychologist":
		report, err = h.appService.GeneratePsychologistReportBySessionID(c.Request.Context(), user.ID, sessionID, c.Query("format"))
	case "client":
		report, err = h.appService.GenerateClientReportBySessionID(c.Request.Context(), user.ID, sessionID, c.Query("format"))
	default:
		writeError(c, http.StatusBadRequest, "Unsupported report audience", map[string]string{
			"audience": "Use psychologist or client",
		})
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, service.ErrTestNotFound):
			writeError(c, http.StatusNotFound, "Submission not found", nil)
		case errors.Is(err, service.ErrReportNotReady):
			writeError(c, http.StatusConflict, "Report is available only for completed test sessions", nil)
		case errors.Is(err, service.ErrInvalidReportFormat):
			writeError(c, http.StatusBadRequest, "Unsupported report format", map[string]string{
				"format": "Use html or docx",
			})
		default:
			writeError(c, http.StatusInternalServerError, "Failed to generate report", nil)
		}
		return
	}

	writeGeneratedReport(c, report)
}

func writeGeneratedReport(c *gin.Context, report service.GeneratedReport) {
	disposition := "inline"
	if report.ContentType == service.ReportContentTypeDOCX {
		disposition = "attachment"
	}

	c.Header("Content-Type", report.ContentType)
	c.Header("Content-Disposition", disposition+`; filename="`+report.Filename+`"`)
	c.Data(http.StatusOK, report.ContentType, report.Content)
}

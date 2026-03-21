package api

import (
	"errors"
	"net/http"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateReportTemplate(c *gin.Context) {
	user := mustPsychologist(c)

	var input domain.CreateReportTemplateInput
	if !bindJSON(c, &input) {
		return
	}

	template, err := h.appService.CreateReportTemplate(c.Request.Context(), user.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidReportTemplateInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"name":          "Template name is required",
				"template_body": "Use a valid JSON template configuration",
			})
		default:
			writeError(c, http.StatusInternalServerError, "Failed to create report template", nil)
		}
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *Handler) ListReportTemplates(c *gin.Context) {
	user := mustPsychologist(c)

	templates, err := h.appService.ListReportTemplates(c.Request.Context(), user.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to list report templates", nil)
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (h *Handler) GetReportTemplate(c *gin.Context) {
	user := mustPsychologist(c)
	templateID, ok := parseIDParam(c, "templateId")
	if !ok {
		return
	}

	template, err := h.appService.GetReportTemplateByID(c.Request.Context(), user.ID, templateID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReportTemplateNotFound):
			writeError(c, http.StatusNotFound, "Report template not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to load report template", nil)
		}
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *Handler) UpdateReportTemplate(c *gin.Context) {
	user := mustPsychologist(c)
	templateID, ok := parseIDParam(c, "templateId")
	if !ok {
		return
	}

	var input domain.UpdateReportTemplateInput
	if !bindJSON(c, &input) {
		return
	}

	template, err := h.appService.UpdateReportTemplate(c.Request.Context(), user.ID, templateID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidReportTemplateInput):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"name":          "Template name is required",
				"template_body": "Use a valid JSON template configuration",
			})
		case errors.Is(err, service.ErrReportTemplateNotFound):
			writeError(c, http.StatusNotFound, "Report template not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update report template", nil)
		}
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *Handler) DeleteReportTemplate(c *gin.Context) {
	user := mustPsychologist(c)
	templateID, ok := parseIDParam(c, "templateId")
	if !ok {
		return
	}

	if err := h.appService.DeleteReportTemplate(c.Request.Context(), user.ID, templateID); err != nil {
		switch {
		case errors.Is(err, service.ErrReportTemplateNotFound):
			writeError(c, http.StatusNotFound, "Report template not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to delete report template", nil)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

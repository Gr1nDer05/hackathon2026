package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Message     string            `json:"message"`
	FieldErrors map[string]string `json:"field_errors"`
}

func writeError(c *gin.Context, status int, message string, fieldErrors map[string]string) {
	if message == "" {
		message = http.StatusText(status)
	}

	c.JSON(status, errorResponse{
		Message:     message,
		FieldErrors: fieldErrors,
	})
}

func abortWithError(c *gin.Context, status int, message string, fieldErrors map[string]string) {
	writeError(c, status, message, fieldErrors)
	c.Abort()
}

func writeValidationError(c *gin.Context, fieldErrors map[string]string) {
	writeError(c, http.StatusBadRequest, "Validation failed", fieldErrors)
}

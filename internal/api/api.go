package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

type Handler struct {
	appService *service.AppService
	db         *sql.DB
}

func NewHandler(appService *service.AppService, db *sql.DB) *Handler {
	return &Handler{
		appService: appService,
		db:         db,
	}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, h.appService.Status())
}

func (h *Handler) TestDB(c *gin.Context) {
	if err := h.db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "database connection failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "database connection successful",
	})
}

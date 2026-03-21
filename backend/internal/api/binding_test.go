package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func TestConfigureBindingRejectsUnknownJSONFields(t *testing.T) {
	t.Helper()

	old := binding.EnableDecoderDisallowUnknownFields
	t.Cleanup(func() {
		binding.EnableDecoderDisallowUnknownFields = old
	})

	ConfigureBinding()
	gin.SetMode(gin.TestMode)

	type payload struct {
		Name string `json:"name" binding:"required"`
	}

	router := gin.New()
	router.POST("/strict", func(c *gin.Context) {
		var input payload
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodPost, "/strict", strings.NewReader(`{"name":"ok","extra":"nope"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

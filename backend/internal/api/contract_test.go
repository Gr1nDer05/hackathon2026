package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func TestBindJSONReturnsUnifiedFieldErrors(t *testing.T) {
	t.Helper()

	old := binding.EnableDecoderDisallowUnknownFields
	t.Cleanup(func() {
		binding.EnableDecoderDisallowUnknownFields = old
	})

	ConfigureBinding()
	gin.SetMode(gin.TestMode)

	type payload struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	router := gin.New()
	router.POST("/login", func(c *gin.Context) {
		var input payload
		if !bindJSON(c, &input) {
			return
		}

		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	var response errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Message != "Validation failed" {
		t.Fatalf("expected validation message, got %q", response.Message)
	}

	if response.FieldErrors["email"] != "This field is required" {
		t.Fatalf("expected email field error, got %#v", response.FieldErrors)
	}
}

func TestCORSPreflightReturnsBrowserHeaders(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	handler := &Handler{
		allowedOrigins: map[string]struct{}{
			"http://localhost:5173": {},
		},
	}

	router := gin.New()
	router.Use(handler.CORS())
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/auth/admin/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, X-CSRF-Token")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNoContent, rec.Code, rec.Body.String())
	}

	if origin := rec.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:5173" {
		t.Fatalf("expected allow origin header, got %q", origin)
	}

	if credentials := rec.Header().Get("Access-Control-Allow-Credentials"); credentials != "true" {
		t.Fatalf("expected credentials header, got %q", credentials)
	}

	allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	if !strings.Contains(allowHeaders, "Content-Type") || !strings.Contains(allowHeaders, "X-CSRF-Token") {
		t.Fatalf("expected Content-Type and X-CSRF-Token in allow headers, got %q", allowHeaders)
	}
}

func TestFrontendPsychologistValidationRules(t *testing.T) {
	t.Helper()

	createErrors := validateCreatePsychologistInput(domain.CreatePsychologistInput{
		Email:    "bad-email",
		Password: "password",
		FullName: "Ivanov Ivan",
	})
	if createErrors["email"] == "" || createErrors["password"] == "" || createErrors["full_name"] == "" {
		t.Fatalf("expected email, password and full_name errors, got %#v", createErrors)
	}

	profileErrors := validatePsychologistProfileInput(domain.UpdatePsychologistProfileInput{
		City:           "Berlin",
		Specialization: "IT",
	}, true)
	if profileErrors["city"] == "" || profileErrors["specialization"] == "" {
		t.Fatalf("expected city and specialization errors, got %#v", profileErrors)
	}

	cardErrors := validatePsychologistCardInput(domain.UpdatePsychologistCardInput{
		ContactPhone: "89991234567",
	}, true)
	if cardErrors["contact_phone"] == "" {
		t.Fatalf("expected contact_phone error, got %#v", cardErrors)
	}
}

func TestFrontendCompatibilityAliasRoutesAreRegistered(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &Handler{}
	handler.RegisterRoutes(router)

	routes := make(map[string]struct{}, len(router.Routes()))
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	required := []string{
		"PUT /admins/me",
		"POST /admins/me/email/verification-code",
		"POST /admins/me/email/confirm",
		"GET /psychologists/results/:sessionId",
		"GET /public/sessions/:token",
		"POST /public/sessions/:token/start",
		"PUT /public/sessions/:token/answers",
		"POST /public/sessions/:token/complete",
	}

	for _, route := range required {
		if _, ok := routes[route]; !ok {
			t.Fatalf("expected route %q to be registered", route)
		}
	}
}

func TestRequireAdminEmailBoundBlocksPlaceholderEmail(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(authenticatedAdminKey, domain.AuthenticatedUser{
			Email: "admin@admin.local",
			Role:  domain.RoleAdmin,
		})
		c.Next()
	})
	router.Use(handler.RequireAdminEmailBound())
	router.GET("/admins/psychologists", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admins/psychologists", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestRequireAdminEmailVerifiedBlocksUnverifiedEmail(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(authenticatedAdminKey, domain.AuthenticatedUser{
			Email: "admin@example.com",
			Role:  domain.RoleAdmin,
		})
		c.Next()
	})
	router.Use(handler.RequireAdminEmailVerified())
	router.GET("/admins/psychologists", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admins/psychologists", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

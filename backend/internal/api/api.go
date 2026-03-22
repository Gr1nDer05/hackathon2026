package api

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Gr1nDer05/Hackathon2026/docs"
	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

type Handler struct {
	appService       *service.AppService
	db               *sql.DB
	secureCookies    bool
	allowedOrigins   map[string]struct{}
	loginRateLimiter *loginRateLimiter
}

func NewHandler(appService *service.AppService, db *sql.DB) *Handler {
	return &Handler{
		appService:       appService,
		db:               db,
		secureCookies:    isSecureCookiesEnabled(),
		allowedOrigins:   loadAllowedOrigins(),
		loginRateLimiter: newLoginRateLimiter(10*time.Minute, 5, 15*time.Minute),
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.Use(h.SecurityHeaders(), h.CORS())
	docs.RegisterRoutes(router)
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	router.GET("/health", h.Health)
	router.GET("/testdb", h.TestDB)

	router.POST("/auth/psychologists/login", h.LoginPsychologist)
	router.POST("/auth/psychologists/logout", h.RequirePsychologistAuth(), h.RequireCSRFCookie(), h.LogoutPsychologist)
	router.POST("/auth/admin/login", h.LoginAdmin)
	router.POST("/auth/admin/logout", h.RequireAdminAuth(), h.RequireCSRFCookie(), h.LogoutAdmin)

	psychologists := router.Group("/psychologists/me")
	psychologists.Use(h.RequirePsychologistAuth(), h.RequireCSRFCookie())
	psychologists.GET("", h.GetPsychologistWorkspace)
	psychologists.GET("/profile", h.GetPsychologistProfile)
	psychologists.PUT("/profile", h.UpdatePsychologistProfile)
	psychologists.GET("/card", h.GetPsychologistCard)
	psychologists.PUT("/card", h.UpdatePsychologistCard)
	psychologists.POST("/subscription/purchase", h.CreateSubscriptionPurchaseRequest)

	reportTemplates := router.Group("/psychologists/report-templates")
	reportTemplates.Use(h.RequirePsychologistAuth(), h.RequirePsychologistActiveSubscription(), h.RequireCSRFCookie())
	reportTemplates.POST("/generate", h.GenerateReportTemplateDraft)
	reportTemplates.POST("", h.CreateReportTemplate)
	reportTemplates.GET("", h.ListReportTemplates)
	reportTemplates.GET("/:templateId", h.GetReportTemplate)
	reportTemplates.PUT("/:templateId", h.UpdateReportTemplate)
	reportTemplates.DELETE("/:templateId", h.DeleteReportTemplate)

	psychologistTests := router.Group("/psychologists/tests")
	psychologistTests.Use(h.RequirePsychologistAuth(), h.RequirePsychologistActiveSubscription(), h.RequireCSRFCookie())
	psychologistTests.POST("", h.CreatePsychologistTest)
	psychologistTests.GET("", h.ListPsychologistTests)
	psychologistTests.GET("/:id", h.GetPsychologistTest)
	psychologistTests.PUT("/:id", h.UpdatePsychologistTest)
	psychologistTests.DELETE("/:id", h.DeletePsychologistTest)
	psychologistTests.POST("/:id/questions", h.CreatePsychologistQuestion)
	psychologistTests.GET("/:id/questions", h.ListPsychologistQuestions)
	psychologistTests.GET("/:id/questions/:questionId", h.GetPsychologistQuestion)
	psychologistTests.PUT("/:id/questions/:questionId", h.UpdatePsychologistQuestion)
	psychologistTests.DELETE("/:id/questions/:questionId", h.DeletePsychologistQuestion)
	psychologistTests.POST("/:id/publish", h.PublishPsychologistTest)
	psychologistTests.GET("/:id/results", h.ListPsychologistTestSubmissions)
	psychologistTests.GET("/:id/results/:sessionId", h.GetPsychologistTestSubmission)
	psychologistTests.POST("/:id/formulas", h.CreateFormulaRule)
	psychologistTests.GET("/:id/formulas", h.ListFormulaRules)
	psychologistTests.GET("/:id/formulas/:ruleId", h.GetFormulaRule)
	psychologistTests.PUT("/:id/formulas/:ruleId", h.UpdateFormulaRule)
	psychologistTests.DELETE("/:id/formulas/:ruleId", h.DeleteFormulaRule)
	psychologistTests.POST("/:id/formulas/calculate", h.CalculateFormulaPreview)

	psychologistResults := router.Group("/psychologists/results")
	psychologistResults.Use(h.RequirePsychologistAuth(), h.RequirePsychologistActiveSubscription(), h.RequireCSRFCookie())
	psychologistResults.GET("/:sessionId", h.GetPsychologistSubmissionBySessionID)
	psychologistResults.GET("/:sessionId/report", h.GetPsychologistReportBySessionID)

	publicTests := router.Group("/public/tests")
	publicTests.GET("/:slug", h.GetPublicTest)
	publicTests.GET("/:slug/report", h.GetPublicTestReport)
	publicTests.POST("/:slug/start", h.StartPublicTest)
	publicTests.POST("/:slug/progress", h.SavePublicTestProgress)
	publicTests.POST("/:slug/submit", h.SubmitPublicTest)

	admins := router.Group("/admins/me")
	admins.Use(h.RequireAdminAuth(), h.RequireCSRFCookie())
	admins.GET("", h.GetAdminMe)
	admins.GET("/subscription-purchase-requests", h.ListPendingSubscriptionPurchaseRequests)
	admins.PUT("", h.UpdateAdminMe)

	adminPsychologists := router.Group("/admins/psychologists")
	adminPsychologists.Use(h.RequireAdminAuth(), h.RequireCSRFCookie())
	adminPsychologists.POST("", h.CreatePsychologistByAdmin)
	adminPsychologists.GET("", h.ListPsychologists)
	adminPsychologists.GET("/:id/workspace", h.GetPsychologistWorkspaceByAdmin)
	adminPsychologists.PUT("/:id", h.UpdatePsychologistAccountByAdmin)
	adminPsychologists.PUT("/:id/access", h.UpdatePsychologistAccessByAdmin)
	adminPsychologists.PUT("/:id/profile", h.UpdatePsychologistProfileByAdmin)
	adminPsychologists.PUT("/:id/card", h.UpdatePsychologistCardByAdmin)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, h.appService.Status())
}

func (h *Handler) TestDB(c *gin.Context) {
	if err := h.db.Ping(); err != nil {
		writeError(c, http.StatusServiceUnavailable, "Database connection failed", map[string]string{
			"database": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "database connection successful",
	})
}

type loginAttempt struct {
	count       int
	windowStart time.Time
	blockedTill time.Time
}

type loginRateLimiter struct {
	mu          sync.Mutex
	window      time.Duration
	maxAttempts int
	blockFor    time.Duration
	attempts    map[string]loginAttempt
}

func newLoginRateLimiter(window time.Duration, maxAttempts int, blockFor time.Duration) *loginRateLimiter {
	return &loginRateLimiter{
		window:      window,
		maxAttempts: maxAttempts,
		blockFor:    blockFor,
		attempts:    make(map[string]loginAttempt),
	}
}

func (l *loginRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	attempt := l.attempts[ip]
	if attempt.blockedTill.After(now) {
		return false
	}

	if attempt.windowStart.IsZero() || now.Sub(attempt.windowStart) > l.window {
		attempt = loginAttempt{
			count:       0,
			windowStart: now,
		}
	}

	attempt.count++
	if attempt.count > l.maxAttempts {
		attempt.blockedTill = now.Add(l.blockFor)
		attempt.count = 0
		attempt.windowStart = now
		l.attempts[ip] = attempt
		return false
	}

	l.attempts[ip] = attempt
	return true
}

func (l *loginRateLimiter) Reset(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

func isSecureCookiesEnabled() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("COOKIE_SECURE")), "true") {
		return true
	}

	return strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}

func loadAllowedOrigins() map[string]struct{} {
	raw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	result := make(map[string]struct{})
	if raw == "" {
		return result
	}

	for _, item := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(item)
		if origin != "" {
			result[origin] = struct{}{}
		}
	}

	return result
}

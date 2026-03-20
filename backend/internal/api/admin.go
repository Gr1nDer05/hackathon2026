package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

const authenticatedAdminKey = "authenticated_admin"
const adminSessionCookieName = "admin_session_id"

func (h *Handler) LoginAdmin(c *gin.Context) {
	clientIP := c.ClientIP()
	if !h.loginRateLimiter.Allow("admin:" + clientIP) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, try again later"})
		return
	}

	var input domain.AdminLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID, response, err := h.appService.LoginAdmin(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login admin"})
		}
		return
	}

	h.loginRateLimiter.Reset("admin:" + clientIP)
	h.setAdminSessionCookie(c, sessionID)
	if err := h.issueCSRFCookie(c); err != nil {
		h.clearAdminSessionCookie(c)
		h.clearCSRFCookie(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize session security"})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) LogoutAdmin(c *gin.Context) {
	sessionID, _ := c.Cookie(adminSessionCookieName)
	if err := h.appService.LogoutAdmin(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout admin"})
		return
	}

	h.clearAdminSessionCookie(c)
	h.clearCSRFCookie(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) RequireAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(adminSessionCookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or expired admin session"})
			c.Abort()
			return
		}

		user, err := h.appService.AuthenticateAdmin(c.Request.Context(), sessionID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrForbidden):
				c.JSON(http.StatusForbidden, gin.H{"error": "access is allowed only for admins"})
			default:
				h.clearAdminSessionCookie(c)
				h.clearCSRFCookie(c)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired admin session"})
			}
			c.Abort()
			return
		}

		c.Set(authenticatedAdminKey, user)
		c.Next()
	}
}

func (h *Handler) GetAdminMe(c *gin.Context) {
	user := mustAdmin(c)

	admin, err := h.appService.GetAdminMe(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load admin account"})
		return
	}

	c.JSON(http.StatusOK, admin)
}

func (h *Handler) ListAdminNotifications(c *gin.Context) {
	notifications, err := h.appService.ListAdminNotifications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load admin notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func mustAdmin(c *gin.Context) domain.AuthenticatedUser {
	value, _ := c.Get(authenticatedAdminKey)
	user, _ := value.(domain.AuthenticatedUser)
	return user
}

func (h *Handler) setAdminSessionCookie(c *gin.Context, sessionID string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		adminSessionCookieName,
		sessionID,
		int(service.SessionTTL.Seconds()),
		"/",
		"",
		h.secureCookies,
		true,
	)
}

func (h *Handler) clearAdminSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(adminSessionCookieName, "", -1, "/", "", h.secureCookies, true)
}

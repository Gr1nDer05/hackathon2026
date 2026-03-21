package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

const authenticatedAdminKey = "authenticated_admin"
const adminSessionCookieName = "admin_session_id"

func (h *Handler) LoginAdmin(c *gin.Context) {
	clientIP := c.ClientIP()
	if !h.loginRateLimiter.Allow("admin:" + clientIP) {
		writeError(c, http.StatusTooManyRequests, "Too many login attempts, try again later", nil)
		return
	}

	var input domain.AdminLoginInput
	if !bindJSON(c, &input) {
		return
	}

	sessionID, response, err := h.appService.LoginAdmin(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(c, http.StatusUnauthorized, "Invalid login or password", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to login admin", nil)
		}
		return
	}

	h.loginRateLimiter.Reset("admin:" + clientIP)
	h.setAdminSessionCookie(c, sessionID, response.ExpiresAt)
	if err := h.issueCSRFCookieWithExpiry(c, response.ExpiresAt); err != nil {
		h.clearAdminSessionCookie(c)
		h.clearCSRFCookie(c)
		writeError(c, http.StatusInternalServerError, "Failed to initialize session security", nil)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) LogoutAdmin(c *gin.Context) {
	sessionID, _ := c.Cookie(adminSessionCookieName)
	if err := h.appService.LogoutAdmin(c.Request.Context(), sessionID); err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to logout admin", nil)
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
			abortWithError(c, http.StatusUnauthorized, "Missing or expired admin session", nil)
			return
		}

		user, err := h.appService.AuthenticateAdmin(c.Request.Context(), sessionID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrForbidden):
				writeError(c, http.StatusForbidden, "Access is allowed only for admins", nil)
			default:
				h.clearAdminSessionCookie(c)
				h.clearCSRFCookie(c)
				writeError(c, http.StatusUnauthorized, "Invalid or expired admin session", nil)
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
		writeError(c, http.StatusInternalServerError, "Failed to load admin account", nil)
		return
	}

	c.JSON(http.StatusOK, admin)
}

func (h *Handler) UpdateAdminMe(c *gin.Context) {
	user := mustAdmin(c)

	var input domain.UpdateAdminMeInput
	if !bindJSON(c, &input) {
		return
	}

	admin, err := h.appService.UpdateAdminMe(c.Request.Context(), user.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAdminEmail):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"email": "Use a real email address",
			})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			writeError(c, http.StatusConflict, "Validation failed", map[string]string{
				"email": "Email already exists",
			})
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Admin not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to update admin account", nil)
		}
		return
	}

	c.JSON(http.StatusOK, admin)
}

func (h *Handler) SendAdminEmailVerificationCode(c *gin.Context) {
	user := mustAdmin(c)

	if err := h.appService.SendAdminEmailVerificationCode(c.Request.Context(), user.ID); err != nil {
		switch {
		case errors.Is(err, service.ErrAdminEmailNotBound):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"email": "Set a real email address before requesting a verification code",
			})
		case errors.Is(err, service.ErrMailerNotConfigured):
			writeError(c, http.StatusServiceUnavailable, "Email delivery is not configured", nil)
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Admin not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to send verification code", nil)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"ttl_seconds": int(service.AdminEmailVerificationCodeTTL / time.Second),
	})
}

func (h *Handler) ConfirmAdminEmail(c *gin.Context) {
	user := mustAdmin(c)

	var input domain.ConfirmAdminEmailInput
	if !bindJSON(c, &input) {
		return
	}

	admin, err := h.appService.ConfirmAdminEmail(c.Request.Context(), user.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAdminVerificationCode):
			writeError(c, http.StatusBadRequest, "Validation failed", map[string]string{
				"code": "Invalid or expired verification code",
			})
		case errors.Is(err, sql.ErrNoRows):
			writeError(c, http.StatusNotFound, "Admin not found", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to confirm admin email", nil)
		}
		return
	}

	c.JSON(http.StatusOK, admin)
}

func (h *Handler) ListAdminNotifications(c *gin.Context) {
	notifications, err := h.appService.ListAdminNotifications(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to load admin notifications", nil)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *Handler) RequireAdminEmailBound() gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := mustAdmin(c)
		if service.IsAdminEmailBound(admin.Email) {
			c.Next()
			return
		}

		abortWithError(c, http.StatusForbidden, "Bind your email before using the admin panel", map[string]string{
			"email": "Set a real email address in /admins/me",
		})
	}
}

func (h *Handler) RequireAdminEmailVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := mustAdmin(c)
		if !service.IsAdminEmailBound(admin.Email) {
			abortWithError(c, http.StatusForbidden, "Bind your email before using the admin panel", map[string]string{
				"email": "Set a real email address in /admins/me",
			})
			return
		}
		if !admin.EmailVerifiedAt.IsZero() {
			c.Next()
			return
		}

		abortWithError(c, http.StatusForbidden, "Confirm your email before using the admin panel", map[string]string{
			"email": "Request a code in /admins/me/email/verification-code and confirm it in /admins/me/email/confirm",
		})
	}
}

func mustAdmin(c *gin.Context) domain.AuthenticatedUser {
	value, _ := c.Get(authenticatedAdminKey)
	user, _ := value.(domain.AuthenticatedUser)
	return user
}

func (h *Handler) setAdminSessionCookie(c *gin.Context, sessionID string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		adminSessionCookieName,
		sessionID,
		cookieMaxAge(expiresAt),
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

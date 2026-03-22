package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

const authenticatedPsychologistKey = "authenticated_psychologist"
const psychologistSessionCookieName = "session_id"

func (h *Handler) LoginPsychologist(c *gin.Context) {
	clientIP := c.ClientIP()
	if !h.loginRateLimiter.Allow("psychologist:" + clientIP) {
		writeError(c, http.StatusTooManyRequests, "Too many login attempts, try again later", singleFieldError("email", "Too many login attempts, try again later"))
		return
	}

	var input domain.PsychologistLoginInput
	if !bindJSON(c, &input) {
		return
	}

	sessionID, response, err := h.appService.LoginPsychologist(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeError(c, http.StatusUnauthorized, "Invalid email or password", credentialFieldErrors("email", "Invalid email or password"))
		case errors.Is(err, service.ErrAccountDisabled):
			writeError(c, http.StatusForbidden, "Psychologist access is disabled by administrator", nil)
		case errors.Is(err, service.ErrPortalAccessExpired):
			writeError(c, http.StatusForbidden, "Portal access subscription has expired", nil)
		case errors.Is(err, service.ErrAccountTemporarilyBlocked):
			writeError(c, http.StatusForbidden, "Psychologist account is temporarily blocked", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to login psychologist", nil)
		}
		return
	}

	h.loginRateLimiter.Reset("psychologist:" + clientIP)
	h.setPsychologistSessionCookie(c, sessionID, response.ExpiresAt)
	if err := h.issueCSRFCookieWithExpiry(c, response.ExpiresAt); err != nil {
		h.clearPsychologistSessionCookie(c)
		h.clearCSRFCookie(c)
		writeError(c, http.StatusInternalServerError, "Failed to initialize session security", nil)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) RequirePsychologistAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(psychologistSessionCookieName)
		if err != nil {
			abortWithError(c, http.StatusUnauthorized, "Missing or expired session", nil)
			return
		}

		user, err := h.appService.AuthenticatePsychologist(c.Request.Context(), sessionID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrForbidden):
				writeError(c, http.StatusForbidden, "Access is allowed only for psychologists", nil)
			case errors.Is(err, service.ErrAccountDisabled):
				h.clearPsychologistSessionCookie(c)
				h.clearCSRFCookie(c)
				writeError(c, http.StatusForbidden, "Psychologist access is disabled by administrator", nil)
			case errors.Is(err, service.ErrAccountTemporarilyBlocked):
				h.clearPsychologistSessionCookie(c)
				h.clearCSRFCookie(c)
				writeError(c, http.StatusForbidden, "Psychologist account is temporarily blocked", nil)
			default:
				h.clearPsychologistSessionCookie(c)
				h.clearCSRFCookie(c)
				writeError(c, http.StatusUnauthorized, "Invalid or expired session", nil)
			}
			c.Abort()
			return
		}

		c.Set(authenticatedPsychologistKey, user)
		c.Next()
	}
}

func (h *Handler) RequirePsychologistActiveSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := mustPsychologist(c)
		if subscriptionExpiredForPsychologist(user, time.Now()) {
			abortWithError(c, http.StatusForbidden, "Portal access subscription has expired", singleFieldError("subscription_status", "Portal access subscription has expired"))
			return
		}

		c.Next()
	}
}

func subscriptionExpiredForPsychologist(user domain.AuthenticatedUser, now time.Time) bool {
	if user.Role != domain.RolePsychologist || user.PortalAccessUntil.IsZero() {
		return false
	}

	until, err := time.Parse(time.RFC3339, user.PortalAccessUntil.String())
	if err != nil {
		return false
	}

	return !until.After(now)
}

func (h *Handler) GetPsychologistWorkspace(c *gin.Context) {
	user := mustPsychologist(c)

	workspace, err := h.appService.GetPsychologistWorkspace(c.Request.Context(), user.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to load psychologist workspace", nil)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *Handler) GetPsychologistProfile(c *gin.Context) {
	user := mustPsychologist(c)

	workspace, err := h.appService.GetPsychologistWorkspace(c.Request.Context(), user.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to load psychologist profile", nil)
		return
	}

	c.JSON(http.StatusOK, workspace.Profile)
}

func (h *Handler) UpdatePsychologistProfile(c *gin.Context) {
	user := mustPsychologist(c)

	var input domain.UpdatePsychologistProfileInput
	if !bindJSON(c, &input) {
		return
	}

	if fieldErrors := validatePsychologistProfileInput(input, false); fieldErrors != nil {
		writeValidationError(c, fieldErrors)
		return
	}

	profile, err := h.appService.UpdatePsychologistProfile(c.Request.Context(), user.ID, input)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to update psychologist profile", nil)
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) GetPsychologistCard(c *gin.Context) {
	user := mustPsychologist(c)

	workspace, err := h.appService.GetPsychologistWorkspace(c.Request.Context(), user.ID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to load psychologist card", nil)
		return
	}

	c.JSON(http.StatusOK, workspace.Card)
}

func (h *Handler) UpdatePsychologistCard(c *gin.Context) {
	user := mustPsychologist(c)

	var input domain.UpdatePsychologistCardInput
	if !bindJSON(c, &input) {
		return
	}

	if fieldErrors := validatePsychologistCardInput(input, false); fieldErrors != nil {
		writeValidationError(c, fieldErrors)
		return
	}

	card, err := h.appService.UpdatePsychologistCard(c.Request.Context(), user.ID, input)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to update psychologist card", nil)
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *Handler) CreateSubscriptionPurchaseRequest(c *gin.Context) {
	user := mustPsychologist(c)

	var input domain.CreateSubscriptionPurchaseRequestInput
	if !bindJSON(c, &input) {
		return
	}

	request, err := h.appService.CreateSubscriptionPurchaseRequest(c.Request.Context(), user, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidSubscriptionPurchaseRequest):
			writeError(c, http.StatusBadRequest, "Validation failed", singleFieldError("subscription_plan", "Use basic or pro"))
		case errors.Is(err, service.ErrForbidden):
			writeError(c, http.StatusForbidden, "Access is allowed only for psychologists", nil)
		default:
			writeError(c, http.StatusInternalServerError, "Failed to create subscription purchase request", nil)
		}
		return
	}

	c.JSON(http.StatusCreated, request)
}

func (h *Handler) LogoutPsychologist(c *gin.Context) {
	sessionID, _ := c.Cookie(psychologistSessionCookieName)
	if err := h.appService.LogoutPsychologist(c.Request.Context(), sessionID); err != nil {
		writeError(c, http.StatusInternalServerError, "Failed to logout psychologist", nil)
		return
	}

	h.clearPsychologistSessionCookie(c)
	h.clearCSRFCookie(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func mustPsychologist(c *gin.Context) domain.AuthenticatedUser {
	value, _ := c.Get(authenticatedPsychologistKey)
	user, _ := value.(domain.AuthenticatedUser)
	return user
}

func (h *Handler) setPsychologistSessionCookie(c *gin.Context, sessionID string, expiresAt time.Time) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		psychologistSessionCookieName,
		sessionID,
		cookieMaxAge(expiresAt),
		"/",
		"",
		h.secureCookies,
		true,
	)
}

func (h *Handler) clearPsychologistSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(psychologistSessionCookieName, "", -1, "/", "", h.secureCookies, true)
}

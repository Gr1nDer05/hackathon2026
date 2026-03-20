package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Gr1nDer05/Hackathon2026/internal/domain"
	"github.com/Gr1nDer05/Hackathon2026/internal/service"
)

const authenticatedPsychologistKey = "authenticated_psychologist"
const psychologistSessionCookieName = "session_id"

func (h *Handler) LoginPsychologist(c *gin.Context) {
	clientIP := c.ClientIP()
	if !h.loginRateLimiter.Allow("psychologist:" + clientIP) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, try again later"})
		return
	}

	var input domain.PsychologistLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID, response, err := h.appService.LoginPsychologist(c.Request.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case errors.Is(err, service.ErrAccountDisabled):
			c.JSON(http.StatusForbidden, gin.H{"error": "psychologist access is disabled by administrator"})
		case errors.Is(err, service.ErrPortalAccessExpired):
			c.JSON(http.StatusForbidden, gin.H{"error": "portal access subscription has expired"})
		case errors.Is(err, service.ErrAccountTemporarilyBlocked):
			c.JSON(http.StatusForbidden, gin.H{"error": "psychologist account is temporarily blocked"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login psychologist"})
		}
		return
	}

	h.loginRateLimiter.Reset("psychologist:" + clientIP)
	h.setPsychologistSessionCookie(c, sessionID)
	if err := h.issueCSRFCookie(c); err != nil {
		h.clearPsychologistSessionCookie(c)
		h.clearCSRFCookie(c)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize session security"})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) RequirePsychologistAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(psychologistSessionCookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or expired session"})
			c.Abort()
			return
		}

		user, err := h.appService.AuthenticatePsychologist(c.Request.Context(), sessionID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrForbidden):
				c.JSON(http.StatusForbidden, gin.H{"error": "access is allowed only for psychologists"})
			case errors.Is(err, service.ErrAccountDisabled):
				h.clearPsychologistSessionCookie(c)
				h.clearCSRFCookie(c)
				c.JSON(http.StatusForbidden, gin.H{"error": "psychologist access is disabled by administrator"})
			case errors.Is(err, service.ErrPortalAccessExpired):
				h.clearPsychologistSessionCookie(c)
				h.clearCSRFCookie(c)
				c.JSON(http.StatusForbidden, gin.H{"error": "portal access subscription has expired"})
			case errors.Is(err, service.ErrAccountTemporarilyBlocked):
				h.clearPsychologistSessionCookie(c)
				h.clearCSRFCookie(c)
				c.JSON(http.StatusForbidden, gin.H{"error": "psychologist account is temporarily blocked"})
			default:
				h.clearPsychologistSessionCookie(c)
				h.clearCSRFCookie(c)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			}
			c.Abort()
			return
		}

		c.Set(authenticatedPsychologistKey, user)
		c.Next()
	}
}

func (h *Handler) GetPsychologistWorkspace(c *gin.Context) {
	user := mustPsychologist(c)

	workspace, err := h.appService.GetPsychologistWorkspace(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load psychologist workspace"})
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *Handler) GetPsychologistProfile(c *gin.Context) {
	user := mustPsychologist(c)

	workspace, err := h.appService.GetPsychologistWorkspace(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load psychologist profile"})
		return
	}

	c.JSON(http.StatusOK, workspace.Profile)
}

func (h *Handler) UpdatePsychologistProfile(c *gin.Context) {
	user := mustPsychologist(c)

	var input domain.UpdatePsychologistProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.appService.UpdatePsychologistProfile(c.Request.Context(), user.ID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update psychologist profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *Handler) GetPsychologistCard(c *gin.Context) {
	user := mustPsychologist(c)

	workspace, err := h.appService.GetPsychologistWorkspace(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load psychologist card"})
		return
	}

	c.JSON(http.StatusOK, workspace.Card)
}

func (h *Handler) UpdatePsychologistCard(c *gin.Context) {
	user := mustPsychologist(c)

	var input domain.UpdatePsychologistCardInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.appService.UpdatePsychologistCard(c.Request.Context(), user.ID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update psychologist card"})
		return
	}

	c.JSON(http.StatusOK, card)
}

func (h *Handler) LogoutPsychologist(c *gin.Context) {
	sessionID, _ := c.Cookie(psychologistSessionCookieName)
	if err := h.appService.LogoutPsychologist(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout psychologist"})
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

func (h *Handler) setPsychologistSessionCookie(c *gin.Context, sessionID string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		psychologistSessionCookieName,
		sessionID,
		int(service.SessionTTL.Seconds()),
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

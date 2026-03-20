package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/Gr1nDer05/Hackathon2026/internal/service"
	"github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

func (h *Handler) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := c.Writer.Header()
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Referrer-Policy", "no-referrer")
		headers.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		headers.Set("Cache-Control", "no-store")
		c.Next()
	}
}

func (h *Handler) RequireCSRFCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		if err := h.verifyOrigin(c); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "request origin is not allowed"})
			c.Abort()
			return
		}

		if err := h.verifyCSRFFromRequest(c); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid csrf token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *Handler) verifyCSRFFromRequest(c *gin.Context) error {
	csrfCookie, err := c.Cookie(csrfCookieName)
	if err != nil || strings.TrimSpace(csrfCookie) == "" {
		return errors.New("missing csrf cookie")
	}

	csrfHeader := strings.TrimSpace(c.GetHeader(csrfHeaderName))
	if csrfHeader == "" {
		return errors.New("missing csrf header")
	}

	if subtle.ConstantTimeCompare([]byte(csrfCookie), []byte(csrfHeader)) != 1 {
		return errors.New("csrf mismatch")
	}

	return nil
}

func (h *Handler) verifyOrigin(c *gin.Context) error {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" {
		return nil
	}

	if len(h.allowedOrigins) > 0 {
		if _, ok := h.allowedOrigins[origin]; ok {
			return nil
		}
		return errors.New("origin not in allow-list")
	}

	parsedOrigin, err := url.Parse(origin)
	if err != nil {
		return err
	}

	if !strings.EqualFold(parsedOrigin.Host, c.Request.Host) {
		return errors.New("origin host mismatch")
	}

	return nil
}

func (h *Handler) issueCSRFCookie(c *gin.Context) error {
	token, err := generateSecurityToken(32)
	if err != nil {
		return err
	}

	h.setCSRFCookie(c, token)
	return nil
}

func (h *Handler) setCSRFCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		csrfCookieName,
		token,
		int(service.SessionTTL.Seconds()),
		"/",
		"",
		h.secureCookies,
		false,
	)
}

func (h *Handler) clearCSRFCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(csrfCookieName, "", -1, "/", "", h.secureCookies, false)
}

func generateSecurityToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

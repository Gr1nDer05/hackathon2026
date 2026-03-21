package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"
const corsAllowHeaders = "Content-Type, X-CSRF-Token"
const corsAllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"

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

func (h *Handler) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin != "" && h.verifyOrigin(c) == nil {
			headers := c.Writer.Header()
			headers.Add("Vary", "Origin")
			headers.Add("Vary", "Access-Control-Request-Method")
			headers.Add("Vary", "Access-Control-Request-Headers")
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Credentials", "true")
			headers.Set("Access-Control-Allow-Headers", corsAllowHeaders)
			headers.Set("Access-Control-Allow-Methods", corsAllowMethods)
		}

		if c.Request.Method == http.MethodOptions {
			if origin != "" && h.verifyOrigin(c) != nil {
				abortWithError(c, http.StatusForbidden, "Request origin is not allowed", nil)
				return
			}

			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

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
			abortWithError(c, http.StatusForbidden, "Request origin is not allowed", nil)
			return
		}

		if err := h.verifyCSRFFromRequest(c); err != nil {
			abortWithError(c, http.StatusForbidden, "Invalid CSRF token", nil)
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

func (h *Handler) issueCSRFCookieWithExpiry(c *gin.Context, expiresAt time.Time) error {
	token, err := generateSecurityToken(32)
	if err != nil {
		return err
	}

	h.setCSRFCookie(c, token, cookieMaxAge(expiresAt))
	return nil
}

func (h *Handler) setCSRFCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		csrfCookieName,
		token,
		maxAge,
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

func cookieMaxAge(expiresAt time.Time) int {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		return 1
	}

	return maxAge
}

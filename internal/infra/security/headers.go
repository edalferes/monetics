// Package security provides HTTP security middlewares: secure response
// headers and a simple in-memory rate limiter (RN-S1, RN-S2).
package security

import (
	"github.com/labstack/echo/v4"
)

// SecureHeaders sets a conservative set of security-related response headers.
// Values follow OWASP Secure Headers Project recommendations.
func SecureHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			return next(c)
		}
	}
}

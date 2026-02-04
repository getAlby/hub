package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type ShutdownNotifier interface {
	IsShuttingDown() bool
}

func ShutdownMiddleware(notifier ShutdownNotifier) echo.MiddlewareFunc {
	safeRoutes := map[string]bool{
		"/api/health":      true,
		"/api/node/status": true,
		"/api/info":        true,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()

			if safeRoutes[path] || strings.HasPrefix(path, "/api/alby/") {
				return next(c)
			}

			if notifier.IsShuttingDown() {
				return c.JSON(http.StatusServiceUnavailable, map[string]string{
					"message": "Node is shutting down. Please wait.",
				})
			}
			return next(c)
		}
	}
}

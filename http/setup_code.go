package http

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/getAlby/hub/logger"
)

const setupCodeHeader = "X-Setup-Code"

// createSetupMiddlewares returns the middlewares for unauthenticated endpoints
// that modify the hub's configuration before setup is completed
func (httpSvc *HttpService) createSetupMiddlewares() []echo.MiddlewareFunc {
	// a single global bucket, like the unlock rate limiter
	setupRateLimiter := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 1, Burst: 3},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return "", nil
		},
	})
	return []echo.MiddlewareFunc{setupRateLimiter, httpSvc.requireSetupCode}
}

// requireSetupCode rejects requests that do not include the configured setup
// code until setup is completed. The check runs before the request body is read.
func (httpSvc *HttpService) requireSetupCode(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		setupCode := httpSvc.cfg.GetEnv().SetupCode
		if setupCode == "" {
			return next(c)
		}

		setupCompleted, err := httpSvc.cfg.SetupCompleted()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to check if setup is completed")
			return c.JSON(http.StatusInternalServerError, ErrorResponse{
				Message: "Failed to check if setup is completed",
			})
		}
		if setupCompleted {
			return next(c)
		}

		providedSetupCode := c.Request().Header.Get(setupCodeHeader)
		if subtle.ConstantTimeCompare([]byte(providedSetupCode), []byte(setupCode)) != 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{
				Message: "Invalid setup code",
			})
		}

		return next(c)
	}
}

// verifySetupCodeHandler lets the frontend check the setup code before the user
// continues with setup. The code is checked by requireSetupCode.
func (httpSvc *HttpService) verifySetupCodeHandler(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

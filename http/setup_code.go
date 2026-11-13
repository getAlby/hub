package http

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/getAlby/hub/logger"
)

const setupCodeHeader = "X-Setup-Code"

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

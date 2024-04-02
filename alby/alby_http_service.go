package alby

import (
	"fmt"
	"net/http"

	"github.com/getAlby/nostr-wallet-connect/models/api"
	models "github.com/getAlby/nostr-wallet-connect/models/http"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type AlbyHttpService struct {
	albyOAuthSvc *AlbyOAuthService
	logger       *logrus.Logger
}

func NewAlbyHttpService(albyOAuthSvc *AlbyOAuthService, logger *logrus.Logger) *AlbyHttpService {
	return &AlbyHttpService{
		albyOAuthSvc: albyOAuthSvc,
		logger:       logger,
	}
}

func (albyHttpSvc *AlbyHttpService) RegisterSharedRoutes(e *echo.Echo, authMiddleware func(next echo.HandlerFunc) echo.HandlerFunc) {
	e.GET("/api/alby/callback", albyHttpSvc.albyCallbackHandler, authMiddleware)
	e.GET("/api/alby/me", albyHttpSvc.albyMeHandler, authMiddleware)
	e.GET("/api/alby/balance", albyHttpSvc.albyBalanceHandler, authMiddleware)
	e.POST("/api/alby/pay", albyHttpSvc.albyPayHandler, authMiddleware)
}

func (albyHttpSvc *AlbyHttpService) albyCallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")

	err := albyHttpSvc.albyOAuthSvc.CallbackHandler(c.Request().Context(), code)
	if err != nil {
		albyHttpSvc.logger.WithError(err).Error("Failed to handle Alby OAuth callback")
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("Failed to handle Alby OAuth callback: %s", err.Error()),
		})
	}

	// FIXME: redirects will not work for wails
	return c.Redirect(302, albyHttpSvc.albyOAuthSvc.appConfig.BaseUrl)
}

func (albyHttpSvc *AlbyHttpService) albyMeHandler(c echo.Context) error {
	me, err := albyHttpSvc.albyOAuthSvc.GetMe(c.Request().Context())
	if err != nil {
		albyHttpSvc.logger.WithError(err).Error("Failed to request alby me endpoint")
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("Failed to request alby me endpoint: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, me)
}

func (albyHttpSvc *AlbyHttpService) albyBalanceHandler(c echo.Context) error {
	balance, err := albyHttpSvc.albyOAuthSvc.GetBalance(c.Request().Context())
	if err != nil {
		albyHttpSvc.logger.WithError(err).Error("Failed to request alby balance endpoint")
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("Failed to request alby balance endpoint: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, &api.AlbyBalanceResponse{
		Sats: balance.Balance,
	})
}

func (albyHttpSvc *AlbyHttpService) albyPayHandler(c echo.Context) error {
	var payRequest api.AlbyPayRequest
	if err := c.Bind(&payRequest); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := albyHttpSvc.albyOAuthSvc.SendPayment(c.Request().Context(), payRequest.Invoice)
	if err != nil {
		albyHttpSvc.logger.WithError(err).Error("Failed to request alby pay endpoint")
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("Failed to request alby pay endpoint: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

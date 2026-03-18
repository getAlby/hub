package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/labstack/echo/v4"
)

type AlbyHttpService struct {
	albySvc      alby.AlbyService
	albyOAuthSvc alby.AlbyOAuthService
	appConfig    *config.AppConfig
	svc          service.Service
}

func NewAlbyHttpService(svc service.Service, albySvc alby.AlbyService, albyOAuthSvc alby.AlbyOAuthService, appConfig *config.AppConfig) *AlbyHttpService {
	return &AlbyHttpService{
		albySvc:      albySvc,
		albyOAuthSvc: albyOAuthSvc,
		appConfig:    appConfig,
		svc:          svc,
	}
}

func (albyHttpSvc *AlbyHttpService) RegisterSharedRoutes(readOnlyApiGroup *echo.Group, fullAccessApiGroup *echo.Group, e *echo.Echo) {
	e.GET("/api/alby/callback", albyHttpSvc.albyCallbackHandler)
	e.GET("/api/alby/info", albyHttpSvc.albyInfoHandler)
	e.GET("/api/alby/rates", albyHttpSvc.albyBitcoinRateHandler)
	readOnlyApiGroup.GET("/alby/me", albyHttpSvc.albyMeHandler)
	fullAccessApiGroup.POST("/alby/link-account", albyHttpSvc.albyLinkAccountHandler)
	fullAccessApiGroup.POST("/alby/auto-channel", albyHttpSvc.autoChannelHandler)
	fullAccessApiGroup.POST("/alby/unlink-account", albyHttpSvc.unlinkHandler)
}

func (albyHttpSvc *AlbyHttpService) autoChannelHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var autoChannelRequest alby.AutoChannelRequest
	if err := c.Bind(&autoChannelRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	autoChannelResponseResponse, err := albyHttpSvc.albyOAuthSvc.RequestAutoChannel(ctx, albyHttpSvc.svc.GetLNClient(), autoChannelRequest.IsPublic)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request auto channel: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, autoChannelResponseResponse)
}

func (albyHttpSvc *AlbyHttpService) unlinkHandler(c echo.Context) error {
	ctx := c.Request().Context()

	err := albyHttpSvc.albyOAuthSvc.UnlinkAccount(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to unlink: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (albyHttpSvc *AlbyHttpService) albyInfoHandler(c echo.Context) error {
	info, err := albyHttpSvc.albySvc.GetInfo(c.Request().Context())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request alby info endpoint")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request alby info endpoint: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, info)
}

func (albyHttpSvc *AlbyHttpService) albyBitcoinRateHandler(c echo.Context) error {
	rate, err := albyHttpSvc.albySvc.GetBitcoinRate(c.Request().Context())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bitcoin rate")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to get Bitcoin rate: %s", err.Error()),
		})
	}
	return c.JSON(http.StatusOK, rate)
}

func (albyHttpSvc *AlbyHttpService) albyCallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")

	err := albyHttpSvc.albyOAuthSvc.CallbackHandler(c.Request().Context(), code)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to handle Alby OAuth callback")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to handle Alby OAuth callback: %s", err.Error()),
		})
	}

	if albyHttpSvc.appConfig.IsDefaultClientId() {
		// do not redirect if using default OAuth client
		// redirect will be handled by the frontend instead
		return c.NoContent(http.StatusNoContent)
	}

	redirectUrl := albyHttpSvc.appConfig.GetBaseFrontendUrl()

	if redirectUrl == "" {
		// OAuth using a custom client requires a base URL set for the callback
		return errors.New("no BASE_URL set")
	}

	return c.Redirect(http.StatusFound, redirectUrl)
}

func (albyHttpSvc *AlbyHttpService) albyMeHandler(c echo.Context) error {
	me, err := albyHttpSvc.albyOAuthSvc.GetMe(c.Request().Context())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request alby me endpoint")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request alby me endpoint: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, me)
}

func (albyHttpSvc *AlbyHttpService) albyLinkAccountHandler(c echo.Context) error {
	var linkAccountRequest alby.AlbyLinkAccountRequest
	if err := c.Bind(&linkAccountRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := albyHttpSvc.albyOAuthSvc.LinkAccount(c.Request().Context(), albyHttpSvc.svc.GetLNClient(), linkAccountRequest.Budget, linkAccountRequest.Renewal)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to connect alby account")
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

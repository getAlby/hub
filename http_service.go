package main

import (
	"errors"
	"fmt"
	"net/http"

	echologrus "github.com/davrux/echo-logrus/v4"
	"github.com/getAlby/nostr-wallet-connect/frontend"
	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

type HttpService struct {
	svc *Service
	api *API
}

func NewHttpService(svc *Service) *HttpService {
	return &HttpService{
		svc: svc,
		api: NewAPI(svc),
	}
}

func (httpSvc *HttpService) validateUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !httpSvc.isUnlocked(c) {
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}

func (httpSvc *HttpService) RegisterSharedRoutes(e *echo.Echo) {
	e.HideBanner = true
	e.Use(echologrus.Middleware())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "header:X-CSRF-Token",
	}))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(httpSvc.svc.cfg.CookieSecret))))

	authMiddleware := httpSvc.validateUserMiddleware
	e.GET("/api/apps", httpSvc.appsListHandler, authMiddleware)
	e.GET("/api/apps/:pubkey", httpSvc.appsShowHandler, authMiddleware)
	e.DELETE("/api/apps/:pubkey", httpSvc.appsDeleteHandler, authMiddleware)
	e.POST("/api/apps", httpSvc.appsCreateHandler, authMiddleware)

	e.GET("/api/csrf", httpSvc.csrfHandler)
	e.GET("/api/info", httpSvc.infoHandler)
	e.POST("/api/logout", httpSvc.logoutHandler)
	e.POST("/api/setup", httpSvc.setupHandler)

	// allow one unlock request per second
	unlockRateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1))
	e.POST("/api/start", httpSvc.startHandler, unlockRateLimiter)
	e.POST("/api/unlock", httpSvc.unlockHandler, unlockRateLimiter)

	// TODO: below could be supported by NIP-47
	e.GET("/api/channels", httpSvc.channelsListHandler, authMiddleware)
	e.POST("/api/channels", httpSvc.openChannelHandler, authMiddleware)
	// TODO: should this be DELETE /api/channels:id?
	e.POST("/api/channels/close", httpSvc.closeChannelHandler, authMiddleware)
	e.GET("/api/node/connection-info", httpSvc.nodeConnectionInfoHandler, authMiddleware)
	e.POST("/api/peers", httpSvc.connectPeerHandler, authMiddleware)
	e.POST("/api/wallet/new-address", httpSvc.newOnchainAddressHandler, authMiddleware)
	e.POST("/api/wallet/redeem-onchain-funds", httpSvc.redeemOnchainFundsHandler, authMiddleware)
	e.GET("/api/wallet/balance", httpSvc.onchainBalanceHandler, authMiddleware)

	e.GET("/api/mempool/lightning/nodes/:pubkey", httpSvc.mempoolLightningNodeHandler, authMiddleware)

	frontend.RegisterHandlers(e)
}

func (httpSvc *HttpService) csrfHandler(c echo.Context) error {
	csrf, _ := c.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
	if csrf == "" {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "CSRF token not available",
		})
	}
	return c.JSON(http.StatusOK, csrf)
}

func (httpSvc *HttpService) infoHandler(c echo.Context) error {
	responseBody := httpSvc.api.GetInfo()
	responseBody.Unlocked = httpSvc.isUnlocked(c)
	return c.JSON(http.StatusOK, responseBody)
}

func (httpSvc *HttpService) startHandler(c echo.Context) error {
	var startRequest api.StartRequest
	if err := c.Bind(&startRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := httpSvc.api.Start(&startRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to start node: %s", err.Error()),
		})
	}

	err = httpSvc.saveSessionCookie(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to save session: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) unlockHandler(c echo.Context) error {
	var unlockRequest api.UnlockRequest
	if err := c.Bind(&unlockRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	if !httpSvc.svc.cfg.CheckUnlockPassword(unlockRequest.UnlockPassword) {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: "Invalid password",
		})
	}

	err := httpSvc.saveSessionCookie(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to save session: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) isUnlocked(c echo.Context) bool {
	sess, _ := session.Get(SessionCookieName, c)
	return sess.Values[SessionCookieAuthKey] == true
}

func (httpSvc *HttpService) saveSessionCookie(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values[SessionCookieAuthKey] = true
	err := sess.Save(c.Request(), c.Response())
	if err != nil {
		httpSvc.svc.Logger.Errorf("Failed to save session: %v", err)
	}
	return err
}

func (httpSvc *HttpService) logoutHandler(c echo.Context) error {
	sess, err := session.Get(SessionCookieName, c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to get session",
		})
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to save session",
		})
	}
	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) channelsListHandler(c echo.Context) error {

	channels, err := httpSvc.api.ListChannels()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, channels)
}

func (httpSvc *HttpService) nodeConnectionInfoHandler(c echo.Context) error {

	info, err := httpSvc.api.GetNodeConnectionInfo()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, info)
}

func (httpSvc *HttpService) onchainBalanceHandler(c echo.Context) error {

	onchainBalanceResponse, err := httpSvc.api.GetOnchainBalance()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, onchainBalanceResponse)
}

func (httpSvc *HttpService) mempoolLightningNodeHandler(c echo.Context) error {
	pubkey := c.Param("pubkey")
	if pubkey == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid pubkey parameter",
		})
	}

	response, err := httpSvc.api.GetMempoolLightningNode(pubkey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request mempool API: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (httpSvc *HttpService) connectPeerHandler(c echo.Context) error {
	var connectPeerRequest api.ConnectPeerRequest
	if err := c.Bind(&connectPeerRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := httpSvc.api.ConnectPeer(&connectPeerRequest)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to connect peer: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) openChannelHandler(c echo.Context) error {
	var openChannelRequest api.OpenChannelRequest
	if err := c.Bind(&openChannelRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	openChannelResponse, err := httpSvc.api.OpenChannel(&openChannelRequest)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to open channel: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, openChannelResponse)
}

func (httpSvc *HttpService) closeChannelHandler(c echo.Context) error {
	var closeChannelRequest api.CloseChannelRequest
	if err := c.Bind(&closeChannelRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	closeChannelResponse, err := httpSvc.api.CloseChannel(&closeChannelRequest)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to close channel: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, closeChannelResponse)
}

func (httpSvc *HttpService) newOnchainAddressHandler(c echo.Context) error {
	newAddressResponse, err := httpSvc.api.GetNewOnchainAddress()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request new onchain address: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, newAddressResponse)
}

func (httpSvc *HttpService) redeemOnchainFundsHandler(c echo.Context) error {
	var redeemOnchainFundsRequest api.RedeemOnchainFundsRequest
	if err := c.Bind(&redeemOnchainFundsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	redeemOnchainFundsResponse, err := httpSvc.api.RedeemOnchainFunds(redeemOnchainFundsRequest.ToAddress)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request new onchain address: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, redeemOnchainFundsResponse)
}

func (httpSvc *HttpService) appsListHandler(c echo.Context) error {

	apps, err := httpSvc.api.ListApps()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, apps)
}

func (httpSvc *HttpService) appsShowHandler(c echo.Context) error {
	app := App{}
	findResult := httpSvc.svc.db.Where("nostr_pubkey = ?", c.Param("pubkey")).First(&app)

	if findResult.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "App does not exist",
		})
	}

	response := httpSvc.api.GetApp(&app)

	return c.JSON(http.StatusOK, response)
}

func (httpSvc *HttpService) appsDeleteHandler(c echo.Context) error {
	pubkey := c.Param("pubkey")
	if pubkey == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid pubkey parameter",
		})
	}
	app := App{}
	result := httpSvc.svc.db.Where("nostr_pubkey = ?", pubkey).First(&app)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Message: "App not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to fetch app",
		})
	}

	if err := httpSvc.api.DeleteApp(&app); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to delete app",
		})
	}
	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) appsCreateHandler(c echo.Context) error {
	var requestData api.CreateAppRequest
	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	responseBody, err := httpSvc.api.CreateApp(&requestData)

	if err != nil {
		httpSvc.svc.Logger.Errorf("Failed to save app: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to save app: %v", err),
		})
	}

	return c.JSON(http.StatusOK, responseBody)
}

func (httpSvc *HttpService) setupHandler(c echo.Context) error {
	var setupRequest api.SetupRequest
	if err := c.Bind(&setupRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	if httpSvc.svc.lnClient != nil && !httpSvc.isUnlocked(c) {
		return c.NoContent(http.StatusUnauthorized)
	}

	err := httpSvc.api.Setup(&setupRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to setup node: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

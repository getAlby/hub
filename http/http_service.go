package http

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/frontend"
)

type jwtCustomClaims struct {
	// we can add extra claims here
	// Name  string `json:"name"`
	// Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

type HttpService struct {
	api            api.API
	albyHttpSvc    *AlbyHttpService
	cfg            config.Config
	eventPublisher events.EventPublisher
	db             *gorm.DB
}

func NewHttpService(svc service.Service, eventPublisher events.EventPublisher) *HttpService {
	return &HttpService{
		api:            api.NewAPI(svc, svc.GetDB(), svc.GetConfig(), svc.GetKeys(), svc.GetAlbyOAuthSvc(), svc.GetEventPublisher()),
		albyHttpSvc:    NewAlbyHttpService(svc, svc.GetAlbyOAuthSvc(), svc.GetConfig().GetEnv()),
		cfg:            svc.GetConfig(),
		eventPublisher: eventPublisher,
		db:             svc.GetDB(),
	}
}

func (httpSvc *HttpService) RegisterSharedRoutes(e *echo.Echo) {
	e.HideBanner = true

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogHost:      true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			logger.Logger.WithFields(logrus.Fields{
				"uri":        values.URI,
				"status":     values.Status,
				"remote_ip":  values.RemoteIP,
				"user_agent": values.UserAgent,
				"host":       values.Host,
				"request_id": values.RequestID,
			}).Info("handled API request")
			return nil
		},
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.GET("/api/info", httpSvc.infoHandler)
	e.POST("/api/setup", httpSvc.setupHandler)
	e.POST("/api/restore", httpSvc.restoreBackupHandler)

	// allow one unlock request per second
	unlockRateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(1))
	e.POST("/api/start", httpSvc.startHandler, unlockRateLimiter)
	e.POST("/api/unlock", httpSvc.unlockHandler, unlockRateLimiter)
	e.PATCH("/api/unlock-password", httpSvc.changeUnlockPasswordHandler, unlockRateLimiter)

	frontend.RegisterHandlers(e)

	// restricted routes
	// Configure middleware with the custom claims type
	jwtConfig := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: []byte(httpSvc.cfg.GetJWTSecret()),
	}
	r := e.Group("")
	r.Use(echojwt.WithConfig(jwtConfig))

	r.GET("/api/apps", httpSvc.appsListHandler)
	r.GET("/api/apps/:pubkey", httpSvc.appsShowHandler)
	r.PATCH("/api/apps/:pubkey", httpSvc.appsUpdateHandler)
	r.DELETE("/api/apps/:pubkey", httpSvc.appsDeleteHandler)
	r.POST("/api/apps", httpSvc.appsCreateHandler)
	r.GET("/api/encrypted-mnemonic", httpSvc.encryptedMnemonicHandler)
	r.PATCH("/api/backup-reminder", httpSvc.backupReminderHandler)
	r.GET("/api/channels", httpSvc.channelsListHandler)
	r.POST("/api/channels", httpSvc.openChannelHandler)
	r.GET("/api/channels/suggestions", httpSvc.channelPeerSuggestionsHandler)
	r.POST("/api/lsp-orders", httpSvc.newInstantChannelInvoiceHandler)
	r.GET("/api/node/connection-info", httpSvc.nodeConnectionInfoHandler)
	r.GET("/api/node/status", httpSvc.nodeStatusHandler)
	r.GET("/api/node/network-graph", httpSvc.nodeNetworkGraphHandler)
	r.GET("/api/peers", httpSvc.listPeers)
	r.POST("/api/peers", httpSvc.connectPeerHandler)
	r.DELETE("/api/peers/:peerId", httpSvc.disconnectPeerHandler)
	r.DELETE("/api/peers/:peerId/channels/:channelId", httpSvc.closeChannelHandler)
	r.PATCH("/api/peers/:peerId/channels/:channelId", httpSvc.updateChannelHandler)
	r.GET("/api/wallet/address", httpSvc.onchainAddressHandler)
	r.POST("/api/wallet/new-address", httpSvc.newOnchainAddressHandler)
	r.POST("/api/wallet/redeem-onchain-funds", httpSvc.redeemOnchainFundsHandler)
	r.POST("/api/wallet/sign-message", httpSvc.signMessageHandler)
	r.POST("/api/wallet/sync", httpSvc.walletSyncHandler)
	r.GET("/api/wallet/capabilities", httpSvc.capabilitiesHandler)
	r.POST("/api/payments/:invoice", httpSvc.sendPaymentHandler)
	r.POST("/api/invoices", httpSvc.makeInvoiceHandler)
	r.GET("/api/transactions", httpSvc.listTransactionsHandler)
	r.GET("/api/transactions/:paymentHash", httpSvc.lookupTransactionHandler)
	r.GET("/api/balances", httpSvc.balancesHandler)
	r.POST("/api/reset-router", httpSvc.resetRouterHandler)
	r.POST("/api/stop", httpSvc.stopHandler)
	r.GET("/api/mempool", httpSvc.mempoolApiHandler)
	r.POST("/api/send-payment-probes", httpSvc.sendPaymentProbesHandler)
	r.POST("/api/send-spontaneous-payment-probes", httpSvc.sendSpontaneousPaymentProbesHandler)
	r.GET("/api/log/:type", httpSvc.getLogOutputHandler)
	r.POST("/api/backup", httpSvc.createBackupHandler)

	httpSvc.albyHttpSvc.RegisterSharedRoutes(r)
}

func (httpSvc *HttpService) infoHandler(c echo.Context) error {
	responseBody, err := httpSvc.api.GetInfo(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if parts[0] == "Bearer" {
			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(httpSvc.cfg.GetJWTSecret()), nil
			})
			if err != nil {
				logger.Logger.WithError(err).Error("failed to parse token")
			}
			responseBody.Unlocked = err == nil && token != nil && token.Valid
		}
	}

	return c.JSON(http.StatusOK, responseBody)
}

func (httpSvc *HttpService) encryptedMnemonicHandler(c echo.Context) error {
	responseBody := httpSvc.api.GetEncryptedMnemonic()
	return c.JSON(http.StatusOK, responseBody)
}

func (httpSvc *HttpService) backupReminderHandler(c echo.Context) error {
	var backupReminderRequest api.BackupReminderRequest
	if err := c.Bind(&backupReminderRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := httpSvc.api.SetNextBackupReminder(&backupReminderRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to store backup reminder: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) startHandler(c echo.Context) error {
	var startRequest api.StartRequest
	if err := c.Bind(&startRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	if !httpSvc.cfg.CheckUnlockPassword(startRequest.UnlockPassword) {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: "Invalid password",
		})
	}

	token, err := httpSvc.createJWT()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to save session: %s", err.Error()),
		})
	}

	go func() {
		err := httpSvc.api.Start(&startRequest)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to start node")
		}
	}()

	return c.JSON(http.StatusOK, token)
}

func (httpSvc *HttpService) unlockHandler(c echo.Context) error {
	var unlockRequest api.UnlockRequest
	if err := c.Bind(&unlockRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	if !httpSvc.cfg.CheckUnlockPassword(unlockRequest.UnlockPassword) {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: "Invalid password",
		})
	}

	token, err := httpSvc.createJWT()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to save session: %s", err.Error()),
		})
	}

	httpSvc.eventPublisher.Publish(&events.Event{
		Event: "nwc_unlocked",
	})

	return c.JSON(http.StatusOK, token)
}

func (httpSvc *HttpService) changeUnlockPasswordHandler(c echo.Context) error {
	var changeUnlockPasswordRequest api.ChangeUnlockPasswordRequest
	if err := c.Bind(&changeUnlockPasswordRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := httpSvc.api.ChangeUnlockPassword(&changeUnlockPasswordRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to change unlock password: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) createJWT() (string, error) {

	// Set custom claims
	claims := &jwtCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * time.Duration(httpSvc.cfg.GetEnv().JWTExpiryDays))),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if token == nil {
		return "", errors.New("failed to create token")
	}

	signed, err := token.SignedString([]byte(httpSvc.cfg.GetJWTSecret()))
	if err != nil {
		return "", err
	}
	return signed, nil
}

func (httpSvc *HttpService) channelsListHandler(c echo.Context) error {
	ctx := c.Request().Context()

	channels, err := httpSvc.api.ListChannels(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, channels)
}

func (httpSvc *HttpService) channelPeerSuggestionsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	suggestions, err := httpSvc.api.GetChannelPeerSuggestions(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, suggestions)
}

func (httpSvc *HttpService) resetRouterHandler(c echo.Context) error {
	var resetRouterRequest api.ResetRouterRequest
	if err := c.Bind(&resetRouterRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := httpSvc.api.ResetRouter(resetRouterRequest.Key)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) stopHandler(c echo.Context) error {

	err := httpSvc.api.Stop()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) nodeConnectionInfoHandler(c echo.Context) error {
	ctx := c.Request().Context()

	info, err := httpSvc.api.GetNodeConnectionInfo(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, info)
}

func (httpSvc *HttpService) nodeStatusHandler(c echo.Context) error {
	ctx := c.Request().Context()

	info, err := httpSvc.api.GetNodeStatus(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, info)
}

func (httpSvc *HttpService) nodeNetworkGraphHandler(c echo.Context) error {
	nodeIds := strings.Split(c.QueryParam("nodeIds"), ",")

	info, err := httpSvc.api.GetNetworkGraph(nodeIds)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, info)
}

func (httpSvc *HttpService) balancesHandler(c echo.Context) error {
	ctx := c.Request().Context()

	balances, err := httpSvc.api.GetBalances(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, balances)
}

func (httpSvc *HttpService) sendPaymentHandler(c echo.Context) error {
	ctx := c.Request().Context()

	paymentResponse, err := httpSvc.api.SendPayment(ctx, c.Param("invoice"))

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, paymentResponse)
}

func (httpSvc *HttpService) makeInvoiceHandler(c echo.Context) error {
	var makeInvoiceRequest api.MakeInvoiceRequest
	if err := c.Bind(&makeInvoiceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	invoice, err := httpSvc.api.CreateInvoice(c.Request().Context(), makeInvoiceRequest.Amount, makeInvoiceRequest.Description)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, invoice)
}

func (httpSvc *HttpService) lookupTransactionHandler(c echo.Context) error {
	ctx := c.Request().Context()

	transaction, err := httpSvc.api.LookupInvoice(ctx, c.Param("paymentHash"))

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, transaction)
}

func (httpSvc *HttpService) listTransactionsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	limit := uint64(20)
	offset := uint64(0)

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseUint(limitParam, 10, 64); err == nil {
			limit = parsedLimit
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.ParseUint(offsetParam, 10, 64); err == nil {
			offset = parsedOffset
		}
	}

	transactions, err := httpSvc.api.ListTransactions(ctx, limit, offset)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, transactions)
}

func (httpSvc *HttpService) walletSyncHandler(c echo.Context) error {
	httpSvc.api.SyncWallet()

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) mempoolApiHandler(c echo.Context) error {
	endpoint := c.QueryParam("endpoint")
	if endpoint == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid pubkey parameter",
		})
	}

	response, err := httpSvc.api.RequestMempoolApi(endpoint)
	if err != nil {
		logger.Logger.WithField("endpoint", endpoint).WithError(err).Error("Failed to request mempool API")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request mempool API: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (httpSvc *HttpService) capabilitiesHandler(c echo.Context) error {
	response, err := httpSvc.api.GetWalletCapabilities(c.Request().Context())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request wallet capabilities")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request wallet capabilities: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (httpSvc *HttpService) listPeers(c echo.Context) error {
	peers, err := httpSvc.api.ListPeers(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to list peers: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, peers)
}

func (httpSvc *HttpService) connectPeerHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var connectPeerRequest api.ConnectPeerRequest
	if err := c.Bind(&connectPeerRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := httpSvc.api.ConnectPeer(ctx, &connectPeerRequest)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to connect peer: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) openChannelHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var openChannelRequest api.OpenChannelRequest
	if err := c.Bind(&openChannelRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	openChannelResponse, err := httpSvc.api.OpenChannel(ctx, &openChannelRequest)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to open channel: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, openChannelResponse)
}

func (httpSvc *HttpService) disconnectPeerHandler(c echo.Context) error {
	ctx := c.Request().Context()

	err := httpSvc.api.DisconnectPeer(ctx, c.Param("peerId"))

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to disconnect peer: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) closeChannelHandler(c echo.Context) error {
	ctx := c.Request().Context()

	closeChannelResponse, err := httpSvc.api.CloseChannel(ctx, c.Param("peerId"), c.Param("channelId"), c.QueryParam("force") == "true")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to close channel: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, closeChannelResponse)
}

func (httpSvc *HttpService) updateChannelHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var updateChannelRequest api.UpdateChannelRequest
	if err := c.Bind(&updateChannelRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	updateChannelRequest.NodeId = c.Param("peerId")
	updateChannelRequest.ChannelId = c.Param("channelId")

	err := httpSvc.api.UpdateChannel(ctx, &updateChannelRequest)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to update channel: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) newInstantChannelInvoiceHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var newWrappedInvoiceRequest api.LSPOrderRequest
	if err := c.Bind(&newWrappedInvoiceRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	newWrappedInvoiceResponse, err := httpSvc.api.RequestLSPOrder(ctx, &newWrappedInvoiceRequest)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request wrapped invoice: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, newWrappedInvoiceResponse)
}

func (httpSvc *HttpService) onchainAddressHandler(c echo.Context) error {
	ctx := c.Request().Context()

	address, err := httpSvc.api.GetUnusedOnchainAddress(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request new onchain address: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, address)
}

func (httpSvc *HttpService) newOnchainAddressHandler(c echo.Context) error {
	ctx := c.Request().Context()

	address, err := httpSvc.api.GetNewOnchainAddress(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to request new onchain address: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, address)
}

func (httpSvc *HttpService) redeemOnchainFundsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var redeemOnchainFundsRequest api.RedeemOnchainFundsRequest
	if err := c.Bind(&redeemOnchainFundsRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	redeemOnchainFundsResponse, err := httpSvc.api.RedeemOnchainFunds(ctx, redeemOnchainFundsRequest.ToAddress)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to redeem onchain funds: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, redeemOnchainFundsResponse)
}

func (httpSvc *HttpService) signMessageHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var signMessageRequest api.SignMessageRequest
	if err := c.Bind(&signMessageRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	signMessageResponse, err := httpSvc.api.SignMessage(ctx, signMessageRequest.Message)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to sign messae: %s", err.Error()),
		})
	}
	return c.JSON(http.StatusOK, signMessageResponse)
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

	// TODO: move this to DB service
	dbApp := db.App{}
	findResult := httpSvc.db.Where("nostr_pubkey = ?", c.Param("pubkey")).First(&dbApp)

	if findResult.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "App does not exist",
		})
	}

	response := httpSvc.api.GetApp(&dbApp)

	return c.JSON(http.StatusOK, response)
}

func (httpSvc *HttpService) appsUpdateHandler(c echo.Context) error {
	var requestData api.UpdateAppRequest
	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	// TODO: move this to DB service
	dbApp := db.App{}
	findResult := httpSvc.db.Where("nostr_pubkey = ?", c.Param("pubkey")).First(&dbApp)

	if findResult.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "App does not exist",
		})
	}

	err := httpSvc.api.UpdateApp(&dbApp, &requestData)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to update app")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to update app: %v", err),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) appsDeleteHandler(c echo.Context) error {
	pubkey := c.Param("pubkey")
	if pubkey == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid pubkey parameter",
		})
	}
	// TODO: move this to DB service
	dbApp := db.App{}
	result := httpSvc.db.Where("nostr_pubkey = ?", pubkey).First(&dbApp)
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

	if err := httpSvc.api.DeleteApp(&dbApp); err != nil {
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
		logger.Logger.WithField("requestData", requestData).WithError(err).Error("Failed to save app")
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

	err := httpSvc.api.Setup(c.Request().Context(), &setupRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to setup node: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (httpSvc *HttpService) sendPaymentProbesHandler(c echo.Context) error {
	var sendPaymentProbesRequest api.SendPaymentProbesRequest
	if err := c.Bind(&sendPaymentProbesRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	sendPaymentProbesResponse, err := httpSvc.api.SendPaymentProbes(c.Request().Context(), &sendPaymentProbesRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to send payment probes: %v", err),
		})
	}

	return c.JSON(http.StatusOK, sendPaymentProbesResponse)
}

func (httpSvc *HttpService) sendSpontaneousPaymentProbesHandler(c echo.Context) error {
	var sendSpontaneousPaymentProbesRequest api.SendSpontaneousPaymentProbesRequest
	if err := c.Bind(&sendSpontaneousPaymentProbesRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	sendSpontaneousPaymentProbesResponse, err := httpSvc.api.SendSpontaneousPaymentProbes(c.Request().Context(), &sendSpontaneousPaymentProbesRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to send spontaneous payment probes: %v", err),
		})
	}

	return c.JSON(http.StatusOK, sendSpontaneousPaymentProbesResponse)
}

func (httpSvc *HttpService) getLogOutputHandler(c echo.Context) error {
	var getLogRequest api.GetLogOutputRequest
	if err := c.Bind(&getLogRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	logType := c.Param("type")
	if logType != api.LogTypeNode && logType != api.LogTypeApp {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Invalid log type parameter: '%s'", logType),
		})
	}

	getLogResponse, err := httpSvc.api.GetLogOutput(c.Request().Context(), logType, &getLogRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to get log output: %v", err),
		})
	}

	return c.JSON(http.StatusOK, getLogResponse)
}

func (httpSvc *HttpService) createBackupHandler(c echo.Context) error {
	var backupRequest api.BasicBackupRequest
	if err := c.Bind(&backupRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	if !httpSvc.cfg.CheckUnlockPassword(backupRequest.UnlockPassword) {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Message: "Invalid password",
		})
	}

	var buffer bytes.Buffer
	err := httpSvc.api.CreateBackup(backupRequest.UnlockPassword, &buffer)
	if err != nil {
		return c.String(500, fmt.Sprintf("Failed to create backup: %v", err))
	}

	c.Response().Header().Set("Content-Type", "application/octet-stream")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=nwc.bkp")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write(buffer.Bytes())
	return nil
}

func (httpSvc *HttpService) restoreBackupHandler(c echo.Context) error {
	info, err := httpSvc.api.GetInfo(c.Request().Context())
	if err != nil {
		return err
	}
	if info.SetupCompleted {
		return errors.New("Setup already completed")
	}

	password := c.FormValue("unlockPassword")

	fileHeader, err := c.FormFile("backup")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Failed to get backup file header: %v", err),
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to open backup file: %v", err),
		})
	}
	defer file.Close()

	err = httpSvc.api.RestoreBackup(password, file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to restore backup: %v", err),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

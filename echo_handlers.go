package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	echologrus "github.com/davrux/echo-logrus/v4"
	"github.com/getAlby/nostr-wallet-connect/frontend"
	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
	ddEcho "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo.v4"
	"gorm.io/gorm"
)

func (svc *Service) ValidateUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := svc.GetUser(c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: fmt.Sprintf("Bad arguments %s", err.Error()),
			})
		}
		if user == nil {
			return c.NoContent(http.StatusUnauthorized)
		}
		c.Set("user", user)
		return next(c)
	}
}

func (svc *Service) RegisterSharedRoutes(e *echo.Echo) {
	e.HideBanner = true
	e.Use(echologrus.Middleware())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "header:X-CSRF-Token",
	}))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(svc.cfg.CookieSecret))))
	e.Use(ddEcho.Middleware(ddEcho.WithServiceName("nostr-wallet-connect")))

	authMiddleware := svc.ValidateUserMiddleware
	e.GET("/api/user/me", svc.UserMeHandler, authMiddleware)
	e.GET("/api/apps", svc.AppsListHandler, authMiddleware)
	e.GET("/api/apps/:pubkey", svc.AppsShowHandler, authMiddleware)
	e.DELETE("/api/apps/:pubkey", svc.AppsDeleteHandler, authMiddleware)
	e.POST("/api/apps", svc.AppsCreateHandler, authMiddleware)

	e.GET("/api/csrf", svc.CSRFHandler)
	e.GET("/api/info", svc.InfoHandler)
	e.POST("/api/logout", svc.LogoutHandler)

	frontend.RegisterHandlers(e)
}

func (svc *Service) CSRFHandler(c echo.Context) error {
	csrf, _ := c.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
	if csrf == "" {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "CSRF token not available",
		})
	}
	return c.JSON(http.StatusOK, csrf)
}

func (svc *Service) InfoHandler(c echo.Context) error {
	responseBody := &api.InfoResponse{}
	responseBody.BackendType = svc.cfg.LNBackendType
	return c.JSON(http.StatusOK, responseBody)
}

func (svc *Service) LogoutHandler(c echo.Context) error {
	sess, err := session.Get(CookieName, c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to get session",
		})
	}
	sess.Options.MaxAge = -1
	if svc.cfg.CookieDomain != "" {
		sess.Options.Domain = svc.cfg.CookieDomain
	}
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to save session",
		})
	}
	return c.NoContent(http.StatusNoContent)
}

func (svc *Service) UserMeHandler(c echo.Context) error {
	user, _ := c.Get("user").(*User)
	responseBody := api.User{
		Email: user.Email,
	}
	return c.JSON(http.StatusOK, responseBody)
}

func (svc *Service) AppsListHandler(c echo.Context) error {
	user, _ := c.Get("user").(*User)
	userApps := user.Apps
	apps := []api.App{}

	err := svc.ListApps(&userApps, &apps)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, apps)
}

func (svc *Service) AppsShowHandler(c echo.Context) error {
	user, _ := c.Get("user").(*User)
	app := App{}
	findResult := svc.db.Where("user_id = ? AND nostr_pubkey = ?", user.ID, c.Param("pubkey")).First(&app)

	if findResult.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Message: "App does not exist",
		})
	}

	var lastEvent NostrEvent
	lastEventResult := svc.db.Where("app_id = ?", app.ID).Order("id desc").Limit(1).Find(&lastEvent)

	paySpecificPermission := AppPermission{}
	appPermissions := []AppPermission{}
	var expiresAt *time.Time
	svc.db.Where("app_id = ?", app.ID).Find(&appPermissions)

	requestMethods := []string{}
	for _, appPerm := range appPermissions {
		if !appPerm.ExpiresAt.IsZero() {
			expiresAt = &appPerm.ExpiresAt
		}
		if appPerm.RequestMethod == NIP_47_PAY_INVOICE_METHOD {
			//find the pay_invoice-specific permissions
			paySpecificPermission = appPerm
		}
		requestMethods = append(requestMethods, appPerm.RequestMethod)
	}

	//renewsIn := ""
	budgetUsage := int64(0)
	maxAmount := paySpecificPermission.MaxAmount
	if maxAmount > 0 {
		budgetUsage = svc.GetBudgetUsage(&paySpecificPermission)
	}

	response := api.App{
		Name:           app.Name,
		Description:    app.Description,
		CreatedAt:      app.CreatedAt,
		UpdatedAt:      app.UpdatedAt,
		NostrPubkey:    app.NostrPubkey,
		ExpiresAt:      expiresAt,
		MaxAmount:      maxAmount,
		RequestMethods: requestMethods,
		BudgetUsage:    budgetUsage,
		BudgetRenewal:  paySpecificPermission.BudgetRenewal,
	}

	if lastEventResult.RowsAffected > 0 {
		response.LastEventAt = &lastEvent.CreatedAt
	}

	return c.JSON(http.StatusOK, response)
}

func (svc *Service) AppsDeleteHandler(c echo.Context) error {
	user, _ := c.Get("user").(*User)
	pubkey := c.Param("pubkey")
	if pubkey == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid pubkey parameter",
		})
	}
	app := App{}
	result := svc.db.Where("user_id = ? AND nostr_pubkey = ?", user.ID, pubkey).First(&app)
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
	if err := svc.db.Delete(&app).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: "Failed to delete app",
		})
	}
	return c.NoContent(http.StatusNoContent)
}

func (svc *Service) AppsCreateHandler(c echo.Context) error {
	user, _ := c.Get("user").(*User)
	var requestData api.CreateAppRequest
	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	name := requestData.Name
	var pairingPublicKey string
	var pairingSecretKey string
	if requestData.Pubkey == "" {
		pairingSecretKey = nostr.GeneratePrivateKey()
		pairingPublicKey, _ = nostr.GetPublicKey(pairingSecretKey)
	} else {
		pairingPublicKey = requestData.Pubkey
		//validate public key
		decoded, err := hex.DecodeString(pairingPublicKey)
		if err != nil || len(decoded) != 32 {
			svc.Logger.Errorf("Invalid public key format: %s", pairingPublicKey)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: fmt.Sprintf("Invalid public key format: %s", pairingPublicKey),
			})
		}
	}

	// make sure there is not already a pubkey is already associated with an app
	// as apps are currently indexed by pubkey
	// TODO: shouldn't this be a database constraint?
	existingApp := App{}

	findResult := svc.db.Where("user_id = ? AND nostr_pubkey = ?", user.ID, pairingPublicKey).First(&existingApp)

	if findResult.RowsAffected > 0 {
		return c.JSON(http.StatusConflict, ErrorResponse{
			Message: "Pubkey already in use: " + existingApp.NostrPubkey,
		})
	}

	app := App{Name: name, NostrPubkey: pairingPublicKey}
	maxAmount := requestData.MaxAmount
	budgetRenewal := requestData.BudgetRenewal

	expiresAt := time.Time{}
	if requestData.ExpiresAt != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, requestData.ExpiresAt)
		if err != nil {
			svc.Logger.Errorf("Invalid expiresAt: %s", pairingPublicKey)
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Message: fmt.Sprintf("Invalid expiresAt: %v", err),
			})
		}
	}

	if !expiresAt.IsZero() {
		expiresAt = time.Date(expiresAt.Year(), expiresAt.Month(), expiresAt.Day(), 23, 59, 59, 0, expiresAt.Location())
	}

	err := svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&user).Association("Apps").Append(&app)
		if err != nil {
			return err
		}

		requestMethods := requestData.RequestMethods
		if requestMethods == "" {
			return fmt.Errorf("Won't create an app without request methods.")
		}
		//request methods should be space separated list of known request kinds
		methodsToCreate := strings.Split(requestMethods, " ")
		for _, m := range methodsToCreate {
			//if we don't know this method, we return an error
			if !strings.Contains(NIP_47_CAPABILITIES, m) {
				return fmt.Errorf("Did not recognize request method: %s", m)
			}
			appPermission := AppPermission{
				App:           app,
				RequestMethod: m,
				ExpiresAt:     expiresAt,
				//these fields are only relevant for pay_invoice
				MaxAmount:     maxAmount,
				BudgetRenewal: budgetRenewal,
			}
			err = tx.Create(&appPermission).Error
			if err != nil {
				return err
			}
		}
		// commit transaction
		return nil
	})

	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"pairingPublicKey": pairingPublicKey,
			"name":             name,
		}).Errorf("Failed to save app: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to save app: %v", err),
		})
	}

	publicRelayUrl := svc.cfg.PublicRelay
	if publicRelayUrl == "" {
		publicRelayUrl = svc.cfg.Relay
	}

	responseBody := &api.CreateAppResponse{}
	responseBody.Name = name
	responseBody.Pubkey = pairingPublicKey
	responseBody.PairingSecret = pairingSecretKey

	if requestData.ReturnTo != "" {
		returnToUrl, err := url.Parse(requestData.ReturnTo)
		if err == nil {
			query := returnToUrl.Query()
			query.Add("relay", publicRelayUrl)
			query.Add("pubkey", svc.cfg.IdentityPubkey)
			if user.LightningAddress != "" {
				query.Add("lud16", user.LightningAddress)
			}
			returnToUrl.RawQuery = query.Encode()
			responseBody.ReturnTo = returnToUrl.String()
		}
	}

	var lud16 string
	if user.LightningAddress != "" {
		lud16 = fmt.Sprintf("&lud16=%s", user.LightningAddress)
	}
	responseBody.PairingUri = fmt.Sprintf("nostr+walletconnect://%s?relay=%s&secret=%s%s", svc.cfg.IdentityPubkey, publicRelayUrl, pairingSecretKey, lud16)
	return c.JSON(http.StatusOK, responseBody)
}

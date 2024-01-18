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

// TODO: echo methods should not be on Service object

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

	authMiddleware := svc.ValidateUserMiddleware
	e.GET("/api/user/me", svc.UserMeHandler, authMiddleware)
	e.GET("/api/apps", svc.AppsListHandler, authMiddleware)
	e.GET("/api/apps/:pubkey", svc.AppsShowHandler, authMiddleware)
	e.DELETE("/api/apps/:pubkey", svc.AppsDeleteHandler, authMiddleware)
	e.POST("/api/apps", svc.AppsCreateHandler, authMiddleware)

	e.GET("/api/csrf", svc.CSRFHandler)
	e.GET("/api/info", svc.InfoHandler)
	e.POST("/api/logout", svc.LogoutHandler)
	e.POST("/api/setup", svc.SetupHandler)

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
	responseBody := svc.GetInfo()
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

	apps, err := svc.ListApps(&userApps)

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

	response := svc.GetApp(&app)

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

	if err := svc.DeleteApp(&app); err != nil {
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

	responseBody, err := svc.CreateApp(user, &requestData)

	if err != nil {
		svc.Logger.Errorf("Failed to save app: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to save app: %v", err),
		})
	}

	return c.JSON(http.StatusOK, responseBody)
}

func (svc *Service) SetupHandler(c echo.Context) error {
	var setupRequest api.SetupRequest
	if err := c.Bind(&setupRequest); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: fmt.Sprintf("Bad request: %s", err.Error()),
		})
	}

	err := svc.Setup(&setupRequest)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to setup node: %s", err.Error()),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

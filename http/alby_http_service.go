package http

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
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
	e.GET("/api/alby/auth-url", albyHttpSvc.albyAuthUrlHandler)
	e.GET("/api/alby/callback", albyHttpSvc.albyCallbackHandler)
	e.GET("/api/alby/info", albyHttpSvc.albyInfoHandler)
	e.GET("/api/alby/rates/:currency", albyHttpSvc.albyBitcoinRateHandler)
	e.GET("/api/alby/currencies", albyHttpSvc.albyCurrenciesHandler)
	e.GET("/api/alby/stories", albyHttpSvc.albyStoriesHandler)
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
	rate, err := albyHttpSvc.albySvc.GetBitcoinRate(c.Request().Context(), c.Param("currency"))
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bitcoin rate")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to get Bitcoin rate: %s", err.Error()),
		})
	}
	return c.JSON(http.StatusOK, rate)
}

func (albyHttpSvc *AlbyHttpService) albyCurrenciesHandler(c echo.Context) error {
	currencies, err := albyHttpSvc.albySvc.GetCurrencies(c.Request().Context())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get currencies")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to get currencies: %s", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, currencies)
}

func (albyHttpSvc *AlbyHttpService) albyStoriesHandler(c echo.Context) error {
	stories, err := albyHttpSvc.albyOAuthSvc.GetStories(c.Request().Context())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get stories")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to get stories: %s", err.Error()),
		})
	}
	return c.JSON(http.StatusOK, stories)
}

const (
	albyOAuthStateCookieName    = "alby_oauth_state"
	albyOAuthVerifierCookieName = "alby_oauth_verifier"
	albyOAuthCookieMaxAge       = 60 * 60
	albyOAuthVerifierBytes      = 32
)

func (albyHttpSvc *AlbyHttpService) albyAuthUrlHandler(c echo.Context) error {
	state, verifier, err := resolveOAuthSession(c)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to generate OAuth state")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to generate OAuth state: %s", err.Error()),
		})
	}

	setOAuthCookie(c, albyOAuthStateCookieName, state, albyOAuthCookieMaxAge)
	setOAuthCookie(c, albyOAuthVerifierCookieName, verifier, albyOAuthCookieMaxAge)

	return c.JSON(http.StatusOK, &alby.AlbyAuthUrlResponse{
		Url: albyHttpSvc.albyOAuthSvc.GetAuthUrl(state, verifier),
	})
}

func (albyHttpSvc *AlbyHttpService) albyCallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")

	verifier, ok := checkOAuthSession(c)
	if !ok {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: "Invalid OAuth state. Please try connecting your Alby Account again.",
		})
	}

	err := albyHttpSvc.albyOAuthSvc.CallbackHandler(c.Request().Context(), code, verifier)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to handle Alby OAuth callback")
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Message: fmt.Sprintf("Failed to handle Alby OAuth callback: %s", err.Error()),
		})
	}

	setOAuthCookie(c, albyOAuthStateCookieName, "", -1)
	setOAuthCookie(c, albyOAuthVerifierCookieName, "", -1)

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

// resolveOAuthSession reuses the state and verifier from an earlier auth URL
// request so that refetching the auth URL does not invalidate a URL that was
// already opened
func resolveOAuthSession(c echo.Context) (string, string, error) {
	stateCookie, stateErr := c.Cookie(albyOAuthStateCookieName)
	verifierCookie, verifierErr := c.Cookie(albyOAuthVerifierCookieName)
	if stateErr == nil && verifierErr == nil &&
		isValidOAuthState(stateCookie.Value) && isValidOAuthVerifier(verifierCookie.Value) {
		return stateCookie.Value, verifierCookie.Value, nil
	}

	state, err := alby.GenerateOAuthState()
	if err != nil {
		return "", "", err
	}
	return state, oauth2.GenerateVerifier(), nil
}

func setOAuthCookie(c echo.Context, name string, value string, maxAge int) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/api/alby",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   c.Scheme() == "https",
		SameSite: http.SameSiteLaxMode,
	})
}

// checkOAuthSession returns the verifier for the token exchange if the
// callback state matches the state issued to this browser
func checkOAuthSession(c echo.Context) (string, bool) {
	state := c.QueryParam("state")
	if !isValidOAuthState(state) {
		return "", false
	}

	stateCookie, err := c.Cookie(albyOAuthStateCookieName)
	if err != nil {
		return "", false
	}
	if subtle.ConstantTimeCompare([]byte(state), []byte(stateCookie.Value)) != 1 {
		return "", false
	}

	verifierCookie, err := c.Cookie(albyOAuthVerifierCookieName)
	if err != nil || !isValidOAuthVerifier(verifierCookie.Value) {
		return "", false
	}

	return verifierCookie.Value, true
}

func isValidOAuthState(state string) bool {
	decoded, err := hex.DecodeString(state)
	return err == nil && len(decoded) == alby.OAuthStateBytes
}

func isValidOAuthVerifier(verifier string) bool {
	decoded, err := base64.RawURLEncoding.DecodeString(verifier)
	return err == nil && len(decoded) == albyOAuthVerifierBytes
}

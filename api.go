package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	models "github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"
)

type API struct {
	svc *Service
}

func NewAPI(svc *Service) *API {
	return &API{
		svc: svc,
	}
}

func (api *API) CreateApp(createAppRequest *models.CreateAppRequest) (*models.CreateAppResponse, error) {
	name := createAppRequest.Name
	var pairingPublicKey string
	var pairingSecretKey string
	if createAppRequest.Pubkey == "" {
		pairingSecretKey = nostr.GeneratePrivateKey()
		pairingPublicKey, _ = nostr.GetPublicKey(pairingSecretKey)
	} else {
		pairingPublicKey = createAppRequest.Pubkey
		//validate public key
		decoded, err := hex.DecodeString(pairingPublicKey)
		if err != nil || len(decoded) != 32 {
			api.svc.Logger.Errorf("Invalid public key format: %s", pairingPublicKey)
			return nil, errors.New(fmt.Sprintf("Invalid public key format: %s", pairingPublicKey))

		}
	}

	app := App{Name: name, NostrPubkey: pairingPublicKey}
	maxAmount := createAppRequest.MaxAmount
	budgetRenewal := createAppRequest.BudgetRenewal

	expiresAt := time.Time{}
	if createAppRequest.ExpiresAt != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, createAppRequest.ExpiresAt)
		if err != nil {
			api.svc.Logger.Errorf("Invalid expiresAt: %s", pairingPublicKey)
			return nil, errors.New(fmt.Sprintf("Invalid expiresAt: %v", err))
		}
	}

	if !expiresAt.IsZero() {
		expiresAt = time.Date(expiresAt.Year(), expiresAt.Month(), expiresAt.Day(), 23, 59, 59, 0, expiresAt.Location())
	}

	err := api.svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&app).Error
		if err != nil {
			return err
		}

		requestMethods := createAppRequest.RequestMethods
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
		return nil, err
	}

	relayUrl, _ := api.svc.cfg.Get("Relay", "")

	responseBody := &models.CreateAppResponse{}
	responseBody.Name = name
	responseBody.Pubkey = pairingPublicKey
	responseBody.PairingSecret = pairingSecretKey

	if createAppRequest.ReturnTo != "" {
		returnToUrl, err := url.Parse(createAppRequest.ReturnTo)
		if err == nil {
			query := returnToUrl.Query()
			query.Add("relay", relayUrl)
			query.Add("pubkey", api.svc.cfg.NostrPublicKey)
			// if user.LightningAddress != "" {
			// 	query.Add("lud16", user.LightningAddress)
			// }
			returnToUrl.RawQuery = query.Encode()
			responseBody.ReturnTo = returnToUrl.String()
		}
	}

	var lud16 string
	// if user.LightningAddress != "" {
	// 	lud16 = fmt.Sprintf("&lud16=%s", user.LightningAddress)
	// }
	responseBody.PairingUri = fmt.Sprintf("nostr+walletconnect://%s?relay=%s&secret=%s%s", api.svc.cfg.NostrPublicKey, relayUrl, pairingSecretKey, lud16)
	return responseBody, nil
}

func (api *API) DeleteApp(userApp *App) error {
	return api.svc.db.Delete(userApp).Error
}

func (api *API) GetApp(userApp *App) *models.App {

	var lastEvent RequestEvent
	lastEventResult := api.svc.db.Where("app_id = ?", userApp.ID).Order("id desc").Limit(1).Find(&lastEvent)

	paySpecificPermission := AppPermission{}
	appPermissions := []AppPermission{}
	var expiresAt *time.Time
	api.svc.db.Where("app_id = ?", userApp.ID).Find(&appPermissions)

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
		budgetUsage = api.svc.GetBudgetUsage(&paySpecificPermission)
	}

	response := models.App{
		Name:           userApp.Name,
		Description:    userApp.Description,
		CreatedAt:      userApp.CreatedAt,
		UpdatedAt:      userApp.UpdatedAt,
		NostrPubkey:    userApp.NostrPubkey,
		ExpiresAt:      expiresAt,
		MaxAmount:      maxAmount,
		RequestMethods: requestMethods,
		BudgetUsage:    budgetUsage,
		BudgetRenewal:  paySpecificPermission.BudgetRenewal,
	}

	if lastEventResult.RowsAffected > 0 {
		response.LastEventAt = &lastEvent.CreatedAt
	}

	return &response

}

func (api *API) ListApps() ([]models.App, error) {
	apps := []App{}
	api.svc.db.Find(&apps)

	apiApps := []models.App{}
	for _, userApp := range apps {
		apiApp := models.App{
			// ID:          app.ID,
			Name:        userApp.Name,
			Description: userApp.Description,
			CreatedAt:   userApp.CreatedAt,
			UpdatedAt:   userApp.UpdatedAt,
			NostrPubkey: userApp.NostrPubkey,
		}

		var lastEvent RequestEvent
		result := api.svc.db.Where("app_id = ?", userApp.ID).Order("id desc").Limit(1).Find(&lastEvent)
		if result.Error != nil {
			api.svc.Logger.Errorf("Failed to fetch last event %v", result.Error)
			return nil, errors.New("Failed to fetch last event")
		}
		if result.RowsAffected > 0 {
			apiApp.LastEventAt = &lastEvent.CreatedAt
		}
		apiApps = append(apiApps, apiApp)
	}
	return apiApps, nil
}

func (api *API) GetInfo() *models.InfoResponse {
	info := models.InfoResponse{}
	backendType, _ := api.svc.cfg.Get("LNBackendType", "")
	unlockPasswordCheck, _ := api.svc.cfg.Get("UnlockPasswordCheck", "")
	info.SetupCompleted = unlockPasswordCheck != ""
	info.Running = api.svc.lnClient != nil
	info.BackendType = backendType
	return &info
}

func (api *API) Start(startRequest *models.StartRequest) error {
	return api.svc.StartApp(startRequest.UnlockPassword)
}

func (api *API) Setup(setupRequest *models.SetupRequest) error {
	api.svc.cfg.SavePasswordCheck(setupRequest.UnlockPassword)

	// only update non-empty values
	if setupRequest.LNBackendType != "" {
		api.svc.cfg.SetUpdate("LNBackendType", setupRequest.LNBackendType, "")
	}
	if setupRequest.BreezAPIKey != "" {
		api.svc.cfg.SetUpdate("BreezAPIKey", setupRequest.BreezAPIKey, setupRequest.UnlockPassword)
	}
	if setupRequest.BreezMnemonic != "" {
		api.svc.cfg.SetUpdate("BreezMnemonic", setupRequest.BreezMnemonic, setupRequest.UnlockPassword)
	}
	if setupRequest.GreenlightInviteCode != "" {
		api.svc.cfg.SetUpdate("GreenlightInviteCode", setupRequest.GreenlightInviteCode, setupRequest.UnlockPassword)
	}
	if setupRequest.LNDAddress != "" {
		api.svc.cfg.SetUpdate("LNDAddress", setupRequest.LNDAddress, setupRequest.UnlockPassword)
	}
	if setupRequest.LNDCertHex != "" {
		api.svc.cfg.SetUpdate("LNDCertHex", setupRequest.LNDCertHex, setupRequest.UnlockPassword)
	}
	if setupRequest.LNDMacaroonHex != "" {
		api.svc.cfg.SetUpdate("LNDMacaroonHex", setupRequest.LNDMacaroonHex, setupRequest.UnlockPassword)
	}

	return nil
}

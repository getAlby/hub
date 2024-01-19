package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/getAlby/nostr-wallet-connect/models/db"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"
)

// TODO: these methods should be moved to a separate object, not in Service

// TODO: remove user parameter in single-user fork
func (svc *Service) CreateApp(user *User, createAppRequest *api.CreateAppRequest) (*api.CreateAppResponse, error) {
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
			svc.Logger.Errorf("Invalid public key format: %s", pairingPublicKey)
			return nil, errors.New(fmt.Sprintf("Invalid public key format: %s", pairingPublicKey))

		}
	}

	// make sure there is not already a pubkey is already associated with an app
	// as apps are currently indexed by pubkey
	// TODO: shouldn't this be a database constraint?
	existingApp := App{}
	findResult := svc.db.Where("user_id = ? AND nostr_pubkey = ?", user.ID, pairingPublicKey).First(&existingApp)
	if findResult.RowsAffected > 0 {
		return nil, errors.New("Pubkey already in use: " + existingApp.NostrPubkey)
	}

	app := App{Name: name, NostrPubkey: pairingPublicKey}
	maxAmount := createAppRequest.MaxAmount
	budgetRenewal := createAppRequest.BudgetRenewal

	expiresAt := time.Time{}
	if createAppRequest.ExpiresAt != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, createAppRequest.ExpiresAt)
		if err != nil {
			svc.Logger.Errorf("Invalid expiresAt: %s", pairingPublicKey)
			return nil, errors.New(fmt.Sprintf("Invalid expiresAt: %v", err))
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

	publicRelayUrl := svc.cfg.PublicRelay
	if publicRelayUrl == "" {
		publicRelayUrl = svc.cfg.Relay
	}

	responseBody := &api.CreateAppResponse{}
	responseBody.Name = name
	responseBody.Pubkey = pairingPublicKey
	responseBody.PairingSecret = pairingSecretKey

	if createAppRequest.ReturnTo != "" {
		returnToUrl, err := url.Parse(createAppRequest.ReturnTo)
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
	return responseBody, nil
}

func (svc *Service) DeleteApp(userApp *App) error {
	return svc.db.Delete(userApp).Error
}

func (svc *Service) GetApp(userApp *App) *api.App {

	var lastEvent NostrEvent
	lastEventResult := svc.db.Where("app_id = ?", userApp.ID).Order("id desc").Limit(1).Find(&lastEvent)

	paySpecificPermission := AppPermission{}
	appPermissions := []AppPermission{}
	var expiresAt *time.Time
	svc.db.Where("app_id = ?", userApp.ID).Find(&appPermissions)

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

func (svc *Service) ListApps(userApps *[]App) ([]api.App, error) {
	apps := []api.App{}
	for _, userApp := range *userApps {
		apiApp := api.App{
			// ID:          app.ID,
			Name:        userApp.Name,
			Description: userApp.Description,
			CreatedAt:   userApp.CreatedAt,
			UpdatedAt:   userApp.UpdatedAt,
			NostrPubkey: userApp.NostrPubkey,
		}

		var lastEvent NostrEvent
		result := svc.db.Where("app_id = ?", userApp.ID).Order("id desc").Limit(1).Find(&lastEvent)
		if result.Error != nil {
			svc.Logger.Errorf("Failed to fetch last event %v", result.Error)
			return nil, errors.New("Failed to fetch last event")
		}
		if result.RowsAffected > 0 {
			apiApp.LastEventAt = &lastEvent.CreatedAt
		}
		apps = append(apps, apiApp)
	}
	return apps, nil
}

func (svc *Service) GetInfo() *api.InfoResponse {
	info := api.InfoResponse{}
	info.BackendType = svc.cfg.LNBackendType
	info.SetupCompleted = svc.lnClient != nil
	return &info
}

func (svc *Service) Setup(setupRequest *api.SetupRequest) error {

	dbConfig := db.Config{}
	err := svc.db.First(&dbConfig).Error

	if err != nil {
		svc.Logger.Errorf("Failed to get db config: %v", err)
		return err
	}

	// only update non-empty values
	if setupRequest.LNBackendType != "" {
		dbConfig.LNBackendType = setupRequest.LNBackendType
	}
	if setupRequest.BreezAPIKey != "" {
		dbConfig.BreezAPIKey = setupRequest.BreezAPIKey
	}
	if setupRequest.BreezMnemonic != "" {
		dbConfig.BreezMnemonic = setupRequest.BreezMnemonic
	}
	if setupRequest.GreenlightInviteCode != "" {
		dbConfig.GreenlightInviteCode = setupRequest.GreenlightInviteCode
	}
	if setupRequest.LNDAddress != "" {
		dbConfig.LNDAddress = setupRequest.LNDAddress
	}
	if setupRequest.LNDCertHex != "" {
		dbConfig.LNDCertHex = setupRequest.LNDCertHex
	}
	if setupRequest.LNDMacaroonHex != "" {
		dbConfig.LNDMacaroonHex = setupRequest.LNDMacaroonHex
	}

	err = svc.db.Save(&dbConfig).Error

	if err != nil {
		svc.Logger.Errorf("Failed to update config: %v", err)
		return err
	}

	return svc.launchLNBackend()
}

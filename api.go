package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	models "github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
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
			return nil, fmt.Errorf("invalid public key format: %s", pairingPublicKey)

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
			return nil, fmt.Errorf("invalid expiresAt: %v", err)
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
			return fmt.Errorf("won't create an app without request methods")
		}
		//request methods should be space separated list of known request kinds
		methodsToCreate := strings.Split(requestMethods, " ")
		for _, m := range methodsToCreate {
			//if we don't know this method, we return an error
			if !strings.Contains(NIP_47_CAPABILITIES, m) {
				return fmt.Errorf("did not recognize request method: %s", m)
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

func (api *API) UpdateApp(userApp *App, updateAppRequest *models.UpdateAppRequest) error {
	maxAmount := updateAppRequest.MaxAmount
	budgetRenewal := updateAppRequest.BudgetRenewal

	expiresAt := time.Time{}
	if updateAppRequest.ExpiresAt != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, updateAppRequest.ExpiresAt)
		if err != nil {
			return fmt.Errorf("invalid expiresAt: %v", err)
		}
	}

	if !expiresAt.IsZero() {
		expiresAt = time.Date(expiresAt.Year(), expiresAt.Month(), expiresAt.Day(), 23, 59, 59, 0, expiresAt.Location())
	}

	err := api.svc.db.Transaction(func(tx *gorm.DB) error {
		var existingPermissionsCount int64
		tx.Model(&AppPermission{}).Where("app_id = ?", userApp.ID).Count(&existingPermissionsCount)

		var requestMethodsToAdd []string
		var requestMethodsToRemove []string
		if updateAppRequest.RequestMethodsToAdd != "" {
			requestMethodsToAdd = strings.Split(updateAppRequest.RequestMethodsToAdd, " ")
		}
		if updateAppRequest.RequestMethodsToRemove != "" {
			requestMethodsToRemove = strings.Split(updateAppRequest.RequestMethodsToRemove, " ")
		}

		resultingPermissionsCount := existingPermissionsCount + int64(len(requestMethodsToAdd)) - int64(len(requestMethodsToRemove))

		if resultingPermissionsCount <= 0 {
			return fmt.Errorf("won't update an app without request methods")
		}

		for _, m := range requestMethodsToAdd {
			//if we don't know this method, we return an error
			if !strings.Contains(NIP_47_CAPABILITIES, m) {
				return fmt.Errorf("did not recognize request method: %s", m)
			}
			appPermission := AppPermission{
				App:           *userApp,
				RequestMethod: m,
				ExpiresAt:     expiresAt,
				//these fields are only relevant for pay_invoice
				MaxAmount:     maxAmount,
				BudgetRenewal: budgetRenewal,
			}
			err := tx.Create(&appPermission).Error
			if err != nil {
				return err
			}
		}
		for _, m := range requestMethodsToRemove {
			//if we don't know this method, we return an error
			if !strings.Contains(NIP_47_CAPABILITIES, m) {
				return fmt.Errorf("did not recognize request method: %s", m)
			}
			// Find the app permission to delete
			var appPermission AppPermission
			err := tx.Where("app_id = ? AND request_method = ?", userApp.ID, m).First(&appPermission).Error
			if err != nil {
				return err
			}
			// Delete the app permission
			err = tx.Delete(&appPermission).Error
			if err != nil {
				return err
			}
		}

		// Update the budget ifno in pay_invoice permission
		// The other permissions will remain same
		paySpecificPermission := AppPermission{}
		findPermissionResult := tx.Find(&paySpecificPermission, &AppPermission{
			AppId:         userApp.ID,
			RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		})
		if findPermissionResult.RowsAffected != 0 {
			paySpecificPermission.MaxAmount = maxAmount
			paySpecificPermission.BudgetRenewal = budgetRenewal
			tx.Save(&paySpecificPermission)
		}

		// commit transaction
		return nil
	})

	return err
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

		permissions := []AppPermission{}
		result := api.svc.db.Where("app_id = ?", userApp.ID).Find(&permissions)
		if result.Error != nil {
			api.svc.Logger.Errorf("Failed to fetch app permissions %v", result.Error)
			return nil, errors.New("failed to fetch app permissions")
		}

		for _, permission := range permissions {
			apiApp.RequestMethods = append(apiApp.RequestMethods, permission.RequestMethod)
			if permission.RequestMethod == NIP_47_PAY_INVOICE_METHOD {
				apiApp.MaxAmount = permission.MaxAmount
				if apiApp.MaxAmount > 0 {
					apiApp.BudgetUsage = api.svc.GetBudgetUsage(&permission)
				}
			}
		}

		apiApps = append(apiApps, apiApp)
	}
	return apiApps, nil
}

func (api *API) ListChannels() ([]lnclient.Channel, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.ListChannels(api.svc.ctx)
}

func (api *API) GetNodeConnectionInfo() (*lnclient.NodeConnectionInfo, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.GetNodeConnectionInfo(api.svc.ctx)
}

func (api *API) ConnectPeer(connectPeerRequest *models.ConnectPeerRequest) error {
	if api.svc.lnClient == nil {
		return errors.New("LNClient not started")
	}
	return api.svc.lnClient.ConnectPeer(api.svc.ctx, connectPeerRequest)
}

func (api *API) OpenChannel(openChannelRequest *models.OpenChannelRequest) (*models.OpenChannelResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.OpenChannel(api.svc.ctx, openChannelRequest)
}

func (api *API) CloseChannel(closeChannelRequest *models.CloseChannelRequest) (*models.CloseChannelResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.CloseChannel(api.svc.ctx, closeChannelRequest)
}

func (api *API) GetNewOnchainAddress() (*models.NewOnchainAddressResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	address, err := api.svc.lnClient.GetNewOnchainAddress(api.svc.ctx)
	if err != nil {
		return nil, err
	}
	return &models.NewOnchainAddressResponse{
		Address: address,
	}, nil
}

func (api *API) GetOnchainBalance() (*models.OnchainBalanceResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	balance, err := api.svc.lnClient.GetOnchainBalance(api.svc.ctx)
	if err != nil {
		return nil, err
	}
	return &models.OnchainBalanceResponse{
		Sats: balance,
	}, nil
}

func (api *API) GetMempoolLightningNode(pubkey string) (interface{}, error) {
	url := "https://mempool.space/api/v1/lightning/nodes/" + pubkey

	spaceClient := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		api.svc.Logger.Errorf("Failed to create http request %s %v", url, err)
		return nil, err
	}

	res, err := spaceClient.Do(req)
	if err != nil {
		api.svc.Logger.Errorf("Failed to request %s %v", url, err)
		return nil, err
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		api.svc.Logger.Errorf("Failed to read response body %s %v", url, err)
		return nil, errors.New("failed to read response body")
	}

	jsonContent := map[string]interface{}{}
	jsonErr := json.Unmarshal(body, &jsonContent)
	if jsonErr != nil {
		api.svc.Logger.Errorf("Failed to deserialize json %s %v", url, err)
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}
	return jsonContent, nil
}

func (api *API) GetInfo() (*models.InfoResponse, error) {
	info := models.InfoResponse{}
	backendType, _ := api.svc.cfg.Get("LNBackendType", "")
	unlockPasswordCheck, _ := api.svc.cfg.Get("UnlockPasswordCheck", "")
	info.SetupCompleted = unlockPasswordCheck != ""
	info.Running = api.svc.lnClient != nil
	info.BackendType = backendType

	if info.BackendType == LNDBackendType {
		info.IsMnemonicBackupDone = true // because otherwise it shows the popup
	} else {
		isMnemonicBackupDone, _ := api.svc.cfg.Get("IsMnemonicBackupDone", "")
		var err error
		parsedTime := time.Time{}
		if isMnemonicBackupDone != "" {
			parsedTime, err = time.Parse("2006-01-02 15:04:05", isMnemonicBackupDone)
			if err != nil {
				api.svc.Logger.Errorf("Error parsing time: %v", err)
				return nil, err
			}
		}
		currentTime := time.Now()
		sixMonthsAgo := currentTime.AddDate(0, -6, 0)
		info.IsMnemonicBackupDone = !(parsedTime.IsZero() || parsedTime.Before(sixMonthsAgo))
	}

	return &info, nil
}

func (api *API) GetMnemonic() *models.MnemonicResponse {
	resp := models.MnemonicResponse{}
	mnemonic, _ := api.svc.cfg.Get("Mnemonic", "")
	resp.Mnemonic = mnemonic
	return &resp
}

func (api *API) BackupMnemonic() error {
	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02 15:04:05")
	api.svc.cfg.SetUpdate("IsMnemonicBackupDone", timeString, "")
	return nil
}

func (api *API) Start(startRequest *models.StartRequest) error {
	return api.svc.StartApp(startRequest.UnlockPassword)
}

func (api *API) Setup(setupRequest *models.SetupRequest) error {
	api.svc.cfg.SavePasswordCheck(setupRequest.UnlockPassword)

	// update mnemonic check
	api.svc.cfg.SetUpdate("IsMnemonicBackupDone", setupRequest.IsMnemonicBackupDone, "")
	// only update non-empty values
	if setupRequest.LNBackendType != "" {
		api.svc.cfg.SetUpdate("LNBackendType", setupRequest.LNBackendType, "")
	}
	if setupRequest.BreezAPIKey != "" {
		api.svc.cfg.SetUpdate("BreezAPIKey", setupRequest.BreezAPIKey, setupRequest.UnlockPassword)
	}
	if setupRequest.Mnemonic != "" {
		api.svc.cfg.SetUpdate("Mnemonic", setupRequest.Mnemonic, setupRequest.UnlockPassword)
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

package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"

	models "github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/getAlby/nostr-wallet-connect/models/config"
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
	"github.com/getAlby/nostr-wallet-connect/models/lsp"
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
	expiry := updateAppRequest.ExpiresAt

	requestMethods := updateAppRequest.RequestMethods
	if requestMethods == "" {
		return fmt.Errorf("won't update an app to have no request methods")
	}
	newRequestMethods := strings.Split(requestMethods, " ")

	expiresAt := time.Time{}
	if expiry != "" {
		var err error
		expiresAt, err = time.Parse(time.RFC3339, expiry)
		if err != nil {
			return fmt.Errorf("invalid expiresAt: %v", err)
		}
	}

	if !expiresAt.IsZero() {
		expiresAt = time.Date(expiresAt.Year(), expiresAt.Month(), expiresAt.Day(), 23, 59, 59, 0, expiresAt.Location())
	}

	err := api.svc.db.Transaction(func(tx *gorm.DB) error {
		// Update existing permissions with new budget and expiry
		err := tx.Model(&AppPermission{}).Where("app_id", userApp.ID).Updates(map[string]interface{}{
			"ExpiresAt":     expiresAt,
			"MaxAmount":     maxAmount,
			"BudgetRenewal": budgetRenewal,
		}).Error
		if err != nil {
			return err
		}

		var existingPermissions []AppPermission
		if err := tx.Where("app_id = ?", userApp.ID).Find(&existingPermissions).Error; err != nil {
			return err
		}

		existingMethodMap := make(map[string]bool)
		for _, perm := range existingPermissions {
			existingMethodMap[perm.RequestMethod] = true
		}

		// Add new permissions
		for _, method := range newRequestMethods {
			if !existingMethodMap[method] {
				perm := AppPermission{
					App:           *userApp,
					RequestMethod: method,
					ExpiresAt:     expiresAt,
					MaxAmount:     maxAmount,
					BudgetRenewal: budgetRenewal,
				}
				if err := tx.Create(&perm).Error; err != nil {
					return err
				}
			}
			delete(existingMethodMap, method)
		}

		// Remove old permissions
		for method := range existingMethodMap {
			if err := tx.Where("app_id = ? AND request_method = ?", userApp.ID, method).Delete(&AppPermission{}).Error; err != nil {
				return err
			}
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

	permissions := []AppPermission{}
	api.svc.db.Find(&permissions)

	permissionsMap := make(map[uint][]AppPermission)
	for _, perm := range permissions {
		permissionsMap[perm.AppId] = append(permissionsMap[perm.AppId], perm)
	}

	// Can also fetch last events but is unused

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

		for _, permission := range permissionsMap[userApp.ID] {
			apiApp.RequestMethods = append(apiApp.RequestMethods, permission.RequestMethod)
			apiApp.ExpiresAt = &permission.ExpiresAt
			if permission.RequestMethod == NIP_47_PAY_INVOICE_METHOD {
				apiApp.BudgetRenewal = permission.BudgetRenewal
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

func (api *API) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.ListChannels(ctx)
}

func (api *API) ResetRouter(ctx context.Context) error {
	if api.svc.lnClient == nil {
		return errors.New("LNClient not started")
	}
	err := api.svc.lnClient.ResetRouter(ctx)
	if err != nil {
		return err
	}
	// Shut down the lnclient
	// The user will be forced to re-enter their unlock password to restart the node
	err = api.svc.lnClient.Shutdown()
	if err == nil {
		api.svc.lnClient = nil
	}
	return err
}

func (api *API) GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.GetNodeConnectionInfo(ctx)
}

func (api *API) ConnectPeer(ctx context.Context, connectPeerRequest *models.ConnectPeerRequest) error {
	if api.svc.lnClient == nil {
		return errors.New("LNClient not started")
	}
	return api.svc.lnClient.ConnectPeer(ctx, connectPeerRequest)
}

func (api *API) OpenChannel(ctx context.Context, openChannelRequest *models.OpenChannelRequest) (*models.OpenChannelResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.OpenChannel(ctx, openChannelRequest)
}

func (api *API) CloseChannel(ctx context.Context, closeChannelRequest *models.CloseChannelRequest) (*models.CloseChannelResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.CloseChannel(ctx, closeChannelRequest)
}

func (api *API) GetNewOnchainAddress(ctx context.Context) (*models.NewOnchainAddressResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	address, err := api.svc.lnClient.GetNewOnchainAddress(ctx)
	if err != nil {
		return nil, err
	}
	return &models.NewOnchainAddressResponse{
		Address: address,
	}, nil
}

func (api *API) RedeemOnchainFunds(ctx context.Context, toAddress string) (*models.RedeemOnchainFundsResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	txId, err := api.svc.lnClient.RedeemOnchainFunds(ctx, toAddress)
	if err != nil {
		return nil, err
	}
	return &models.RedeemOnchainFundsResponse{
		TxId: txId,
	}, nil
}

func (api *API) GetOnchainBalance(ctx context.Context) (*models.OnchainBalanceResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	balance, err := api.svc.lnClient.GetOnchainBalance(ctx)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (api *API) GetMempoolLightningNode(pubkey string) (interface{}, error) {
	url := api.svc.cfg.Env.MempoolApi + "/v1/lightning/nodes/" + pubkey

	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		api.svc.Logger.Errorf("Failed to create http request %s %v", url, err)
		return nil, err
	}

	res, err := client.Do(req)
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

func (api *API) NewWrappedInvoice(ctx context.Context, request *models.NewWrappedInvoiceRequest) (*models.NewWrappedInvoiceResponse, error) {
	var selectedLsp lsp.LSP
	switch request.LSP {
	case "VOLTAGE":
		selectedLsp = lsp.VoltageLSP()
	case "OLYMPUS":
		selectedLsp = lsp.OlympusLSP()
	default:
		return nil, errors.New("unknown LSP")
	}

	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	api.svc.Logger.Infoln("Requesting LSP info")

	var lspInfo lsp.LSPInfo
	{
		client := http.Client{
			Timeout: time.Second * 10,
		}
		req, err := http.NewRequest(http.MethodGet, selectedLsp.Url+"/info", nil)
		if err != nil {
			api.svc.Logger.Errorf("Failed to create lsp info request %s %v", selectedLsp.Url, err)
			return nil, err
		}

		res, err := client.Do(req)
		if err != nil {
			api.svc.Logger.Errorf("Failed to request lsp info %s %v", selectedLsp.Url, err)
			return nil, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			api.svc.Logger.Errorf("Failed to read response body %s %v", selectedLsp.Url, err)
			return nil, errors.New("failed to read response body")
		}

		err = json.Unmarshal(body, &lspInfo)
		if err != nil {
			api.svc.Logger.Errorf("Failed to deserialize json %s %v", selectedLsp.Url, err)
			return nil, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}
	}

	api.svc.Logger.Infoln("Requesting own node info")

	nodeInfo, err := api.svc.lnClient.GetInfo(ctx)
	if err != nil {
		api.svc.Logger.Errorf("Failed to request own node info %v", err)
		return nil, err
	}

	api.svc.Logger.Infof("Connecting to LSP node as a peer: %v", lspInfo)

	if !strings.HasPrefix(lspInfo.ConnectionMethods[0].Type, "ip") {
		api.svc.Logger.Errorf("Expected ipv4/ipv6 connection method, got %s", lspInfo.ConnectionMethods[0].Type)
		return nil, errors.New("unexpected LSP connection method")
	}

	err = api.ConnectPeer(ctx, &models.ConnectPeerRequest{
		Pubkey:  lspInfo.Pubkey,
		Address: lspInfo.ConnectionMethods[0].Address,
		Port:    lspInfo.ConnectionMethods[0].Port,
	})

	if err != nil {
		api.svc.Logger.Errorf("Failed to connect to peer %v", err)
		return nil, err
	}

	api.svc.Logger.Infoln("Requesting fee information")

	var feeResponse lsp.FeeResponse
	{
		client := http.Client{
			Timeout: time.Second * 10,
		}
		payloadBytes, err := json.Marshal(lsp.FeeRequest{
			AmountMsat: request.Amount * 1000,
			Pubkey:     nodeInfo.Pubkey,
		})
		if err != nil {
			return nil, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/fee", bodyReader)
		if err != nil {
			api.svc.Logger.Errorf("Failed to create lsp fee request %s %v", selectedLsp.Url, err)
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			api.svc.Logger.Errorf("Failed to request lsp fee %s %v", selectedLsp.Url, err)
			return nil, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			api.svc.Logger.Errorf("Failed to read response body %s %v", selectedLsp.Url, err)
			return nil, errors.New("failed to read response body")
		}

		err = json.Unmarshal(body, &feeResponse)
		if err != nil {
			api.svc.Logger.Errorf("Failed to deserialize json %s %v", selectedLsp.Url, err)
			return nil, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}

		api.svc.Logger.Infof("Fee response: %+v", feeResponse)
		if feeResponse.Id == "" {
			api.svc.Logger.Errorf("No fee id in fee response %v", feeResponse)
			return nil, fmt.Errorf("no fee id in fee response %v", feeResponse)
		}
	}

	api.svc.Logger.Infoln("Requesting own invoice")

	makeInvoiceResponse, err := api.svc.lnClient.MakeInvoice(ctx, int64(request.Amount)*1000-int64(feeResponse.FeeAmountMsat), "", "", 60*60)
	if err != nil {
		api.svc.Logger.Errorf("Failed to request own invoice %v", err)
		return nil, fmt.Errorf("failed to request own invoice %v", err)
	}

	api.svc.Logger.Infoln("Proposing invoice")

	var proposalResponse lsp.ProposalResponse
	{
		client := http.Client{
			Timeout: time.Second * 10,
		}
		payloadBytes, err := json.Marshal(lsp.ProposalRequest{
			Bolt11: makeInvoiceResponse.Invoice,
			FeeId:  feeResponse.Id,
		})
		if err != nil {
			return nil, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/proposal", bodyReader)
		if err != nil {
			api.svc.Logger.Errorf("Failed to create lsp fee request %s %v", selectedLsp.Url, err)
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			api.svc.Logger.Errorf("Failed to request lsp fee %s %v", selectedLsp.Url, err)
			return nil, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			api.svc.Logger.Errorf("Failed to read response body %s %v", selectedLsp.Url, err)
			return nil, errors.New("failed to read response body")
		}

		err = json.Unmarshal(body, &proposalResponse)
		if err != nil {
			api.svc.Logger.Errorf("Failed to deserialize json %s %v", selectedLsp.Url, err)
			return nil, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}
		api.svc.Logger.Infof("Proposal response: %+v", proposalResponse)
		if proposalResponse.Bolt11 == "" {
			api.svc.Logger.Errorf("No bolt11 in proposal response %v", proposalResponse)
			return nil, fmt.Errorf("no bolt11 in proposal response %v", proposalResponse)
		}
	}
	return &models.NewWrappedInvoiceResponse{
		WrappedInvoice: proposalResponse.Bolt11,
		Fee:            feeResponse.FeeAmountMsat / 1000,
	}, nil
}

func (api *API) GetInfo() (*models.InfoResponse, error) {
	info := models.InfoResponse{}
	backendType, _ := api.svc.cfg.Get("LNBackendType", "")
	unlockPasswordCheck, _ := api.svc.cfg.Get("UnlockPasswordCheck", "")
	info.SetupCompleted = unlockPasswordCheck != ""
	info.Running = api.svc.lnClient != nil
	info.BackendType = backendType

	if info.BackendType != config.LNDBackendType {
		nextBackupReminder, _ := api.svc.cfg.Get("NextBackupReminder", "")
		var err error
		parsedTime := time.Time{}
		if nextBackupReminder != "" {
			parsedTime, err = time.Parse(time.RFC3339, nextBackupReminder)
			if err != nil {
				api.svc.Logger.Errorf("Error parsing time: %v", err)
				return nil, err
			}
		}
		info.ShowBackupReminder = parsedTime.IsZero() || parsedTime.Before(time.Now())
	}

	return &info, nil
}

func (api *API) GetEncryptedMnemonic() *models.EncryptedMnemonicResponse {
	resp := models.EncryptedMnemonicResponse{}
	mnemonic, _ := api.svc.cfg.Get("Mnemonic", "")
	resp.Mnemonic = mnemonic
	return &resp
}

func (api *API) SetNextBackupReminder(backupReminderRequest *models.BackupReminderRequest) error {
	api.svc.cfg.SetUpdate("NextBackupReminder", backupReminderRequest.NextBackupReminder, "")
	return nil
}

func (api *API) Start(startRequest *models.StartRequest) error {
	return api.svc.StartApp(startRequest.UnlockPassword)
}

func (api *API) Setup(setupRequest *models.SetupRequest) error {
	info, err := api.GetInfo()
	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to get info")
		return err
	}
	if info.SetupCompleted {
		api.svc.Logger.Error("Cannot re-setup node")
		return errors.New("setup already completed")
	}

	api.svc.cfg.SavePasswordCheck(setupRequest.UnlockPassword)

	// update next backup reminder
	api.svc.cfg.SetUpdate("NextBackupReminder", setupRequest.NextBackupReminder, "")
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

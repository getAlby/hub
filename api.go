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
	"github.com/sirupsen/logrus"
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
			api.svc.Logger.WithField("pairingPublicKey", pairingPublicKey).Error("Invalid public key format")
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
			api.svc.Logger.WithField("expiresAt", createAppRequest.ExpiresAt).Error("Invalid expiresAt")
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

	// Because the above method has to stop the node to reset the router,
	// We also need to stop the lnclient and ask the user to start it again
	return api.Stop()
}

func (api *API) Stop() error {
	if api.svc.lnClient == nil {
		return errors.New("LNClient not started")
	}
	// stop the lnclient
	// The user will be forced to re-enter their unlock password to restart the node
	err := api.svc.StopLNClient()
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

func (api *API) CloseChannel(ctx context.Context, nodeId, channelId string) (*models.CloseChannelResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.CloseChannel(ctx, &lnclient.CloseChannelRequest{
		NodeId:    nodeId,
		ChannelId: channelId,
	})
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

func (api *API) GetBalances(ctx context.Context) (*models.BalancesResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	balances, err := api.svc.lnClient.GetBalances(ctx)
	if err != nil {
		return nil, err
	}
	return balances, nil
}

func (api *API) GetMempoolLightningNode(pubkey string) (interface{}, error) {
	url := api.svc.cfg.Env.MempoolApi + "/v1/lightning/nodes/" + pubkey

	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create http request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to send request")
		return nil, err
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	jsonContent := map[string]interface{}{}
	jsonErr := json.Unmarshal(body, &jsonContent)
	if jsonErr != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}
	return jsonContent, nil
}

func (api *API) NewInstantChannelInvoice(ctx context.Context, request *models.NewInstantChannelInvoiceRequest) (*models.NewInstantChannelInvoiceResponse, error) {
	var selectedLsp lsp.LSP
	switch request.LSP {
	case "VOLTAGE":
		selectedLsp = lsp.VoltageLSP()
	case "OLYMPUS":
		selectedLsp = lsp.OlympusLSP()
	case "ALBY":
		selectedLsp = lsp.AlbyPlebsLSP()
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
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to create lsp info request")
			return nil, err
		}

		res, err := client.Do(req)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to request lsp info")
			return nil, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to read response body")
			return nil, errors.New("failed to read response body")
		}

		err = json.Unmarshal(body, &lspInfo)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to deserialize json")
			return nil, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}
	}

	api.svc.Logger.Infoln("Requesting own node info")

	nodeInfo, err := api.svc.lnClient.GetInfo(ctx)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to request own node info", err)
		return nil, err
	}

	api.svc.Logger.WithField("lspInfo", lspInfo).Info("Connecting to LSP node as a peer")

	ipIndex := -1
	for i, cm := range lspInfo.ConnectionMethods {
		if strings.HasPrefix(cm.Type, "ip") {
			ipIndex = i
			break
		}
	}

	if ipIndex == -1 {
		api.svc.Logger.Error("No ipv4/ipv6 connection method found in LSP info")
		return nil, errors.New("unexpected LSP connection method")
	}

	err = api.ConnectPeer(ctx, &models.ConnectPeerRequest{
		Pubkey:  lspInfo.Pubkey,
		Address: lspInfo.ConnectionMethods[ipIndex].Address,
		Port:    lspInfo.ConnectionMethods[ipIndex].Port,
	})

	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to connect to peer")
		return nil, err
	}

	invoice := ""
	var fee uint64 = 0

	// TODO: switch on LSPType and extract to separate functions
	if selectedLsp.SupportsWrappedInvoices {

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
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to create lsp fee request")
				return nil, err
			}

			req.Header.Set("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to request lsp fee")
				return nil, err
			}

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to read response body")
				return nil, errors.New("failed to read response body")
			}

			err = json.Unmarshal(body, &feeResponse)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to deserialize json")
				return nil, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
			}

			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url":         selectedLsp.Url,
				"feeResponse": feeResponse,
			}).Info("Got fee response")
			if feeResponse.Id == "" {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"feeResponse": feeResponse,
				}).Error("No fee id in fee response")
				return nil, fmt.Errorf("no fee id in fee response %v", feeResponse)
			}
		}

		// because we don't want the sender to pay the fee
		// see: https://docs.voltage.cloud/voltage-lsp#gqBqV
		makeInvoiceResponse, err := api.svc.lnClient.MakeInvoice(ctx, int64(request.Amount)*1000-int64(feeResponse.FeeAmountMsat), "", "", 60*60)
		if err != nil {
			api.svc.Logger.WithError(err).Error("Failed to request own invoice")
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
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to create lsp fee request")
				return nil, err
			}

			req.Header.Set("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to request lsp fee")
				return nil, err
			}

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to read response body")
				return nil, errors.New("failed to read response body")
			}

			err = json.Unmarshal(body, &proposalResponse)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url": selectedLsp.Url,
				}).Error("Failed to deserialize json")
				return nil, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
			}
			api.svc.Logger.WithField("proposalResponse", proposalResponse).Info("Got proposal response")
			if proposalResponse.Bolt11 == "" {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"url":              selectedLsp.Url,
					"proposalResponse": proposalResponse,
				}).Error("No bolt11 in proposal response")
				return nil, fmt.Errorf("no bolt11 in proposal response %v", proposalResponse)
			}
		}
		invoice = proposalResponse.Bolt11
		fee = feeResponse.FeeAmountMsat / 1000
	} else {
		client := http.Client{
			Timeout: time.Second * 10,
		}
		payloadBytes, err := json.Marshal(lsp.NewInstantChannelRequest{
			ChannelAmount: request.Amount,
			NodePubkey:    nodeInfo.Pubkey,
		})
		if err != nil {
			return nil, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		// TODO: JSON error logging
		req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/new-channel", bodyReader)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to create new channel request")
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to request new channel invoice")
			return nil, err
		}

		// TODO: check status

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to read response body")
			return nil, errors.New("failed to read response body")
		}

		invoice = string(body)

		api.svc.Logger.WithField("invoice", invoice).Info("New Channel response")
	}

	return &models.NewInstantChannelInvoiceResponse{
		Invoice: invoice,
		Fee:     fee,
	}, nil
}

func (api *API) GetInfo() (*models.InfoResponse, error) {
	info := models.InfoResponse{}
	backendType, _ := api.svc.cfg.Get("LNBackendType", "")
	unlockPasswordCheck, _ := api.svc.cfg.Get("UnlockPasswordCheck", "")
	info.SetupCompleted = unlockPasswordCheck != ""
	info.Running = api.svc.lnClient != nil
	info.BackendType = backendType
	info.AlbyAuthUrl = api.svc.AlbyOAuthSvc.GetAuthUrl()
	info.AlbyUserIdentifier = api.svc.AlbyOAuthSvc.GetUserIdentifier()

	if info.BackendType != config.LNDBackendType {
		nextBackupReminder, _ := api.svc.cfg.Get("NextBackupReminder", "")
		var err error
		parsedTime := time.Time{}
		if nextBackupReminder != "" {
			parsedTime, err = time.Parse(time.RFC3339, nextBackupReminder)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{
					"nextBackupReminder": nextBackupReminder,
				}).Error("Error parsing time")
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

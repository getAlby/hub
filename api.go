package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"

	"github.com/getAlby/nostr-wallet-connect/alby"
	models "github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/getAlby/nostr-wallet-connect/models/config"
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
	"github.com/getAlby/nostr-wallet-connect/models/lsp"
	"github.com/getAlby/nostr-wallet-connect/nip47"
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

	expiresAt, err := api.parseExpiresAt(createAppRequest.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expiresAt: %v", err)
	}

	err = api.svc.db.Transaction(func(tx *gorm.DB) error {
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
			if !strings.Contains(nip47.CAPABILITIES, m) {
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

	requestMethods := updateAppRequest.RequestMethods
	if requestMethods == "" {
		return fmt.Errorf("won't update an app to have no request methods")
	}
	newRequestMethods := strings.Split(requestMethods, " ")

	expiresAt, err := api.parseExpiresAt(updateAppRequest.ExpiresAt)
	if err != nil {
		return fmt.Errorf("invalid expiresAt: %v", err)
	}

	err = api.svc.db.Transaction(func(tx *gorm.DB) error {
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
		expiresAt = appPerm.ExpiresAt
		if appPerm.RequestMethod == nip47.PAY_INVOICE_METHOD {
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
	// TODO: join apps and permissions
	apps := []App{}
	api.svc.db.Find(&apps)

	permissions := []AppPermission{}
	api.svc.db.Find(&permissions)

	permissionsMap := make(map[uint][]AppPermission)
	for _, perm := range permissions {
		permissionsMap[perm.AppId] = append(permissionsMap[perm.AppId], perm)
	}

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
			apiApp.ExpiresAt = permission.ExpiresAt
			if permission.RequestMethod == nip47.PAY_INVOICE_METHOD {
				apiApp.BudgetRenewal = permission.BudgetRenewal
				apiApp.MaxAmount = permission.MaxAmount
				if apiApp.MaxAmount > 0 {
					apiApp.BudgetUsage = api.svc.GetBudgetUsage(&permission)
				}
			}
		}

		var lastEvent RequestEvent
		lastEventResult := api.svc.db.Where("app_id = ?", userApp.ID).Order("id desc").Limit(1).Find(&lastEvent)
		if lastEventResult.RowsAffected > 0 {
			apiApp.LastEventAt = &lastEvent.CreatedAt
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

func (api *API) GetChannelPeerSuggestions(ctx context.Context) ([]alby.ChannelPeerSuggestion, error) {
	return api.svc.AlbyOAuthSvc.GetChannelPeerSuggestions(ctx)
}

func (api *API) ResetRouter(ctx context.Context, key string) error {
	if api.svc.lnClient == nil {
		return errors.New("LNClient not started")
	}
	err := api.svc.lnClient.ResetRouter(ctx, key)
	if err != nil {
		return err
	}

	// Because the above method has to stop the node to reset the router,
	// We also need to stop the lnclient and ask the user to start it again
	return api.Stop()
}

func (api *API) ChangeUnlockPassword(changeUnlockPasswordRequest *models.ChangeUnlockPasswordRequest) error {
	if api.svc.lnClient == nil {
		return errors.New("LNClient not started")
	}

	err := api.svc.cfg.ChangeUnlockPassword(changeUnlockPasswordRequest.CurrentUnlockPassword, changeUnlockPasswordRequest.NewUnlockPassword)

	if err != nil {
		api.svc.Logger.WithError(err).Error("failed to change unlock password")
		return err
	}

	// Because all the encrypted fields have changed
	// we also need to stop the lnclient and ask the user to start it again
	return api.Stop()
}

func (api *API) Stop() error {
	api.svc.Logger.Info("Running Stop command")
	if api.svc.lnClient == nil {
		return errors.New("LNClient not started")
	}
	// stop the lnclient
	// The user will be forced to re-enter their unlock password to restart the node
	err := api.svc.StopLNClient()
	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to stop LNClient")
	}
	return err
}

func (api *API) GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.GetNodeConnectionInfo(ctx)
}

func (api *API) GetNodeStatus(ctx context.Context) (*lnclient.NodeStatus, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.GetNodeStatus(ctx)
}

func (api *API) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.ListPeers(ctx)
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

func (api *API) CloseChannel(ctx context.Context, peerId, channelId string) (*models.CloseChannelResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.lnClient.CloseChannel(ctx, &lnclient.CloseChannelRequest{
		NodeId:    peerId,
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

func (api *API) SignMessage(ctx context.Context, message string) (*models.SignMessageResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}
	signature, err := api.svc.lnClient.SignMessage(ctx, message)
	if err != nil {
		return nil, err
	}
	return &models.SignMessageResponse{
		Message:   message,
		Signature: signature,
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

func (api *API) RequestMempoolApi(endpoint string) (interface{}, error) {
	url := api.svc.cfg.Env.MempoolApi + endpoint

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

	var jsonContent interface{}
	jsonErr := json.Unmarshal(body, &jsonContent)
	if jsonErr != nil {
		api.svc.Logger.WithError(jsonErr).WithFields(logrus.Fields{
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
	case "OLYMPUS_FLOW_2_0":
		selectedLsp = lsp.OlympusLSP()
	case "OLYMPUS_MUTINYNET_FLOW_2_0":
		selectedLsp = lsp.OlympusMutinynetFlowLSP()
	case "OLYMPUS_MUTINYNET_LSPS1":
		selectedLsp = lsp.OlympusMutinynetLSPS1LSP()
	case "ALBY":
		selectedLsp = lsp.AlbyPlebsLSP()
	case "ALBY_MUTINYNET":
		selectedLsp = lsp.AlbyMutinynetPlebsLSP()
	default:
		return nil, errors.New("unknown LSP")
	}

	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	api.svc.Logger.Infoln("Requesting LSP info")

	var lspInfo *LSPInfo
	var err error
	switch selectedLsp.LspType {
	case lsp.LSP_TYPE_FLOW_2_0:
		fallthrough
	case lsp.LSP_TYPE_PMLSP:
		lspInfo, err = api.getFlowLSPInfo(selectedLsp.Url + "/info")

	case lsp.LSP_TYPE_LSPS1:
		lspInfo, err = api.getLSPS1LSPInfo(selectedLsp.Url + "/get_info")

	default:
		return nil, fmt.Errorf("unsupported LSP type: %v", selectedLsp.LspType)
	}
	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to request LSP info")
		return nil, err
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

	err = api.ConnectPeer(ctx, &models.ConnectPeerRequest{
		Pubkey:  lspInfo.Pubkey,
		Address: lspInfo.Address,
		Port:    lspInfo.Port,
	})

	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to connect to peer")
		return nil, err
	}

	invoice := ""
	var fee uint64 = 0

	switch selectedLsp.LspType {
	case lsp.LSP_TYPE_FLOW_2_0:
		invoice, fee, err = api.requestFlow20WrappedInvoice(ctx, &selectedLsp, request.Amount, nodeInfo.Pubkey)
	case lsp.LSP_TYPE_PMLSP:
		invoice, fee, err = api.requestPMLSPInvoice(&selectedLsp, request.Amount, nodeInfo.Pubkey)
	case lsp.LSP_TYPE_LSPS1:
		invoice, fee, err = api.requestLSPS1Invoice(&selectedLsp, request.Amount, nodeInfo.Pubkey)

	default:
		return nil, fmt.Errorf("unsupported LSP type: %v", selectedLsp.LspType)
	}
	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to request invoice")
		return nil, err
	}

	newChannelResponse := &models.NewInstantChannelInvoiceResponse{
		Invoice: invoice,
		Fee:     fee,
	}

	api.svc.Logger.WithFields(logrus.Fields{
		"newChannelResponse": newChannelResponse,
	}).Info("New Channel response")

	return newChannelResponse, nil
}

// TODO: move to LSP file
type LSPInfo struct {
	Pubkey  string
	Address string
	Port    uint16
}

// TODO: move to LSP file
func (api *API) getLSPS1LSPInfo(url string) (*LSPInfo, error) {
	type LSPS1LSPInfo struct {
		// TODO: implement options
		Options interface{} `json:"options"`
		URIs    []string    `json:"uris"`
	}
	var lsps1LspInfo LSPS1LSPInfo
	client := http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to request lsp info")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	err = json.Unmarshal(body, &lsps1LspInfo)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}

	uri := lsps1LspInfo.URIs[0]

	// make sure it's a valid IPv4 URI
	regex := regexp.MustCompile(`^([0-9a-f]+)@([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+):([0-9]+)$`)
	parts := regex.FindStringSubmatch(uri)
	api.svc.Logger.WithField("parts", parts).Info("Split URI")
	if parts == nil || len(parts) != 4 {
		api.svc.Logger.WithField("parts", parts).Info("Unsupported URI")
		return nil, errors.New("could not decode LSP URI")
	}

	port, err := strconv.Atoi(parts[3])
	if err != nil {
		api.svc.Logger.WithField("port", parts[3]).WithError(err).Info("Failed to decode port number")

		return nil, err
	}

	return &LSPInfo{
		Pubkey:  parts[1],
		Address: parts[2],
		Port:    uint16(port),
	}, nil
}
func (api *API) getFlowLSPInfo(url string) (*LSPInfo, error) {
	type FlowLSPConnectionMethod struct {
		Address string `json:"address"`
		Port    uint16 `json:"port"`
		Type    string `json:"type"`
	}
	type FlowLSPInfo struct {
		Pubkey            string                    `json:"pubkey"`
		ConnectionMethods []FlowLSPConnectionMethod `json:"connection_methods"`
	}
	var flowLspInfo FlowLSPInfo
	client := http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to request lsp info")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	err = json.Unmarshal(body, &flowLspInfo)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}

	ipIndex := -1
	for i, cm := range flowLspInfo.ConnectionMethods {
		if strings.HasPrefix(cm.Type, "ip") {
			ipIndex = i
			break
		}
	}

	if ipIndex == -1 {
		api.svc.Logger.Error("No ipv4/ipv6 connection method found in LSP info")
		return nil, errors.New("unexpected LSP connection method")
	}

	return &LSPInfo{
		Pubkey:  flowLspInfo.Pubkey,
		Address: flowLspInfo.ConnectionMethods[ipIndex].Address,
		Port:    flowLspInfo.ConnectionMethods[ipIndex].Port,
	}, nil
}

// TODO: move to LSP file
func (api *API) requestFlow20WrappedInvoice(ctx context.Context, selectedLsp *lsp.LSP, amount uint64, pubkey string) (invoice string, fee uint64, err error) {
	api.svc.Logger.Infoln("Requesting fee information")

	type FeeRequest struct {
		AmountMsat uint64 `json:"amount_msat"`
		Pubkey     string `json:"pubkey"`
	}
	type FeeResponse struct {
		FeeAmountMsat uint64 `json:"fee_amount_msat"`
		Id            string `json:"id"`
	}

	var feeResponse FeeResponse
	{
		client := http.Client{
			Timeout: time.Second * 10,
		}
		payloadBytes, err := json.Marshal(FeeRequest{
			AmountMsat: amount * 1000,
			Pubkey:     pubkey,
		})
		if err != nil {
			return "", 0, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/fee", bodyReader)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to create lsp fee request")
			return "", 0, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to request lsp fee")
			return "", 0, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to read response body")
			return "", 0, errors.New("failed to read response body")
		}

		if res.StatusCode >= 300 {
			api.svc.Logger.WithFields(logrus.Fields{
				"body":       string(body),
				"statusCode": res.StatusCode,
			}).Error("fee endpoint returned non-success code")
			return "", 0, fmt.Errorf("fee endpoint returned non-success code: %s", string(body))
		}

		err = json.Unmarshal(body, &feeResponse)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to deserialize json")
			return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}

		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url":         selectedLsp.Url,
			"feeResponse": feeResponse,
		}).Info("Got fee response")
		if feeResponse.Id == "" {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"feeResponse": feeResponse,
			}).Error("No fee id in fee response")
			return "", 0, fmt.Errorf("no fee id in fee response %v", feeResponse)
		}
		fee = feeResponse.FeeAmountMsat / 1000
	}

	// because we don't want the sender to pay the fee
	// see: https://docs.voltage.cloud/voltage-lsp#gqBqV
	makeInvoiceResponse, err := api.svc.lnClient.MakeInvoice(ctx, int64(amount)*1000-int64(feeResponse.FeeAmountMsat), "", "", 60*60)
	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to request own invoice")
		return "", 0, fmt.Errorf("failed to request own invoice %v", err)
	}

	type ProposalRequest struct {
		Bolt11 string `json:"bolt11"`
		FeeId  string `json:"fee_id"`
	}
	type ProposalResponse struct {
		Bolt11 string `json:"jit_bolt11"`
	}

	api.svc.Logger.Infoln("Proposing invoice")

	var proposalResponse ProposalResponse
	{
		client := http.Client{
			Timeout: time.Second * 10,
		}
		payloadBytes, err := json.Marshal(ProposalRequest{
			Bolt11: makeInvoiceResponse.Invoice,
			FeeId:  feeResponse.Id,
		})
		if err != nil {
			return "", 0, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/proposal", bodyReader)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to create lsp fee request")
			return "", 0, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to request lsp fee")
			return "", 0, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to read response body")
			return "", 0, errors.New("failed to read response body")
		}

		if res.StatusCode >= 300 {
			api.svc.Logger.WithFields(logrus.Fields{
				"body":       string(body),
				"statusCode": res.StatusCode,
			}).Error("proposal endpoint returned non-success code")
			return "", 0, fmt.Errorf("proposal endpoint returned non-success code: %s", string(body))
		}

		err = json.Unmarshal(body, &proposalResponse)
		if err != nil {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to deserialize json")
			return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}
		api.svc.Logger.WithField("proposalResponse", proposalResponse).Info("Got proposal response")
		if proposalResponse.Bolt11 == "" {
			api.svc.Logger.WithError(err).WithFields(logrus.Fields{
				"url":              selectedLsp.Url,
				"proposalResponse": proposalResponse,
			}).Error("No bolt11 in proposal response")
			return "", 0, fmt.Errorf("no bolt11 in proposal response %v", proposalResponse)
		}
	}
	invoice = proposalResponse.Bolt11

	return invoice, fee, nil
}

func (api *API) requestPMLSPInvoice(selectedLsp *lsp.LSP, amount uint64, pubkey string) (invoice string, fee uint64, err error) {
	type NewInstantChannelRequest struct {
		Amount uint64 `json:"amount"`
		Pubkey string `json:"pubkey"`
	}

	client := http.Client{
		Timeout: time.Second * 10,
	}
	payloadBytes, err := json.Marshal(NewInstantChannelRequest{
		Amount: amount,
		Pubkey: pubkey,
	})
	if err != nil {
		return "", 0, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/new-channel", bodyReader)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to create new channel request")
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to request new channel invoice")
		return "", 0, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to read response body")
		return "", 0, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		api.svc.Logger.WithFields(logrus.Fields{
			"body":       string(body),
			"statusCode": res.StatusCode,
		}).Error("new-channel endpoint returned non-success code")
		return "", 0, fmt.Errorf("new-channel endpoint returned non-success code: %s", string(body))
	}

	var newChannelResponse lsp.NewInstantChannelResponse

	err = json.Unmarshal(body, &newChannelResponse)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to deserialize json")
		return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
	}

	invoice = newChannelResponse.Invoice
	fee = newChannelResponse.FeeAmountMsat / 1000

	return invoice, fee, nil
}

func (api *API) requestLSPS1Invoice(selectedLsp *lsp.LSP, amount uint64, pubkey string) (invoice string, fee uint64, err error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	type NewLSPS1ChannelRequest struct {
		PublicKey                    string `json:"public_key"`
		LSPBalanceSat                string `json:"lsp_balance_sat"`
		ClientBalanceSat             string `json:"client_balance_sat"`
		RequiredChannelConfirmations uint64 `json:"required_channel_confirmations"`
		FundingConfirmsWithinBlocks  uint64 `json:"funding_confirms_within_blocks"`
		ChannelExpiryBlocks          uint64 `json:"channel_expiry_blocks"`
		Token                        string `json:"token"`
		RefundOnchainAddress         string `json:"refund_onchain_address"`
		AnnounceChannel              bool   `json:"announce_channel"`
	}

	newLSPS1ChannelRequest := NewLSPS1ChannelRequest{
		PublicKey:                    pubkey,
		LSPBalanceSat:                strconv.FormatUint(amount, 10),
		ClientBalanceSat:             "0",
		RequiredChannelConfirmations: 0,
		FundingConfirmsWithinBlocks:  6,
		ChannelExpiryBlocks:          13000, // TODO: do not hardcode this
		Token:                        "",
		RefundOnchainAddress:         "tb1qtz33lpvxl543f4glfe3pecpznn4cked0yhvg78", // TODO: get from node
		AnnounceChannel:              false,                                        // TODO: this should be customizable
	}

	payloadBytes, err := json.Marshal(newLSPS1ChannelRequest)
	if err != nil {
		return "", 0, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/create_order", bodyReader)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to create new channel request")
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to request new channel invoice")
		return "", 0, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to read response body")
		return "", 0, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		api.svc.Logger.WithFields(logrus.Fields{
			"newLSPS1ChannelRequest": newLSPS1ChannelRequest,
			"body":                   string(body),
			"statusCode":             res.StatusCode,
		}).Error("create_order endpoint returned non-success code")
		return "", 0, fmt.Errorf("create_order endpoint returned non-success code: %s", string(body))
	}

	type NewLSPS1ChannelPayment struct {
		LightningInvoice string `json:"lightning_invoice"`
		FeeTotalSat      string `json:"fee_total_sat"`
	}
	type NewLSPS1ChannelResponse struct {
		Payment NewLSPS1ChannelPayment `json:"payment"`
	}

	var newChannelResponse NewLSPS1ChannelResponse

	err = json.Unmarshal(body, &newChannelResponse)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to deserialize json")
		return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
	}

	invoice = newChannelResponse.Payment.LightningInvoice
	fee, err = strconv.ParseUint(newChannelResponse.Payment.FeeTotalSat, 10, 64)
	if err != nil {
		api.svc.Logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to parse fee")
		return "", 0, fmt.Errorf("failed to parse fee %v", err)
	}

	return invoice, fee, nil
}

func (api *API) GetInfo(ctx context.Context) (*models.InfoResponse, error) {
	info := models.InfoResponse{}
	backendType, _ := api.svc.cfg.Get("LNBackendType", "")
	unlockPasswordCheck, _ := api.svc.cfg.Get("UnlockPasswordCheck", "")
	info.SetupCompleted = unlockPasswordCheck != ""
	info.Running = api.svc.lnClient != nil
	info.BackendType = backendType
	info.AlbyAuthUrl = api.svc.AlbyOAuthSvc.GetAuthUrl()
	info.OAuthRedirect = !api.svc.cfg.Env.IsDefaultClientId()
	albyUserIdentifier, err := api.svc.AlbyOAuthSvc.GetUserIdentifier()
	if err != nil {
		api.svc.Logger.WithError(err).Error("Failed to get alby user identifier")
		return nil, err
	}
	info.AlbyUserIdentifier = albyUserIdentifier
	info.AlbyAccountConnected = api.svc.AlbyOAuthSvc.IsConnected(ctx)
	if api.svc.lnClient != nil {
		// TODO: is there a better way to do this?
		if backendType == config.BreezBackendType {
			info.OnboardingCompleted = true
		} else {
			channels, err := api.ListChannels(ctx)
			if err != nil {
				api.svc.Logger.WithError(err).WithFields(logrus.Fields{}).Error("Failed to fetch channels")
				return nil, err
			}
			info.OnboardingCompleted = len(channels) > 0
		}

		nodeInfo, err := api.svc.lnClient.GetInfo(ctx)
		if err != nil {
			api.svc.Logger.WithError(err).Error("Failed to get alby user identifier")
			return nil, err
		}

		info.Network = nodeInfo.Network
	}

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

func (api *API) Setup(ctx context.Context, setupRequest *models.SetupRequest) error {
	info, err := api.GetInfo(ctx)
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

func (api *API) SendPaymentProbes(ctx context.Context, sendPaymentProbesRequest *models.SendPaymentProbesRequest) (*models.SendPaymentProbesResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	var errMessage string
	err := api.svc.lnClient.SendPaymentProbes(ctx, sendPaymentProbesRequest.Invoice)
	if err != nil {
		errMessage = err.Error()
	}

	return &models.SendPaymentProbesResponse{Error: errMessage}, nil
}

func (api *API) SendSpontaneousPaymentProbes(ctx context.Context, sendSpontaneousPaymentProbesRequest *models.SendSpontaneousPaymentProbesRequest) (*models.SendSpontaneousPaymentProbesResponse, error) {
	if api.svc.lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	var errMessage string
	err := api.svc.lnClient.SendSpontaneousPaymentProbes(ctx, sendSpontaneousPaymentProbesRequest.Amount, sendSpontaneousPaymentProbesRequest.NodeId)
	if err != nil {
		errMessage = err.Error()
	}

	return &models.SendSpontaneousPaymentProbesResponse{Error: errMessage}, nil
}

func (api *API) GetLogOutput(ctx context.Context, logType string, getLogRequest *models.GetLogOutputRequest) (*models.GetLogOutputResponse, error) {
	var err error
	var logData []byte

	if logType == models.LogTypeNode {
		if api.svc.lnClient == nil {
			return nil, errors.New("LNClient not started")
		}

		logData, err = api.svc.lnClient.GetLogOutput(ctx, getLogRequest.MaxLen)
		if err != nil {
			return nil, err
		}
	} else if logType == models.LogTypeApp {
		logFileName := api.svc.LogFilePath()

		logData, err = ReadFileTail(logFileName, getLogRequest.MaxLen)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid log type: '%s'", logType)
	}

	return &models.GetLogOutputResponse{Log: string(logData)}, nil
}

func (api *API) CreateBackup(basicBackupRequest *models.BasicBackupRequest, w io.Writer) error {
	var err error

	if !api.svc.cfg.CheckUnlockPassword(basicBackupRequest.UnlockPassword) {
		return errors.New("invalid unlock password")
	}

	workDir, err := filepath.Abs(api.svc.cfg.Env.Workdir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workdir: %w", err)
	}

	lnStorageDir := ""

	if api.svc.lnClient != nil {
		lnStorageDir, err = api.svc.lnClient.GetStorageDir()
		if err != nil {
			return fmt.Errorf("failed to get storage dir: %w", err)
		}
		api.svc.Logger.WithField("path", lnStorageDir).Info("Found node storage dir")
	}

	// Stop the app to ensure no new requests are processed.
	api.svc.StopApp()
	db, err := api.svc.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Closing the database leaves the service in an inconsistent state,
	// but that should not be a problem since the app is not expected
	// to be used after its data is exported.
	err = db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	var filesToArchive []string

	if lnStorageDir != "" {
		lnFiles, err := filepath.Glob(filepath.Join(workDir, lnStorageDir, "*"))
		if err != nil {
			return fmt.Errorf("failed to list files in the LNClient storage directory: %w", err)
		}
		api.svc.Logger.WithField("lnFiles", lnFiles).Info("Listed node storage dir")

		// Avoid backing up log files.
		slices.DeleteFunc(lnFiles, func(s string) bool {
			return filepath.Ext(s) == ".log"
		})

		filesToArchive = append(filesToArchive, lnFiles...)
	}

	cw, err := encryptingWriter(w, basicBackupRequest.UnlockPassword)
	if err != nil {
		return fmt.Errorf("failed to create encrypted writer: %w", err)
	}

	zw := zip.NewWriter(cw)
	defer zw.Close()

	addFileToZip := func(fsPath, zipPath string) error {
		inF, err := os.Open(fsPath)
		if err != nil {
			return fmt.Errorf("failed to open source file for reading: %w", err)
		}
		defer inF.Close()

		outW, err := zw.Create(zipPath)
		if err != nil {
			return fmt.Errorf("failed to create zip entry: %w", err)
		}

		_, err = io.Copy(outW, inF)
		return err
	}

	// Locate the main database file.
	dbFilePath := api.svc.cfg.Env.DatabaseUri
	// Add the database file to the archive.
	api.svc.Logger.WithField("nwc.db", dbFilePath).Info("adding nwc db to zip")
	err = addFileToZip(dbFilePath, "nwc.db")
	if err != nil {
		return fmt.Errorf("failed to write nwc db file to zip: %w", err)
	}

	for _, fileToArchive := range filesToArchive {
		api.svc.Logger.WithField("fileToArchive", fileToArchive).Info("adding file to zip")
		relPath, err := filepath.Rel(workDir, fileToArchive)
		if err != nil {
			return fmt.Errorf("failed to get relative path of input file: %w", err)
		}

		// Ensure forward slashes for zip format compatibility.
		err = addFileToZip(fileToArchive, filepath.ToSlash(relPath))
		if err != nil {
			return fmt.Errorf("failed to write input file to zip: %w", err)
		}
	}

	return nil
}

func (api *API) RestoreBackup(password string, r io.Reader) error {
	workDir, err := filepath.Abs(api.svc.cfg.Env.Workdir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workdir: %w", err)
	}

	if strings.HasPrefix(api.svc.cfg.Env.DatabaseUri, "file:") {
		return errors.New("cannot restore backup when database path is a file URI")
	}

	cr, err := decryptingReader(r, password)
	if err != nil {
		return fmt.Errorf("failed to create decrypted reader: %w", err)
	}

	tmpF, err := os.CreateTemp("", "nwc-*.bkp")
	if err != nil {
		return fmt.Errorf("failed to create temporary output file: %w", err)
	}
	tmpName := tmpF.Name()
	defer os.Remove(tmpName)
	defer tmpF.Close()

	zipSize, err := io.Copy(tmpF, cr)
	if err != nil {
		return fmt.Errorf("failed to decrypt backup data into temporary file: %w", err)
	}

	if err = tmpF.Sync(); err != nil {
		return fmt.Errorf("failed to flush temporary file: %w", err)
	}

	if _, err = tmpF.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to beginning of temporary file: %w", err)
	}

	zr, err := zip.NewReader(tmpF, zipSize)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	extractZipEntry := func(zipFile *zip.File) error {
		fsFilePath := filepath.Join(workDir, "restore", filepath.FromSlash(zipFile.Name))

		if err = os.MkdirAll(filepath.Dir(fsFilePath), 0700); err != nil {
			return fmt.Errorf("failed to create directory for zip entry: %w", err)
		}

		inF, err := zipFile.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry for reading: %w", err)
		}
		defer inF.Close()

		outF, err := os.OpenFile(fsFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		defer outF.Close()

		if _, err = io.Copy(outF, inF); err != nil {
			return fmt.Errorf("failed to write zip entry to destination file: %w", err)
		}

		return nil
	}

	api.svc.Logger.WithField("count", len(zr.File)).Info("Extracting files")
	for _, f := range zr.File {
		api.svc.Logger.WithField("file", f.Name).Info("Extracting file")
		if err = extractZipEntry(f); err != nil {
			return fmt.Errorf("failed to extract zip entry: %w", err)
		}
	}
	api.svc.Logger.WithField("count", len(zr.File)).Info("Extracted files")

	go func() {
		api.svc.Logger.Info("Backup restored. Shutting down Alby Hub...")
		// schedule node shutdown after a few seconds to ensure frontend updates
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

	return nil
}

func (api *API) parseExpiresAt(expiresAtString string) (*time.Time, error) {
	var expiresAt *time.Time
	if expiresAtString != "" {
		var err error
		expiresAtValue, err := time.Parse(time.RFC3339, expiresAtString)
		if err != nil {
			api.svc.Logger.WithField("expiresAt", expiresAtString).Error("Invalid expiresAt")
			return nil, fmt.Errorf("invalid expiresAt: %v", err)
		}
		expiresAtValue = time.Date(expiresAtValue.Year(), expiresAtValue.Month(), expiresAtValue.Day(), 23, 59, 59, 0, expiresAtValue.Location())
		expiresAt = &expiresAtValue
	}
	return expiresAt, nil
}

func encryptingWriter(w io.Writer, password string) (io.Writer, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	encKey := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err = rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	_, err = w.Write(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to write salt: %w", err)
	}

	_, err = w.Write(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to write IV: %w", err)
	}

	stream := cipher.NewOFB(block, iv)
	cw := &cipher.StreamWriter{
		S: stream,
		W: w,
	}

	return cw, nil
}

func decryptingReader(r io.Reader, password string) (io.Reader, error) {
	salt := make([]byte, 8)
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, fmt.Errorf("failed to read salt: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(r, iv); err != nil {
		return nil, fmt.Errorf("failed to read IV: %w", err)
	}

	encKey := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	stream := cipher.NewOFB(block, iv)
	cr := &cipher.StreamReader{
		S: stream,
		R: r,
	}

	return cr, nil
}

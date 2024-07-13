package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	permissions "github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/service"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/utils"
	"github.com/getAlby/hub/version"
)

type api struct {
	db             *gorm.DB
	dbSvc          db.DBService
	cfg            config.Config
	svc            service.Service
	permissionsSvc permissions.PermissionsService
	keys           keys.Keys
	albyOAuthSvc   alby.AlbyOAuthService
}

func NewAPI(svc service.Service, gormDB *gorm.DB, config config.Config, keys keys.Keys, albyOAuthSvc alby.AlbyOAuthService, eventPublisher events.EventPublisher) *api {
	return &api{
		db:             gormDB,
		dbSvc:          db.NewDBService(gormDB, eventPublisher),
		cfg:            config,
		svc:            svc,
		permissionsSvc: permissions.NewPermissionsService(gormDB, eventPublisher),
		keys:           keys,
		albyOAuthSvc:   albyOAuthSvc,
	}
}

func (api *api) CreateApp(createAppRequest *CreateAppRequest) (*CreateAppResponse, error) {
	expiresAt, err := api.parseExpiresAt(createAppRequest.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expiresAt: %v", err)
	}

	if len(createAppRequest.Scopes) == 0 {
		return nil, fmt.Errorf("won't create an app without scopes")
	}

	for _, scope := range createAppRequest.Scopes {
		if !slices.Contains(permissions.AllScopes(), scope) {
			return nil, fmt.Errorf("did not recognize requested scope: %s", scope)
		}
	}

	app, pairingSecretKey, err := api.dbSvc.CreateApp(createAppRequest.Name, createAppRequest.Pubkey, createAppRequest.MaxAmount, createAppRequest.BudgetRenewal, expiresAt, createAppRequest.Scopes)

	if err != nil {
		return nil, err
	}

	relayUrl := api.cfg.GetRelayUrl()

	responseBody := &CreateAppResponse{}
	responseBody.Name = createAppRequest.Name
	responseBody.Pubkey = app.NostrPubkey
	responseBody.PairingSecret = pairingSecretKey

	if createAppRequest.ReturnTo != "" {
		returnToUrl, err := url.Parse(createAppRequest.ReturnTo)
		if err == nil {
			query := returnToUrl.Query()
			query.Add("relay", relayUrl)
			query.Add("pubkey", api.keys.GetNostrPublicKey())
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
	responseBody.PairingUri = fmt.Sprintf("nostr+walletconnect://%s?relay=%s&secret=%s%s", api.keys.GetNostrPublicKey(), relayUrl, pairingSecretKey, lud16)
	return responseBody, nil
}

func (api *api) UpdateApp(userApp *db.App, updateAppRequest *UpdateAppRequest) error {
	maxAmount := updateAppRequest.MaxAmount
	budgetRenewal := updateAppRequest.BudgetRenewal

	if len(updateAppRequest.Scopes) == 0 {
		return fmt.Errorf("won't update an app to have no request methods")
	}
	newScopes := updateAppRequest.Scopes

	expiresAt, err := api.parseExpiresAt(updateAppRequest.ExpiresAt)
	if err != nil {
		return fmt.Errorf("invalid expiresAt: %v", err)
	}

	err = api.db.Transaction(func(tx *gorm.DB) error {
		// Update existing permissions with new budget and expiry
		err := tx.Model(&db.AppPermission{}).Where("app_id", userApp.ID).Updates(map[string]interface{}{
			"ExpiresAt":     expiresAt,
			"MaxAmount":     maxAmount,
			"BudgetRenewal": budgetRenewal,
		}).Error
		if err != nil {
			return err
		}

		var existingPermissions []db.AppPermission
		if err := tx.Where("app_id = ?", userApp.ID).Find(&existingPermissions).Error; err != nil {
			return err
		}

		existingScopeMap := make(map[string]bool)
		for _, perm := range existingPermissions {
			existingScopeMap[perm.Scope] = true
		}

		// Add new permissions
		for _, method := range newScopes {
			if !existingScopeMap[method] {
				perm := db.AppPermission{
					App:           *userApp,
					Scope:         method,
					ExpiresAt:     expiresAt,
					MaxAmount:     int(maxAmount),
					BudgetRenewal: budgetRenewal,
				}
				if err := tx.Create(&perm).Error; err != nil {
					return err
				}
			}
			delete(existingScopeMap, method)
		}

		// Remove old permissions
		for method := range existingScopeMap {
			if err := tx.Where("app_id = ? AND scope = ?", userApp.ID, method).Delete(&db.AppPermission{}).Error; err != nil {
				return err
			}
		}

		// commit transaction
		return nil
	})

	return err
}

func (api *api) DeleteApp(userApp *db.App) error {
	return api.db.Delete(userApp).Error
}

func (api *api) GetApp(userApp *db.App) *App {

	var lastEvent db.RequestEvent
	lastEventResult := api.db.Where("app_id = ?", userApp.ID).Order("id desc").Limit(1).Find(&lastEvent)

	paySpecificPermission := db.AppPermission{}
	appPermissions := []db.AppPermission{}
	var expiresAt *time.Time
	api.db.Where("app_id = ?", userApp.ID).Find(&appPermissions)

	requestMethods := []string{}
	for _, appPerm := range appPermissions {
		expiresAt = appPerm.ExpiresAt
		if appPerm.Scope == permissions.PAY_INVOICE_SCOPE {
			//find the pay_invoice-specific permissions
			paySpecificPermission = appPerm
		}
		requestMethods = append(requestMethods, appPerm.Scope)
	}

	//renewsIn := ""
	budgetUsage := uint64(0)
	maxAmount := uint64(paySpecificPermission.MaxAmount)
	if maxAmount > 0 {
		budgetUsage = api.permissionsSvc.GetBudgetUsage(&paySpecificPermission)
	}

	response := App{
		Name:          userApp.Name,
		Description:   userApp.Description,
		CreatedAt:     userApp.CreatedAt,
		UpdatedAt:     userApp.UpdatedAt,
		NostrPubkey:   userApp.NostrPubkey,
		ExpiresAt:     expiresAt,
		MaxAmount:     maxAmount,
		Scopes:        requestMethods,
		BudgetUsage:   budgetUsage,
		BudgetRenewal: paySpecificPermission.BudgetRenewal,
	}

	if lastEventResult.RowsAffected > 0 {
		response.LastEventAt = &lastEvent.CreatedAt
	}

	return &response

}

func (api *api) ListApps() ([]App, error) {
	// TODO: join dbApps and permissions
	dbApps := []db.App{}
	api.db.Find(&dbApps)

	appPermissions := []db.AppPermission{}
	api.db.Find(&appPermissions)

	permissionsMap := make(map[uint][]db.AppPermission)
	for _, perm := range appPermissions {
		permissionsMap[perm.AppId] = append(permissionsMap[perm.AppId], perm)
	}

	apiApps := []App{}
	for _, userApp := range dbApps {
		apiApp := App{
			// ID:          app.ID,
			Name:        userApp.Name,
			Description: userApp.Description,
			CreatedAt:   userApp.CreatedAt,
			UpdatedAt:   userApp.UpdatedAt,
			NostrPubkey: userApp.NostrPubkey,
		}

		for _, appPermission := range permissionsMap[userApp.ID] {
			apiApp.Scopes = append(apiApp.Scopes, appPermission.Scope)
			apiApp.ExpiresAt = appPermission.ExpiresAt
			if appPermission.Scope == permissions.PAY_INVOICE_SCOPE {
				apiApp.BudgetRenewal = appPermission.BudgetRenewal
				apiApp.MaxAmount = uint64(appPermission.MaxAmount)
				if apiApp.MaxAmount > 0 {
					apiApp.BudgetUsage = api.permissionsSvc.GetBudgetUsage(&appPermission)
				}
			}
		}

		var lastEvent db.RequestEvent
		lastEventResult := api.db.Where("app_id = ?", userApp.ID).Order("id desc").Limit(1).Find(&lastEvent)
		if lastEventResult.RowsAffected > 0 {
			apiApp.LastEventAt = &lastEvent.CreatedAt
		}

		apiApps = append(apiApps, apiApp)
	}
	return apiApps, nil
}

func (api *api) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().ListChannels(ctx)
}

func (api *api) GetChannelPeerSuggestions(ctx context.Context) ([]alby.ChannelPeerSuggestion, error) {
	return api.albyOAuthSvc.GetChannelPeerSuggestions(ctx)
}

func (api *api) ResetRouter(key string) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}
	err := api.svc.GetLNClient().ResetRouter(key)
	if err != nil {
		return err
	}

	// Because the above method has to stop the node to reset the router,
	// We also need to stop the lnclient and ask the user to start it again
	return api.Stop()
}

func (api *api) ChangeUnlockPassword(changeUnlockPasswordRequest *ChangeUnlockPasswordRequest) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}

	err := api.cfg.ChangeUnlockPassword(changeUnlockPasswordRequest.CurrentUnlockPassword, changeUnlockPasswordRequest.NewUnlockPassword)

	if err != nil {
		logger.Logger.WithError(err).Error("failed to change unlock password")
		return err
	}

	// Because all the encrypted fields have changed
	// we also need to stop the lnclient and ask the user to start it again
	return api.Stop()
}

func (api *api) Stop() error {
	logger.Logger.Info("Running Stop command")
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}

	// stop the lnclient, nostr relay etc.
	// The user will be forced to re-enter their unlock password to restart the node
	api.svc.StopApp()

	return nil
}

func (api *api) GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().GetNodeConnectionInfo(ctx)
}

func (api *api) GetNodeStatus(ctx context.Context) (*lnclient.NodeStatus, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().GetNodeStatus(ctx)
}

func (api *api) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().ListPeers(ctx)
}

func (api *api) ConnectPeer(ctx context.Context, connectPeerRequest *ConnectPeerRequest) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().ConnectPeer(ctx, connectPeerRequest)
}

func (api *api) OpenChannel(ctx context.Context, openChannelRequest *OpenChannelRequest) (*OpenChannelResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().OpenChannel(ctx, openChannelRequest)
}

func (api *api) DisconnectPeer(ctx context.Context, peerId string) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}
	logger.Logger.WithFields(logrus.Fields{
		"peer_id": peerId,
	}).Info("Disconnecting peer")
	return api.svc.GetLNClient().DisconnectPeer(ctx, peerId)
}

func (api *api) CloseChannel(ctx context.Context, peerId, channelId string, force bool) (*CloseChannelResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	logger.Logger.WithFields(logrus.Fields{
		"peer_id":    peerId,
		"channel_id": channelId,
		"force":      force,
	}).Info("Closing channel")
	return api.svc.GetLNClient().CloseChannel(ctx, &lnclient.CloseChannelRequest{
		NodeId:    peerId,
		ChannelId: channelId,
		Force:     force,
	})
}

func (api *api) UpdateChannel(ctx context.Context, updateChannelRequest *UpdateChannelRequest) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}
	logger.Logger.WithFields(logrus.Fields{
		"request": updateChannelRequest,
	}).Info("updating channel")
	return api.svc.GetLNClient().UpdateChannel(ctx, updateChannelRequest)
}

func (api *api) GetNewOnchainAddress(ctx context.Context) (string, error) {
	if api.svc.GetLNClient() == nil {
		return "", errors.New("LNClient not started")
	}
	address, err := api.svc.GetLNClient().GetNewOnchainAddress(ctx)
	if err != nil {
		return "", err
	}

	api.cfg.SetUpdate(config.OnchainAddressKey, address, "")

	return address, nil
}

func (api *api) GetUnusedOnchainAddress(ctx context.Context) (string, error) {
	if api.svc.GetLNClient() == nil {
		return "", errors.New("LNClient not started")
	}

	currentAddress, err := api.cfg.Get(config.OnchainAddressKey, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get current address from config")
		return "", err
	}

	if currentAddress != "" {
		// check if address has any transactions
		response, err := api.RequestEsploraApi("/address/" + currentAddress + "/txs")
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get current address transactions")
			return currentAddress, nil
		}

		transactions, ok := response.([]interface{})
		if !ok {
			logger.Logger.WithField("response", response).Error("Failed to cast esplora address txs response", response)
			return currentAddress, nil
		}

		if len(transactions) == 0 {
			// address has not been used yet
			return currentAddress, nil
		}
	}

	newAddress, err := api.GetNewOnchainAddress(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to retrieve new onchain address")
		return "", err
	}
	return newAddress, nil
}

func (api *api) SignMessage(ctx context.Context, message string) (*SignMessageResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	signature, err := api.svc.GetLNClient().SignMessage(ctx, message)
	if err != nil {
		return nil, err
	}
	return &SignMessageResponse{
		Message:   message,
		Signature: signature,
	}, nil
}

func (api *api) RedeemOnchainFunds(ctx context.Context, toAddress string) (*RedeemOnchainFundsResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	txId, err := api.svc.GetLNClient().RedeemOnchainFunds(ctx, toAddress)
	if err != nil {
		return nil, err
	}
	return &RedeemOnchainFundsResponse{
		TxId: txId,
	}, nil
}

func (api *api) GetBalances(ctx context.Context) (*BalancesResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	balances, err := api.svc.GetLNClient().GetBalances(ctx)
	if err != nil {
		return nil, err
	}
	return balances, nil
}

func (api *api) ListTransactions(ctx context.Context, limit uint64, offset uint64) (*ListTransactionsResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transactions, err := api.svc.GetLNClient().ListTransactions(ctx, 0, 0, limit, offset, false, "")
	if err != nil {
		return nil, err
	}
	return &transactions, nil
}

func (api *api) SendPayment(ctx context.Context, invoice string) (*SendPaymentResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	resp, err := api.svc.GetLNClient().SendPaymentSync(ctx, invoice)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (api *api) CreateInvoice(ctx context.Context, amount int64, description string) (*MakeInvoiceResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	invoice, err := api.svc.GetLNClient().MakeInvoice(ctx, amount, description, "", 0)
	return invoice, err
}

func (api *api) LookupInvoice(ctx context.Context, paymentHash string) (*LookupInvoiceResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	invoice, err := api.svc.GetLNClient().LookupInvoice(ctx, paymentHash)
	return invoice, err
}

// TODO: remove dependency on this endpoint
func (api *api) RequestMempoolApi(endpoint string) (interface{}, error) {
	url := api.cfg.GetEnv().MempoolApi + endpoint

	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create http request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to send request")
		return nil, err
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	var jsonContent interface{}
	jsonErr := json.Unmarshal(body, &jsonContent)
	if jsonErr != nil {
		logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}
	return jsonContent, nil
}

func (api *api) GetInfo(ctx context.Context) (*InfoResponse, error) {
	info := InfoResponse{}
	backendType, _ := api.cfg.Get("LNBackendType", "")
	unlockPasswordCheck, _ := api.cfg.Get("UnlockPasswordCheck", "")
	info.SetupCompleted = unlockPasswordCheck != ""
	info.Running = api.svc.GetLNClient() != nil
	info.BackendType = backendType
	info.AlbyAuthUrl = api.albyOAuthSvc.GetAuthUrl()
	info.OAuthRedirect = !api.cfg.GetEnv().IsDefaultClientId()
	info.Version = version.Tag
	albyUserIdentifier, err := api.albyOAuthSvc.GetUserIdentifier()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get alby user identifier")
		return nil, err
	}
	info.AlbyUserIdentifier = albyUserIdentifier
	info.AlbyAccountConnected = api.albyOAuthSvc.IsConnected(ctx)
	if api.svc.GetLNClient() != nil {
		nodeInfo, err := api.svc.GetLNClient().GetInfo(ctx)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get nodeInfo")
			return nil, err
		}

		info.Network = nodeInfo.Network
	}

	info.NextBackupReminder, _ = api.cfg.Get("NextBackupReminder", "")

	return &info, nil
}

func (api *api) GetEncryptedMnemonic() *EncryptedMnemonicResponse {
	resp := EncryptedMnemonicResponse{}
	mnemonic, _ := api.cfg.Get("Mnemonic", "")
	resp.Mnemonic = mnemonic
	return &resp
}

func (api *api) SetNextBackupReminder(backupReminderRequest *BackupReminderRequest) error {
	api.cfg.SetUpdate("NextBackupReminder", backupReminderRequest.NextBackupReminder, "")
	return nil
}

var startMutex sync.Mutex

func (api *api) Start(startRequest *StartRequest) error {
	if !startMutex.TryLock() {
		// do not allow to start twice in case this is somehow called twice
		return errors.New("app is already starting")
	}
	defer startMutex.Unlock()
	return api.svc.StartApp(startRequest.UnlockPassword)
}

func (api *api) Setup(ctx context.Context, setupRequest *SetupRequest) error {
	info, err := api.GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get info")
		return err
	}
	if info.SetupCompleted {
		logger.Logger.Error("Cannot re-setup node")
		return errors.New("setup already completed")
	}

	api.cfg.Setup(setupRequest.UnlockPassword)

	// TODO: move all below code to cfg.Setup()

	// update next backup reminder
	api.cfg.SetUpdate("NextBackupReminder", setupRequest.NextBackupReminder, "")
	// only update non-empty values
	if setupRequest.LNBackendType != "" {
		api.cfg.SetUpdate("LNBackendType", setupRequest.LNBackendType, "")
	}
	if setupRequest.BreezAPIKey != "" {
		api.cfg.SetUpdate("BreezAPIKey", setupRequest.BreezAPIKey, setupRequest.UnlockPassword)
	}
	if setupRequest.Mnemonic != "" {
		api.cfg.SetUpdate("Mnemonic", setupRequest.Mnemonic, setupRequest.UnlockPassword)
	}
	if setupRequest.GreenlightInviteCode != "" {
		api.cfg.SetUpdate("GreenlightInviteCode", setupRequest.GreenlightInviteCode, setupRequest.UnlockPassword)
	}
	if setupRequest.LNDAddress != "" {
		api.cfg.SetUpdate("LNDAddress", setupRequest.LNDAddress, setupRequest.UnlockPassword)
	}
	if setupRequest.LNDCertHex != "" {
		api.cfg.SetUpdate("LNDCertHex", setupRequest.LNDCertHex, setupRequest.UnlockPassword)
	}
	if setupRequest.LNDMacaroonHex != "" {
		api.cfg.SetUpdate("LNDMacaroonHex", setupRequest.LNDMacaroonHex, setupRequest.UnlockPassword)
	}

	if setupRequest.PhoenixdAddress != "" {
		api.cfg.SetUpdate("PhoenixdAddress", setupRequest.PhoenixdAddress, setupRequest.UnlockPassword)
	}
	if setupRequest.PhoenixdAuthorization != "" {
		api.cfg.SetUpdate("PhoenixdAuthorization", setupRequest.PhoenixdAuthorization, setupRequest.UnlockPassword)
	}

	if setupRequest.CashuMintUrl != "" {
		api.cfg.SetUpdate("CashuMintUrl", setupRequest.CashuMintUrl, setupRequest.UnlockPassword)
	}

	return nil
}

func (api *api) GetWalletCapabilities(ctx context.Context) (*WalletCapabilitiesResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	methods := api.svc.GetLNClient().GetSupportedNIP47Methods()
	notificationTypes := api.svc.GetLNClient().GetSupportedNIP47NotificationTypes()

	scopes, err := permissions.RequestMethodsToScopes(methods)
	if err != nil {
		return nil, err
	}
	if len(notificationTypes) > 0 {
		scopes = append(scopes, permissions.NOTIFICATIONS_SCOPE)
	}

	return &WalletCapabilitiesResponse{
		Methods:           methods,
		NotificationTypes: notificationTypes,
		Scopes:            scopes,
	}, nil
}

func (api *api) SendPaymentProbes(ctx context.Context, sendPaymentProbesRequest *SendPaymentProbesRequest) (*SendPaymentProbesResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	var errMessage string
	err := api.svc.GetLNClient().SendPaymentProbes(ctx, sendPaymentProbesRequest.Invoice)
	if err != nil {
		errMessage = err.Error()
	}

	return &SendPaymentProbesResponse{Error: errMessage}, nil
}

func (api *api) SendSpontaneousPaymentProbes(ctx context.Context, sendSpontaneousPaymentProbesRequest *SendSpontaneousPaymentProbesRequest) (*SendSpontaneousPaymentProbesResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	var errMessage string
	err := api.svc.GetLNClient().SendSpontaneousPaymentProbes(ctx, sendSpontaneousPaymentProbesRequest.Amount, sendSpontaneousPaymentProbesRequest.NodeId)
	if err != nil {
		errMessage = err.Error()
	}

	return &SendSpontaneousPaymentProbesResponse{Error: errMessage}, nil
}

func (api *api) GetNetworkGraph(nodeIds []string) (NetworkGraphResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().GetNetworkGraph(nodeIds)
}

func (api *api) SyncWallet() error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}
	api.svc.GetLNClient().UpdateLastWalletSyncRequest()
	return nil
}

func (api *api) GetLogOutput(ctx context.Context, logType string, getLogRequest *GetLogOutputRequest) (*GetLogOutputResponse, error) {
	var err error
	var logData []byte

	if logType == LogTypeNode {
		if api.svc.GetLNClient() == nil {
			return nil, errors.New("LNClient not started")
		}

		logData, err = api.svc.GetLNClient().GetLogOutput(ctx, getLogRequest.MaxLen)
		if err != nil {
			return nil, err
		}
	} else if logType == LogTypeApp {
		logFileName := logger.GetLogFilePath()

		logData, err = utils.ReadFileTail(logFileName, getLogRequest.MaxLen)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid log type: '%s'", logType)
	}

	return &GetLogOutputResponse{Log: string(logData)}, nil
}

func (api *api) parseExpiresAt(expiresAtString string) (*time.Time, error) {
	var expiresAt *time.Time
	if expiresAtString != "" {
		var err error
		expiresAtValue, err := time.Parse(time.RFC3339, expiresAtString)
		if err != nil {
			logger.Logger.WithField("expiresAt", expiresAtString).Error("Invalid expiresAt")
			return nil, fmt.Errorf("invalid expiresAt: %v", err)
		}
		expiresAtValue = time.Date(expiresAtValue.Year(), expiresAtValue.Month(), expiresAtValue.Day(), 23, 59, 59, 0, expiresAtValue.Location())
		expiresAt = &expiresAtValue
	}
	return expiresAt, nil
}

package api

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
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
	db               *gorm.DB
	appsSvc          apps.AppsService
	cfg              config.Config
	svc              service.Service
	permissionsSvc   permissions.PermissionsService
	keys             keys.Keys
	albyOAuthSvc     alby.AlbyOAuthService
	startupError     error
	startupErrorTime time.Time
}

func NewAPI(svc service.Service, gormDB *gorm.DB, config config.Config, keys keys.Keys, albyOAuthSvc alby.AlbyOAuthService, eventPublisher events.EventPublisher) *api {
	return &api{
		db:             gormDB,
		appsSvc:        apps.NewAppsService(gormDB, eventPublisher, keys, config),
		cfg:            config,
		svc:            svc,
		permissionsSvc: permissions.NewPermissionsService(gormDB, eventPublisher),
		keys:           keys,
		albyOAuthSvc:   albyOAuthSvc,
	}
}

func (api *api) CreateApp(createAppRequest *CreateAppRequest) (*CreateAppResponse, error) {
	if slices.Contains(createAppRequest.Scopes, constants.SUPERUSER_SCOPE) {
		if !api.cfg.CheckUnlockPassword(createAppRequest.UnlockPassword) {
			return nil, fmt.Errorf(
				"incorrect unlock password to create app with superuser permission")
		}
	}

	expiresAt, err := api.parseExpiresAt(createAppRequest.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("invalid expiresAt: %v", err)
	}

	for _, scope := range createAppRequest.Scopes {
		if !slices.Contains(permissions.AllScopes(), scope) {
			return nil, fmt.Errorf("did not recognize requested scope: %s", scope)
		}
	}

	app, pairingSecretKey, err := api.appsSvc.CreateApp(
		createAppRequest.Name,
		createAppRequest.Pubkey,
		createAppRequest.MaxAmountSat,
		createAppRequest.BudgetRenewal,
		expiresAt,
		createAppRequest.Scopes,
		createAppRequest.Isolated,
		createAppRequest.Metadata,
	)

	if err != nil {
		return nil, err
	}

	relayUrl := api.cfg.GetRelayUrl()

	lightningAddress, err := api.albyOAuthSvc.GetLightningAddress()
	if err != nil {
		return nil, err
	}

	responseBody := &CreateAppResponse{}
	responseBody.Id = app.ID
	responseBody.Name = app.Name
	responseBody.Pubkey = app.AppPubkey
	responseBody.PairingSecret = pairingSecretKey
	responseBody.WalletPubkey = *app.WalletPubkey
	responseBody.RelayUrl = relayUrl
	responseBody.Lud16 = lightningAddress

	if createAppRequest.ReturnTo != "" {
		returnToUrl, err := url.Parse(createAppRequest.ReturnTo)
		if err == nil {
			query := returnToUrl.Query()
			query.Add("relay", relayUrl)
			query.Add("pubkey", *app.WalletPubkey)
			if lightningAddress != "" && !app.Isolated {
				query.Add("lud16", lightningAddress)
			}
			returnToUrl.RawQuery = query.Encode()
			responseBody.ReturnTo = returnToUrl.String()
		}
	}

	var lud16 string
	if lightningAddress != "" && !app.Isolated {
		lud16 = fmt.Sprintf("&lud16=%s", lightningAddress)
	}
	responseBody.PairingUri = fmt.Sprintf("nostr+walletconnect://%s?relay=%s&secret=%s%s", *app.WalletPubkey, relayUrl, pairingSecretKey, lud16)

	return responseBody, nil
}

func (api *api) UpdateApp(userApp *db.App, updateAppRequest *UpdateAppRequest) error {
	name := updateAppRequest.Name

	if name == "" {
		return fmt.Errorf("won't update an app to have no name")
	}

	maxAmount := updateAppRequest.MaxAmountSat
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
		// Update app name if it is not the same
		if name != userApp.Name {
			err := tx.Model(&db.App{}).Where("id", userApp.ID).Update("name", name).Error
			if err != nil {
				return err
			}
		}

		// Update app isolation if it is not the same
		if updateAppRequest.Isolated != userApp.Isolated {
			err := tx.Model(&db.App{}).Where("id", userApp.ID).Update("isolated", updateAppRequest.Isolated).Error
			if err != nil {
				return err
			}
		}

		// Update the app metadata
		if updateAppRequest.Metadata != nil {
			var metadataBytes []byte
			var err error
			metadataBytes, err = json.Marshal(updateAppRequest.Metadata)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to serialize metadata")
				return err
			}
			err = tx.Model(&db.App{}).Where("id", userApp.ID).Update("metadata", datatypes.JSON(metadataBytes)).Error
			if err != nil {
				return err
			}
		}

		// Update existing permissions with new budget and expiry
		err = tx.Model(&db.AppPermission{}).Where("app_id", userApp.ID).Updates(map[string]interface{}{
			"ExpiresAt":     expiresAt,
			"MaxAmountSat":  maxAmount,
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

		if slices.Contains(newScopes, constants.SUPERUSER_SCOPE) && !existingScopeMap[constants.SUPERUSER_SCOPE] {
			return fmt.Errorf(
				"cannot update app to add superuser permission")
		}

		// Add new permissions
		for _, scope := range newScopes {
			if !existingScopeMap[scope] {
				perm := db.AppPermission{
					App:           *userApp,
					Scope:         scope,
					ExpiresAt:     expiresAt,
					MaxAmountSat:  int(maxAmount),
					BudgetRenewal: budgetRenewal,
				}
				if err := tx.Create(&perm).Error; err != nil {
					return err
				}
			}
			delete(existingScopeMap, scope)
		}

		// Remove old permissions
		for scope := range existingScopeMap {
			if err := tx.Where("app_id = ? AND scope = ?", userApp.ID, scope).Delete(&db.AppPermission{}).Error; err != nil {
				return err
			}
		}
		api.svc.GetEventPublisher().Publish(&events.Event{
			Event: "nwc_app_updated",
			Properties: map[string]interface{}{
				"name": name,
				"id":   userApp.ID,
			},
		})

		// commit transaction
		return nil
	})

	return err
}

func (api *api) DeleteApp(userApp *db.App) error {
	return api.appsSvc.DeleteApp(userApp)
}

func (api *api) GetApp(dbApp *db.App) *App {

	var lastEvent db.RequestEvent
	lastEventResult := api.db.Where("app_id = ?", dbApp.ID).Order("id desc").Limit(1).Find(&lastEvent)

	paySpecificPermission := db.AppPermission{}
	appPermissions := []db.AppPermission{}
	var expiresAt *time.Time
	api.db.Where("app_id = ?", dbApp.ID).Find(&appPermissions)

	requestMethods := []string{}
	for _, appPerm := range appPermissions {
		expiresAt = appPerm.ExpiresAt
		if appPerm.Scope == constants.PAY_INVOICE_SCOPE {
			// find the pay_invoice-specific permissions
			paySpecificPermission = appPerm
		}
		requestMethods = append(requestMethods, appPerm.Scope)
	}

	// renewsIn := ""
	budgetUsage := uint64(0)
	maxAmount := uint64(paySpecificPermission.MaxAmountSat)
	budgetUsage = queries.GetBudgetUsageSat(api.db, &paySpecificPermission)

	var metadata Metadata
	if dbApp.Metadata != nil {
		jsonErr := json.Unmarshal(dbApp.Metadata, &metadata)
		if jsonErr != nil {
			logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
				"app_id": dbApp.ID,
			}).Error("Failed to deserialize app metadata")
		}
	}

	walletPubkey := api.keys.GetNostrPublicKey()
	uniqueWalletPubkey := false
	if dbApp.WalletPubkey != nil {
		walletPubkey = *dbApp.WalletPubkey
		uniqueWalletPubkey = true
	}

	response := App{
		ID:                 dbApp.ID,
		Name:               dbApp.Name,
		Description:        dbApp.Description,
		CreatedAt:          dbApp.CreatedAt,
		UpdatedAt:          dbApp.UpdatedAt,
		AppPubkey:          dbApp.AppPubkey,
		ExpiresAt:          expiresAt,
		MaxAmountSat:       maxAmount,
		Scopes:             requestMethods,
		BudgetUsage:        budgetUsage,
		BudgetRenewal:      paySpecificPermission.BudgetRenewal,
		Isolated:           dbApp.Isolated,
		Metadata:           metadata,
		WalletPubkey:       walletPubkey,
		UniqueWalletPubkey: uniqueWalletPubkey,
	}

	if dbApp.Isolated {
		response.Balance = queries.GetIsolatedBalance(api.db, dbApp.ID)
	}

	if lastEventResult.RowsAffected > 0 {
		response.LastEventAt = &lastEvent.CreatedAt
	}

	return &response

}

func (api *api) ListApps() ([]App, error) {
	// TODO: join dbApps and permissions
	dbApps := []db.App{}
	err := api.db.Find(&dbApps).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list apps")
		return nil, err
	}

	appPermissions := []db.AppPermission{}
	err = api.db.Find(&appPermissions).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list app permissions")
		return nil, err
	}

	permissionsMap := make(map[uint][]db.AppPermission)
	for _, perm := range appPermissions {
		permissionsMap[perm.AppId] = append(permissionsMap[perm.AppId], perm)
	}

	apiApps := []App{}
	for _, dbApp := range dbApps {
		walletPubkey := api.keys.GetNostrPublicKey()
		uniqueWalletPubkey := false
		if dbApp.WalletPubkey != nil {
			walletPubkey = *dbApp.WalletPubkey
			uniqueWalletPubkey = true
		}
		apiApp := App{
			ID:                 dbApp.ID,
			Name:               dbApp.Name,
			Description:        dbApp.Description,
			CreatedAt:          dbApp.CreatedAt,
			UpdatedAt:          dbApp.UpdatedAt,
			AppPubkey:          dbApp.AppPubkey,
			Isolated:           dbApp.Isolated,
			WalletPubkey:       walletPubkey,
			UniqueWalletPubkey: uniqueWalletPubkey,
		}

		if dbApp.Isolated {
			apiApp.Balance = queries.GetIsolatedBalance(api.db, dbApp.ID)
		}

		for _, appPermission := range permissionsMap[dbApp.ID] {
			apiApp.Scopes = append(apiApp.Scopes, appPermission.Scope)
			apiApp.ExpiresAt = appPermission.ExpiresAt
			if appPermission.Scope == constants.PAY_INVOICE_SCOPE {
				apiApp.BudgetRenewal = appPermission.BudgetRenewal
				apiApp.MaxAmountSat = uint64(appPermission.MaxAmountSat)
				apiApp.BudgetUsage = queries.GetBudgetUsageSat(api.db, &appPermission)
			}
		}

		var lastEvent db.RequestEvent
		lastEventResult := api.db.Where("app_id = ?", dbApp.ID).Order("id desc").Limit(1).Find(&lastEvent)
		if lastEventResult.RowsAffected > 0 {
			apiApp.LastEventAt = &lastEvent.CreatedAt
		}

		var metadata Metadata
		if dbApp.Metadata != nil {
			jsonErr := json.Unmarshal(dbApp.Metadata, &metadata)
			if jsonErr != nil {
				logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
					"app_id": dbApp.ID,
				}).Error("Failed to deserialize app metadata")
			}
			apiApp.Metadata = metadata
		}

		apiApps = append(apiApps, apiApp)
	}
	return apiApps, nil
}

func (api *api) ListChannels(ctx context.Context) ([]Channel, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	channels, err := api.svc.GetLNClient().ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	apiChannels := []Channel{}
	for _, channel := range channels {
		status := "offline"
		if channel.Active {
			status = "online"
		} else if channel.Confirmations != nil && channel.ConfirmationsRequired != nil && *channel.ConfirmationsRequired > *channel.Confirmations {
			status = "opening"
		}

		apiChannels = append(apiChannels, Channel{
			LocalBalance:                             channel.LocalBalance,
			LocalSpendableBalance:                    channel.LocalSpendableBalance,
			RemoteBalance:                            channel.RemoteBalance,
			Id:                                       channel.Id,
			RemotePubkey:                             channel.RemotePubkey,
			FundingTxId:                              channel.FundingTxId,
			FundingTxVout:                            channel.FundingTxVout,
			Active:                                   channel.Active,
			Public:                                   channel.Public,
			InternalChannel:                          channel.InternalChannel,
			Confirmations:                            channel.Confirmations,
			ConfirmationsRequired:                    channel.ConfirmationsRequired,
			ForwardingFeeBaseMsat:                    channel.ForwardingFeeBaseMsat,
			UnspendablePunishmentReserve:             channel.UnspendablePunishmentReserve,
			CounterpartyUnspendablePunishmentReserve: channel.CounterpartyUnspendablePunishmentReserve,
			Error:                                    channel.Error,
			IsOutbound:                               channel.IsOutbound,
			Status:                                   status,
		})
	}

	slices.SortFunc(apiChannels, func(a, b Channel) int {
		// sort by channel size first
		aSize := a.LocalBalance + a.RemoteBalance
		bSize := b.LocalBalance + b.RemoteBalance
		if aSize != bSize {
			return int(bSize - aSize)
		}

		// then by local balance in the channel
		if a.LocalBalance != b.LocalBalance {
			return int(b.LocalBalance - a.LocalBalance)
		}

		// finally sort by channel ID to prevent sort randomly changing
		return strings.Compare(b.Id, a.Id)
	})

	return apiChannels, nil
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

	autoUnlockPassword, err := api.cfg.Get("AutoUnlockPassword", "")
	if err != nil {
		return err
	}
	if autoUnlockPassword != "" {
		return errors.New("Please disable auto-unlock before using this feature")
	}

	err = api.cfg.ChangeUnlockPassword(changeUnlockPasswordRequest.CurrentUnlockPassword, changeUnlockPasswordRequest.NewUnlockPassword)

	if err != nil {
		logger.Logger.WithError(err).Error("failed to change unlock password")
		return err
	}

	// Because all the encrypted fields have changed
	// we also need to stop the lnclient and ask the user to start it again
	return api.Stop()
}

func (api *api) SetAutoUnlockPassword(unlockPassword string) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}

	err := api.cfg.SetAutoUnlockPassword(unlockPassword)

	if err != nil {
		logger.Logger.WithError(err).Error("failed to set auto unlock password")
		return err
	}

	return nil
}

func (api *api) Stop() error {
	if !startMutex.TryLock() {
		// do not allow to stop twice in case this is somehow called twice
		return errors.New("app is busy")
	}
	defer startMutex.Unlock()

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

func (api *api) GetAutoSwapsConfig() (*GetAutoSwapsConfigResponse, error) {
	swapBalanceThresholdStr, _ := api.cfg.Get(config.AutoSwapBalanceThresholdKey, "")
	swapAmountStr, _ := api.cfg.Get(config.AutoSwapAmountKey, "")
	swapDestination, _ := api.cfg.Get(config.AutoSwapDestinationKey, "")

	enabled := swapBalanceThresholdStr != "" &&
		swapAmountStr != "" &&
		swapDestination != ""
	var swapBalanceThreshold, swapAmount uint64
	if enabled {
		var err error
		if swapBalanceThreshold, err = strconv.ParseUint(swapBalanceThresholdStr, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid autoswap balance threshold: %w", err)
		}
		if swapAmount, err = strconv.ParseUint(swapAmountStr, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid autoswap amount: %w", err)
		}
	}

	swapFees, err := api.svc.GetSwapsService().CalculateFee()
	if err != nil {
		logger.Logger.WithError(err).Error("failed to calculate fee info")
		return nil, err
	}

	return &GetAutoSwapsConfigResponse{
		Enabled:          enabled,
		BalanceThreshold: swapBalanceThreshold,
		SwapAmount:       swapAmount,
		Destination:      swapDestination,
		AlbyServiceFee:   swapFees.AlbyServiceFee,
		BoltzServiceFee:  swapFees.BoltzServiceFee,
		BoltzNetworkFee:  swapFees.BoltzNetworkFee,
	}, nil
}

func (api *api) EnableAutoSwaps(ctx context.Context, enableAutoSwapsRequest *EnableAutoSwapsRequest) error {
	err := api.cfg.SetUpdate(config.AutoSwapBalanceThresholdKey, strconv.FormatUint(enableAutoSwapsRequest.BalanceThreshold, 10), "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save autoswap balance threshold to config")
		return err
	}

	err = api.cfg.SetUpdate(config.AutoSwapAmountKey, strconv.FormatUint(enableAutoSwapsRequest.SwapAmount, 10), "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save autoswap amount to config")
		return err
	}

	err = api.cfg.SetUpdate(config.AutoSwapDestinationKey, enableAutoSwapsRequest.Destination, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save autoswap destination to config")
		return err
	}

	return api.svc.StartAutoSwaps()
}

func (api *api) DisableAutoSwaps() error {
	if err := api.cfg.SetUpdate(config.AutoSwapBalanceThresholdKey, "", ""); err != nil {
		logger.Logger.WithError(err).Error("Failed to remove autoswap balance threshold")
		return err
	}
	if err := api.cfg.SetUpdate(config.AutoSwapAmountKey, "", ""); err != nil {
		logger.Logger.WithError(err).Error("Failed to remove autoswap amount")
		return err
	}
	if err := api.cfg.SetUpdate(config.AutoSwapDestinationKey, "", ""); err != nil {
		logger.Logger.WithError(err).Error("Failed to remove autoswap destination")
		return err
	}

	api.svc.GetSwapsService().StopAutoSwaps()

	return nil
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

func (api *api) MakeOffer(ctx context.Context, description string) (string, error) {
	if api.svc.GetLNClient() == nil {
		return "", errors.New("LNClient not started")
	}
	offer, err := api.svc.GetLNClient().MakeOffer(ctx, description)
	if err != nil {
		return "", err
	}

	return offer, nil
}

func (api *api) GetNewOnchainAddress(ctx context.Context) (string, error) {
	if api.svc.GetLNClient() == nil {
		return "", errors.New("LNClient not started")
	}
	address, err := api.svc.GetLNClient().GetNewOnchainAddress(ctx)
	if err != nil {
		return "", err
	}

	err = api.cfg.SetUpdate(config.OnchainAddressKey, address, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save new onchain address to config")
	}

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

func (api *api) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (*RedeemOnchainFundsResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	txId, err := api.svc.GetLNClient().RedeemOnchainFunds(ctx, toAddress, amount, feeRate, sendAll)
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
	balances, err := api.svc.GetLNClient().GetBalances(ctx, false)
	if err != nil {
		return nil, err
	}
	return balances, nil
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
	ldkVssEnabled, _ := api.cfg.Get("LdkVssEnabled", "")
	autoUnlockPassword, _ := api.cfg.Get("AutoUnlockPassword", "")
	info.SetupCompleted = api.cfg.SetupCompleted()
	info.Currency = api.cfg.GetCurrency()
	info.StartupState = api.svc.GetStartupState()
	if api.startupError != nil {
		info.StartupError = api.startupError.Error()
		info.StartupErrorTime = api.startupErrorTime
	}
	info.Running = api.svc.GetLNClient() != nil
	info.BackendType = backendType
	info.AlbyAuthUrl = api.albyOAuthSvc.GetAuthUrl()
	info.OAuthRedirect = !api.cfg.GetEnv().IsDefaultClientId()
	info.Version = version.Tag
	info.EnableAdvancedSetup = api.cfg.GetEnv().EnableAdvancedSetup
	info.LdkVssEnabled = ldkVssEnabled == "true"
	info.VssSupported = backendType == config.LDKBackendType && api.cfg.GetEnv().LDKVssUrl != ""
	info.AutoUnlockPasswordEnabled = autoUnlockPassword != ""
	info.AutoUnlockPasswordSupported = api.cfg.GetEnv().IsDefaultClientId()
	albyUserIdentifier, err := api.albyOAuthSvc.GetUserIdentifier()
	info.Relay = api.cfg.GetRelayUrl()
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

	info.NodeAlias, _ = api.cfg.Get("NodeAlias", "")

	return &info, nil
}

func (api *api) SetCurrency(currency string) error {
	if currency == "" {
		return fmt.Errorf("currency value cannot be empty")
	}

	err := api.cfg.SetCurrency(currency)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to update currency")
		return err
	}

	return nil
}

func (api *api) SetNodeAlias(nodeAlias string) error {
	err := api.cfg.SetUpdate("NodeAlias", nodeAlias, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save node alias to config")
		return err
	}

	return nil
}

func (api *api) GetMnemonic(unlockPassword string) (*MnemonicResponse, error) {
	if !api.cfg.CheckUnlockPassword(unlockPassword) {
		return nil, fmt.Errorf("wrong password")
	}

	mnemonic, err := api.cfg.Get("Mnemonic", unlockPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch encryption key: %w", err)
	}

	resp := MnemonicResponse{
		Mnemonic: mnemonic,
	}

	return &resp, nil
}

func (api *api) SetNextBackupReminder(backupReminderRequest *BackupReminderRequest) error {
	err := api.cfg.SetUpdate("NextBackupReminder", backupReminderRequest.NextBackupReminder, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save next backup reminder to config")
	}
	return nil
}

var startMutex sync.Mutex

func (api *api) Start(startRequest *StartRequest) {
	api.startupError = nil
	err := api.StartInternal(startRequest)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to start node")
		api.startupError = err
		api.startupErrorTime = time.Now()
	}
}

func (api *api) StartInternal(startRequest *StartRequest) (err error) {
	if !startMutex.TryLock() {
		// do not allow to start twice in case this is somehow called twice
		return errors.New("app is busy")
	}
	defer startMutex.Unlock()
	return api.svc.StartApp(startRequest.UnlockPassword)
}

func (api *api) Setup(ctx context.Context, setupRequest *SetupRequest) error {
	if !startMutex.TryLock() {
		// do not allow to start twice in case this is somehow called twice
		return errors.New("app is busy")
	}
	defer startMutex.Unlock()
	info, err := api.GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get info")
		return err
	}
	if info.SetupCompleted {
		logger.Logger.Error("Cannot re-setup node")
		return errors.New("setup already completed")
	}

	if setupRequest.UnlockPassword == "" {
		return errors.New("no unlock password provided")
	}

	err = api.cfg.SaveUnlockPasswordCheck(setupRequest.UnlockPassword)
	if err != nil {
		return err
	}

	// update next backup reminder
	err = api.cfg.SetUpdate("NextBackupReminder", setupRequest.NextBackupReminder, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save next backup reminder")
	}

	// only update non-empty values
	if setupRequest.LNBackendType != "" {
		err = api.cfg.SetUpdate("LNBackendType", setupRequest.LNBackendType, "")
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save backend type")
			return err
		}
	}
	if setupRequest.Mnemonic != "" {
		err = api.cfg.SetUpdate("Mnemonic", setupRequest.Mnemonic, setupRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save encrypted mnemonic")
			return err
		}
	}
	if setupRequest.LNDAddress != "" {
		err = api.cfg.SetUpdate("LNDAddress", setupRequest.LNDAddress, setupRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save lnd address")
			return err
		}
	}
	if setupRequest.LNDCertHex != "" {
		err = api.cfg.SetUpdate("LNDCertHex", setupRequest.LNDCertHex, setupRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save lnd cert hex")
			return err
		}
	}
	if setupRequest.LNDMacaroonHex != "" {
		err = api.cfg.SetUpdate("LNDMacaroonHex", setupRequest.LNDMacaroonHex, setupRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save lnd macaroon hex")
			return err
		}
	}

	if setupRequest.PhoenixdAddress != "" {
		err = api.cfg.SetUpdate("PhoenixdAddress", setupRequest.PhoenixdAddress, setupRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save phoenix address")
			return err
		}
	}
	if setupRequest.PhoenixdAuthorization != "" {
		err = api.cfg.SetUpdate("PhoenixdAuthorization", setupRequest.PhoenixdAuthorization, setupRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save phoenix auth")
			return err
		}
	}

	if setupRequest.CashuMintUrl != "" {
		err = api.cfg.SetUpdate("CashuMintUrl", setupRequest.CashuMintUrl, setupRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save cashu mint url")
			return err
		}
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
		scopes = append(scopes, constants.NOTIFICATIONS_SCOPE)
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

func (api *api) MigrateNodeStorage(ctx context.Context, to string) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}
	if to != "VSS" {
		return fmt.Errorf("Migration type not supported: %s", to)
	}

	ldkVssEnabled, err := api.cfg.Get("LdkVssEnabled", "")
	if err != nil {
		return err
	}

	if ldkVssEnabled == "true" {
		return errors.New("VSS already enabled")
	}

	if api.cfg.GetEnv().LDKVssUrl == "" {
		return errors.New("No VSS URL set")
	}

	api.cfg.SetUpdate("LdkVssEnabled", "true", "")
	api.cfg.SetUpdate("LdkMigrateStorage", "VSS", "")
	return api.Stop()
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

func (api *api) GetNetworkGraph(ctx context.Context, nodeIds []string) (NetworkGraphResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().GetNetworkGraph(ctx, nodeIds)
}

func (api *api) SyncWallet() error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}
	api.svc.GetLNClient().UpdateLastWalletSyncRequest()
	return nil
}
func (api *api) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	return api.svc.GetLNClient().ListOnchainTransactions(ctx)
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
		if logFileName == "" {
			logData = []byte("file log is disabled")
		} else {
			logData, err = utils.ReadFileTail(logFileName, getLogRequest.MaxLen)
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, fmt.Errorf("invalid log type: '%s'", logType)
	}

	return &GetLogOutputResponse{Log: string(logData)}, nil
}

func (api *api) Health(ctx context.Context) (*HealthResponse, error) {
	var alarms []HealthAlarm

	albyInfo, err := api.albyOAuthSvc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !albyInfo.Healthy {
		alarms = append(alarms, NewHealthAlarm(HealthAlarmKindAlbyService, albyInfo.Incidents))
	}

	isNostrRelayReady := api.svc.IsRelayReady()
	if !isNostrRelayReady {
		alarms = append(alarms, NewHealthAlarm(HealthAlarmKindNostrRelayOffline, nil))
	}

	ldkVssEnabled, _ := api.cfg.Get("LdkVssEnabled", "")
	if ldkVssEnabled == "true" {
		albyMe, err := api.albyOAuthSvc.GetMe(ctx)
		if err != nil {
			return nil, err
		}
		if albyMe.Subscription.PlanCode == "" {
			alarms = append(alarms, NewHealthAlarm(HealthAlarmKindVssNoSubscription, nil))
		}
	}

	lnClient := api.svc.GetLNClient()

	if lnClient != nil {
		nodeStatus, _ := lnClient.GetNodeStatus(ctx)
		if nodeStatus == nil || !nodeStatus.IsReady {
			alarms = append(alarms, NewHealthAlarm(HealthAlarmKindNodeNotReady, nodeStatus))
		}

		channels, err := lnClient.ListChannels(ctx)
		if err != nil {
			return nil, err
		}

		offlineChannels := slices.DeleteFunc(channels, func(channel lnclient.Channel) bool {
			if channel.Active {
				return true
			}
			if channel.Confirmations == nil || channel.ConfirmationsRequired == nil {
				return false
			}
			return *channel.Confirmations < *channel.ConfirmationsRequired
		})

		if len(offlineChannels) > 0 {
			alarms = append(alarms, NewHealthAlarm(HealthAlarmKindChannelsOffline, nil))
		}
	}

	return &HealthResponse{Alarms: alarms}, nil
}

func (api *api) GetCustomNodeCommands() (*CustomNodeCommandsResponse, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	allCommandDefs := lnClient.GetCustomNodeCommandDefinitions()
	commandDefs := make([]CustomNodeCommandDef, 0, len(allCommandDefs))
	for _, commandDef := range allCommandDefs {
		argDefs := make([]CustomNodeCommandArgDef, 0, len(commandDef.Args))
		for _, argDef := range commandDef.Args {
			argDefs = append(argDefs, CustomNodeCommandArgDef{
				Name:        argDef.Name,
				Description: argDef.Description,
			})
		}
		commandDefs = append(commandDefs, CustomNodeCommandDef{
			Name:        commandDef.Name,
			Description: commandDef.Description,
			Args:        argDefs,
		})
	}

	return &CustomNodeCommandsResponse{Commands: commandDefs}, nil
}

func (api *api) ExecuteCustomNodeCommand(ctx context.Context, command string) (interface{}, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	// Split command line into arguments. Command name must be the first argument.
	parsedArgs, err := utils.ParseCommandLine(command)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node command: %w", err)
	} else if len(parsedArgs) == 0 {
		return nil, errors.New("no command provided")
	}

	// Look up the requested command definition.
	allCommandDefs := lnClient.GetCustomNodeCommandDefinitions()
	commandDefIdx := slices.IndexFunc(allCommandDefs, func(def lnclient.CustomNodeCommandDef) bool {
		return def.Name == parsedArgs[0]
	})
	if commandDefIdx < 0 {
		return nil, fmt.Errorf("unknown command: %q", parsedArgs[0])
	}

	// Build flag set.
	commandDef := allCommandDefs[commandDefIdx]
	flagSet := flag.NewFlagSet(commandDef.Name, flag.ContinueOnError)
	for _, argDef := range commandDef.Args {
		flagSet.String(argDef.Name, "", argDef.Description)
	}

	if err = flagSet.Parse(parsedArgs[1:]); err != nil {
		return nil, fmt.Errorf("failed to parse command arguments: %w", err)
	}

	// Collect flags that have been set.
	argValues := make(map[string]string)
	flagSet.Visit(func(f *flag.Flag) {
		argValues[f.Name] = f.Value.String()
	})

	reqArgs := make([]lnclient.CustomNodeCommandArg, 0, len(argValues))
	for _, argDef := range commandDef.Args {
		if argValue, ok := argValues[argDef.Name]; ok {
			reqArgs = append(reqArgs, lnclient.CustomNodeCommandArg{
				Name:  argDef.Name,
				Value: argValue,
			})
		}
	}

	nodeResp, err := lnClient.ExecuteCustomNodeCommand(ctx, &lnclient.CustomNodeCommandRequest{
		Name: commandDef.Name,
		Args: reqArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("node failed to execute custom command: %w", err)
	}

	return nodeResp.Response, nil
}

func (api *api) SendEvent(event string) {
	api.svc.GetEventPublisher().Publish(&events.Event{
		Event: event,
	})
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
		expiresAt = &expiresAtValue
	}
	return expiresAt, nil
}

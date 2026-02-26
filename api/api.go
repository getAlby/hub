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
	"github.com/getAlby/hub/swaps"
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
	albySvc          alby.AlbyService
	startupError     error
	startupErrorTime time.Time
	eventPublisher   events.EventPublisher
}

func NewAPI(svc service.Service, gormDB *gorm.DB, config config.Config, keys keys.Keys, albySvc alby.AlbyService, albyOAuthSvc alby.AlbyOAuthService, eventPublisher events.EventPublisher) *api {
	return &api{
		db:             gormDB,
		appsSvc:        apps.NewAppsService(gormDB, eventPublisher, keys, config),
		cfg:            config,
		svc:            svc,
		permissionsSvc: permissions.NewPermissionsService(gormDB, eventPublisher),
		keys:           keys,
		albySvc:        albySvc,
		albyOAuthSvc:   albyOAuthSvc,
		eventPublisher: eventPublisher,
	}
}

func (api *api) CreateApp(createAppRequest *CreateAppRequest) (*CreateAppResponse, error) {
	if slices.Contains(createAppRequest.Scopes, constants.SUPERUSER_SCOPE) {
		if !api.cfg.CheckUnlockPassword(createAppRequest.UnlockPassword) {
			return nil, fmt.Errorf(
				"incorrect unlock password to create app with superuser permission")
		}
	}

	if createAppRequest.Name == alby.ALBY_ACCOUNT_APP_NAME {
		return nil, fmt.Errorf("Reserved app name: %s", alby.ALBY_ACCOUNT_APP_NAME)
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

	relayUrls := api.cfg.GetRelayUrls()

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
	responseBody.RelayUrls = relayUrls
	responseBody.Lud16 = lightningAddress

	if createAppRequest.ReturnTo != "" {
		returnToUrl, err := url.Parse(createAppRequest.ReturnTo)
		if err == nil {
			query := returnToUrl.Query()
			for _, relayUrl := range relayUrls {
				query.Add("relay", relayUrl)
			}
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
	responseBody.PairingUri = fmt.Sprintf("nostr+walletconnect://%s?relay=%s&secret=%s%s", *app.WalletPubkey, strings.Join(relayUrls, "&relay="), pairingSecretKey, lud16)

	return responseBody, nil
}

func (api *api) UpdateApp(userApp *db.App, updateAppRequest *UpdateAppRequest) error {
	err := api.db.Transaction(func(tx *gorm.DB) error {
		// Initialize name with current app name, update if provided
		name := userApp.Name

		// Update app name if provided and different
		if updateAppRequest.Name != nil {
			name = *updateAppRequest.Name

			if name == "" {
				return fmt.Errorf("won't update an app to have no name")
			}
			if name != userApp.Name {
				err := tx.Model(&db.App{}).Where("id", userApp.ID).Update("name", name).Error
				if err != nil {
					return err
				}
			}
		}

		// Update app isolation if provided and different
		if updateAppRequest.Isolated != nil {
			isolated := *updateAppRequest.Isolated
			if isolated != userApp.Isolated {
				if !isolated {
					var existingMetadata Metadata
					if userApp.Metadata != nil {
						err := json.Unmarshal(userApp.Metadata, &existingMetadata)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"app_id": userApp.ID,
							}).Error("Failed to deserialize app metadata")
							return err
						}
						if existingMetadata["app_store_app_id"] == constants.SUBWALLET_APPSTORE_APP_ID {
							return errors.New("Cannot update sub-wallet to be non-isolated")
						}
					}
				}

				err := tx.Model(&db.App{}).Where("id", userApp.ID).Update("isolated", isolated).Error
				if err != nil {
					return err
				}
			}
		}

		// Update the app metadata if provided
		if updateAppRequest.Metadata != nil {
			var metadataBytes []byte
			var err error
			metadataBytes, err = json.Marshal(*updateAppRequest.Metadata)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to serialize metadata")
				return err
			}
			err = tx.Model(&db.App{}).Where("id", userApp.ID).Update("metadata", datatypes.JSON(metadataBytes)).Error
			if err != nil {
				return err
			}
		}

		// Handle permissions updates only if any permission-related field is provided
		if updateAppRequest.Scopes != nil || updateAppRequest.MaxAmountSat != nil ||
			updateAppRequest.BudgetRenewal != nil || updateAppRequest.ExpiresAt != nil || updateAppRequest.UpdateExpiresAt {

			// Get current values or use provided ones
			var maxAmount uint64
			var budgetRenewal string
			var expiresAt *time.Time

			// Get existing permissions to use as defaults
			var existingPermissions []db.AppPermission
			if err := tx.Where("app_id = ?", userApp.ID).Find(&existingPermissions).Error; err != nil {
				return err
			}

			// Use existing values as defaults
			if len(existingPermissions) > 0 {
				// Find pay_invoice permission for budget-related fields
				for _, perm := range existingPermissions {
					if perm.Scope == constants.PAY_INVOICE_SCOPE {
						maxAmount = uint64(perm.MaxAmountSat)
						budgetRenewal = perm.BudgetRenewal
						expiresAt = perm.ExpiresAt
						break
					}
				}
			}

			// Override with provided values
			if updateAppRequest.MaxAmountSat != nil {
				maxAmount = *updateAppRequest.MaxAmountSat
			}
			if updateAppRequest.BudgetRenewal != nil {
				budgetRenewal = *updateAppRequest.BudgetRenewal
			}
			if updateAppRequest.ExpiresAt != nil {
				parsedExpiresAt, err := api.parseExpiresAt(*updateAppRequest.ExpiresAt)
				if err != nil {
					return fmt.Errorf("invalid expiresAt: %v", err)
				}
				expiresAt = parsedExpiresAt
			}
			if updateAppRequest.ExpiresAt == nil && updateAppRequest.UpdateExpiresAt {
				expiresAt = nil
			}

			// Update existing permissions with new budget and expiry
			err := tx.Model(&db.AppPermission{}).Where("app_id", userApp.ID).Updates(map[string]interface{}{
				"ExpiresAt":     expiresAt,
				"MaxAmountSat":  maxAmount,
				"BudgetRenewal": budgetRenewal,
			}).Error
			if err != nil {
				return err
			}

			// Handle scope changes only if scopes were provided
			if updateAppRequest.Scopes != nil {

				if len(updateAppRequest.Scopes) == 0 {
					return fmt.Errorf("won't update an app to have no request methods")
				}

				existingScopeMap := make(map[string]bool)
				for _, perm := range existingPermissions {
					existingScopeMap[perm.Scope] = true
				}

				if slices.Contains(updateAppRequest.Scopes, constants.SUPERUSER_SCOPE) && !existingScopeMap[constants.SUPERUSER_SCOPE] {
					return fmt.Errorf("cannot update app to add superuser permission")
				}

				// Add new permissions
				for _, scope := range updateAppRequest.Scopes {
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
			}
		}

		// Publish update event
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
	// Delete lightning address if one exists
	if api.appsSvc.HasLightningAddress(userApp) {
		err := api.DeleteLightningAddress(context.Background(), userApp.ID)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"app_id": userApp.ID,
			}).Error("Failed to delete lightning address during app deletion")
		}
	}

	return api.appsSvc.DeleteApp(userApp)
}

func (api *api) CreateLightningAddress(ctx context.Context, createLightningAddressRequest *CreateLightningAddressRequest) error {
	app := api.appsSvc.GetAppById(createLightningAddressRequest.AppId)
	if app == nil {
		return errors.New("app not found")
	}

	var metadata map[string]interface{}
	err := json.Unmarshal(app.Metadata, &metadata)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"app_id": app.ID,
		}).Error("Failed to deserialize app metadata")
		return err
	}

	createLightningAddressResponse, err := api.albyOAuthSvc.CreateLightningAddress(ctx, createLightningAddressRequest.Address, createLightningAddressRequest.AppId)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create lightning address for app")
		return err
	}

	metadata["lud16"] = createLightningAddressResponse.FullAddress
	err = api.appsSvc.SetAppMetadata(app.ID, metadata)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to add lightning address to app metadata")
		return err
	}
	return nil
}

func (api *api) DeleteLightningAddress(ctx context.Context, appId uint) error {
	app := api.appsSvc.GetAppById(appId)
	if app == nil {
		return errors.New("app not found")
	}

	var metadata map[string]interface{}
	err := json.Unmarshal(app.Metadata, &metadata)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"app_id": app.ID,
		}).Error("Failed to deserialize app metadata")
		return err
	}

	if metadata["lud16"] == nil {
		return errors.New("no lightning address set")
	}

	lud16 := metadata["lud16"].(string)
	if !strings.Contains(lud16, "@") {
		return errors.New("invalid lightning address")
	}
	address := strings.Split(lud16, "@")[0]

	// Call the Alby OAuth service to delete the lightning address
	err = api.albyOAuthSvc.DeleteLightningAddress(ctx, address)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to delete lightning address for app")
		return err
	}

	delete(metadata, "lud16")
	err = api.appsSvc.SetAppMetadata(app.ID, metadata)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove lightning address from app metadata")
		return err
	}

	return nil
}

func (api *api) GetApp(dbApp *db.App) (*App, error) {

	paySpecificPermission := db.AppPermission{}
	appPermissions := []db.AppPermission{}
	var expiresAt *time.Time
	if err := api.db.Where("app_id = ?", dbApp.ID).Find(&appPermissions).Error; err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"app_id": dbApp.ID,
		}).Error("Failed to list app permissions")
		return nil, err
	}

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
	maxAmount := uint64(paySpecificPermission.MaxAmountSat)
	budgetUsage, err := queries.GetBudgetUsageSat(api.db, &paySpecificPermission)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"app_id": dbApp.ID,
		}).Error("Failed to get budget usage for app")
		return nil, err
	}

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
		LastUsedAt:         dbApp.LastUsedAt,
	}

	if dbApp.Isolated {
		balance, err := queries.GetIsolatedBalance(api.db, dbApp.ID)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"app_id": dbApp.ID,
			}).Error("Failed to get isolated app balance")
			return nil, err
		}
		response.Balance = balance
	}

	return &response, nil
}

func (api *api) ListApps(limit uint64, offset uint64, filters ListAppsFilters, orderBy string) (*ListAppsResponse, error) {
	// TODO: join dbApps and permissions
	dbApps := []db.App{}
	query := api.db

	if filters.Name != "" {
		// searching for "Damus" will return "Damus" and "Damus (1)"
		// Use case-insensitive search for both SQLite and PostgreSQL
		if api.db.Dialector.Name() == "postgres" {
			query = query.Where("name ILIKE ?", filters.Name+"%")
		} else {
			query = query.Where("name LIKE ?", filters.Name+"%")
		}
	}

	if filters.AppStoreAppId != "" {
		query = query.Where(datatypes.JSONQuery("metadata").Equals(filters.AppStoreAppId, "app_store_app_id"))
	}

	if filters.Unused {
		// find unused non-subwallet apps not used in the past 60 days
		query = query.Where("last_used_at IS NULL OR last_used_at < ?", time.Now().Add(-60*24*time.Hour))
	}

	if filters.SubWallets != nil && !*filters.SubWallets {
		// exclude subwallets :scream:
		if api.db.Dialector.Name() == "sqlite" {
			query = query.Where("metadata is NULL OR JSON_EXTRACT(metadata, '$.app_store_app_id') IS NULL OR JSON_EXTRACT(metadata, '$.app_store_app_id') != ?", constants.SUBWALLET_APPSTORE_APP_ID)
		} else {
			query = query.Where("metadata IS NULL OR metadata->>'app_store_app_id' IS NULL OR metadata->>'app_store_app_id' != ?", constants.SUBWALLET_APPSTORE_APP_ID)
		}
	}

	if orderBy == "" {
		orderBy = "last_used_at"
	}
	if orderBy == "last_used_at" {
		// when ordering by last used at, apps with last_used_at is NULL should be ordered last
		orderBy = "last_used_at IS NULL, " + orderBy
	}

	query = query.Order(orderBy + " DESC")

	if limit == 0 {
		limit = 100
	}
	var totalCount int64
	result := query.Model(&db.App{}).Count(&totalCount)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to count DB apps")
		return nil, result.Error
	}
	query = query.Offset(int(offset)).Limit(int(limit))

	err := query.Find(&dbApps).Error

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list apps")
		return nil, err
	}

	appIds := []uint64{}
	for _, app := range dbApps {
		appIds = append(appIds, uint64(app.ID))
	}

	appPermissions := []db.AppPermission{}
	err = api.db.Where("app_id IN ?", appIds).Find(&appPermissions).Error
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
			LastUsedAt:         dbApp.LastUsedAt,
		}

		if dbApp.Isolated {
			balance, err := queries.GetIsolatedBalance(api.db, dbApp.ID)
			if err != nil {
				logger.Logger.WithError(err).WithFields(logrus.Fields{
					"app_id": dbApp.ID,
				}).Error("Failed to get isolated app balance")
				return nil, err
			}
			apiApp.Balance = balance
		}

		for _, appPermission := range permissionsMap[dbApp.ID] {
			apiApp.Scopes = append(apiApp.Scopes, appPermission.Scope)
			apiApp.ExpiresAt = appPermission.ExpiresAt
			if appPermission.Scope == constants.PAY_INVOICE_SCOPE {
				apiApp.BudgetRenewal = appPermission.BudgetRenewal
				apiApp.MaxAmountSat = uint64(appPermission.MaxAmountSat)
				budgetUsage, err := queries.GetBudgetUsageSat(api.db, &appPermission)
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"app_id": dbApp.ID,
					}).Error("Failed to get budget usage for app")
					return nil, err
				}
				apiApp.BudgetUsage = budgetUsage
			}
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
	return &ListAppsResponse{
		Apps:       apiApps,
		TotalCount: uint64(totalCount),
	}, nil
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
			ForwardingFeeProportionalMillionths:      channel.ForwardingFeeProportionalMillionths,
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
	return api.albySvc.GetChannelPeerSuggestions(ctx)
}

func (api *api) GetLSPChannelOffer(ctx context.Context) (*alby.LSPChannelOffer, error) {
	return api.albyOAuthSvc.GetLSPChannelOffer(ctx)
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
		return errors.New("please disable auto-unlock before using this feature")
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

func (api *api) RefundSwap(refundSwapRequest *RefundSwapRequest) error {
	if api.svc.GetSwapsService() == nil {
		return errors.New("SwapsService not started")
	}
	return api.svc.GetSwapsService().RefundSwap(refundSwapRequest.SwapId, refundSwapRequest.Address, false)
}

func (api *api) GetAutoSwapConfig() (*GetAutoSwapConfigResponse, error) {
	swapOutBalanceThresholdStr, _ := api.cfg.Get(config.AutoSwapBalanceThresholdKey, "")
	swapOutAmountStr, _ := api.cfg.Get(config.AutoSwapAmountKey, "")
	swapOutDestination, _ := api.cfg.Get(config.AutoSwapDestinationKey, "")

	swapOutEnabled := swapOutBalanceThresholdStr != "" && swapOutAmountStr != ""
	var swapOutBalanceThreshold, swapOutAmount uint64
	if swapOutEnabled {
		var err error
		if swapOutBalanceThreshold, err = strconv.ParseUint(swapOutBalanceThresholdStr, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid autoswap out balance threshold: %w", err)
		}
		if swapOutAmount, err = strconv.ParseUint(swapOutAmountStr, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid autoswap out amount: %w", err)
		}
	}

	return &GetAutoSwapConfigResponse{
		Type:             constants.SWAP_TYPE_OUT,
		Enabled:          swapOutEnabled,
		BalanceThreshold: swapOutBalanceThreshold,
		SwapAmount:       swapOutAmount,
		Destination:      swapOutDestination,
	}, nil
}

func (api *api) LookupSwap(swapId string) (*LookupSwapResponse, error) {
	if api.svc.GetSwapsService() == nil {
		return nil, errors.New("SwapsService not started")
	}
	dbSwap, err := api.svc.GetSwapsService().GetSwap(swapId)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to fetch swap info")
		return nil, err
	}

	return toApiSwap(dbSwap), nil
}

func (api *api) ListSwaps() (*ListSwapsResponse, error) {
	if api.svc.GetSwapsService() == nil {
		return nil, errors.New("SwapsService not started")
	}
	swaps, err := api.svc.GetSwapsService().ListSwaps()
	if err != nil {
		return nil, err
	}

	apiSwaps := []Swap{}
	for _, swap := range swaps {
		apiSwaps = append(apiSwaps, *toApiSwap(&swap))
	}

	return &ListSwapsResponse{
		Swaps: apiSwaps,
	}, nil
}

func toApiSwap(swap *swaps.Swap) *Swap {
	return &Swap{
		Id:                 swap.SwapId,
		Type:               swap.Type,
		State:              swap.State,
		Invoice:            swap.Invoice,
		SendAmount:         swap.SendAmount,
		ReceiveAmount:      swap.ReceiveAmount,
		PaymentHash:        swap.PaymentHash,
		DestinationAddress: swap.DestinationAddress,
		RefundAddress:      swap.RefundAddress,
		LockupAddress:      swap.LockupAddress,
		LockupTxId:         swap.LockupTxId,
		ClaimTxId:          swap.ClaimTxId,
		AutoSwap:           swap.AutoSwap,
		BoltzPubkey:        swap.BoltzPubkey,
		CreatedAt:          swap.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          swap.UpdatedAt.Format(time.RFC3339),
		UsedXpub:           swap.UsedXpub,
	}
}

func (api *api) GetSwapInInfo() (*SwapInfoResponse, error) {
	if api.svc.GetSwapsService() == nil {
		return nil, errors.New("SwapsService not started")
	}
	swapInInfo, err := api.svc.GetSwapsService().GetSwapInInfo()
	if err != nil {
		logger.Logger.WithError(err).Error("failed to calculate fee info")
		return nil, err
	}

	return &SwapInfoResponse{
		AlbyServiceFee:  swapInInfo.AlbyServiceFee,
		BoltzServiceFee: swapInInfo.BoltzServiceFee,
		BoltzNetworkFee: swapInInfo.BoltzNetworkFee,
		MinAmount:       swapInInfo.MinAmount,
		MaxAmount:       swapInInfo.MaxAmount,
	}, nil
}

func (api *api) GetSwapOutInfo() (*SwapInfoResponse, error) {
	if api.svc.GetSwapsService() == nil {
		return nil, errors.New("SwapsService not started")
	}
	swapOutInfo, err := api.svc.GetSwapsService().GetSwapOutInfo()
	if err != nil {
		logger.Logger.WithError(err).Error("failed to calculate fee info")
		return nil, err
	}

	return &SwapInfoResponse{
		AlbyServiceFee:  swapOutInfo.AlbyServiceFee,
		BoltzServiceFee: swapOutInfo.BoltzServiceFee,
		BoltzNetworkFee: swapOutInfo.BoltzNetworkFee,
		MinAmount:       swapOutInfo.MinAmount,
		MaxAmount:       swapOutInfo.MaxAmount,
	}, nil
}

func (api *api) InitiateSwapOut(ctx context.Context, initiateSwapOutRequest *InitiateSwapRequest) (*swaps.SwapResponse, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	if api.svc.GetSwapsService() == nil {
		return nil, errors.New("SwapsService not started")
	}

	amount := initiateSwapOutRequest.SwapAmount
	destination := initiateSwapOutRequest.Destination

	if amount == 0 {
		return nil, errors.New("invalid swap amount")
	}

	swapOutResponse, err := api.svc.GetSwapsService().SwapOut(amount, destination, false, false)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"amount":      amount,
			"destination": destination,
		}).WithError(err).Error("Failed to initiate swap out")
		return nil, err
	}

	return swapOutResponse, nil
}

func (api *api) InitiateSwapIn(ctx context.Context, initiateSwapInRequest *InitiateSwapRequest) (*swaps.SwapResponse, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	if api.svc.GetSwapsService() == nil {
		return nil, errors.New("SwapsService not started")
	}

	amount := initiateSwapInRequest.SwapAmount

	if amount == 0 {
		return nil, errors.New("invalid swap amount")
	}

	swapInResponse, err := api.svc.GetSwapsService().SwapIn(amount, false)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"amount": amount,
		}).WithError(err).Error("Failed to initiate swap in")
		return nil, err
	}

	return swapInResponse, nil
}

func (api *api) EnableAutoSwapOut(ctx context.Context, enableAutoSwapsRequest *EnableAutoSwapRequest) error {
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

	if api.svc.GetSwapsService() == nil {
		return errors.New("SwapsService not started")
	}
	return api.svc.GetSwapsService().EnableAutoSwapOut()
}

func (api *api) DisableAutoSwap() error {
	keys := []string{config.AutoSwapBalanceThresholdKey, config.AutoSwapAmountKey, config.AutoSwapDestinationKey}

	for _, key := range keys {
		if err := api.cfg.SetUpdate(key, "", ""); err != nil {
			logger.Logger.WithError(err).Errorf("Failed to remove autoswap config for key: %s", key)
			return err
		}
	}

	if api.svc.GetSwapsService() != nil {
		api.svc.GetSwapsService().StopAutoSwapOut()
	}
	return nil
}

func (api *api) GetSwapMnemonic() string {
	return api.keys.GetSwapMnemonic()
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
		response, err := api.RequestEsploraApi(ctx, "/address/"+currentAddress+"/txs")
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
func (api *api) RequestMempoolApi(ctx context.Context, endpoint string) (interface{}, error) {
	url := api.cfg.GetEnv().MempoolApi + endpoint

	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	setupCompleted, err := api.cfg.SetupCompleted()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to check if setup is completed")
		return nil, err
	}
	info.SetupCompleted = setupCompleted
	info.Currency = api.cfg.GetCurrency()
	info.BitcoinDisplayFormat = api.cfg.GetBitcoinDisplayFormat()
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
	info.HideUpdateBanner = api.cfg.GetEnv().HideUpdateBanner
	info.LdkVssEnabled = ldkVssEnabled == "true"
	info.VssSupported = backendType == config.LDKBackendType && api.cfg.GetEnv().LDKVssUrl != ""
	info.AutoUnlockPasswordEnabled = autoUnlockPassword != ""
	info.AutoUnlockPasswordSupported = api.cfg.GetEnv().IsDefaultClientId()
	info.Relays = []InfoResponseRelay{}
	for _, relayStatus := range api.svc.GetRelayStatuses() {
		info.Relays = append(info.Relays, InfoResponseRelay{
			Url:    relayStatus.Url,
			Online: relayStatus.Online,
		})
	}

	info.MempoolUrl = api.cfg.GetMempoolUrl()
	info.AlbyAccountConnected = api.albyOAuthSvc.IsConnected(ctx)

	albyUserIdentifier, err := api.albyOAuthSvc.GetUserIdentifier()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get alby user identifier")
		return nil, err
	}
	info.AlbyUserIdentifier = albyUserIdentifier

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

func (api *api) SetBitcoinDisplayFormat(format string) error {
	if format != constants.BITCOIN_DISPLAY_FORMAT_SATS && format != constants.BITCOIN_DISPLAY_FORMAT_BIP177 {
		return fmt.Errorf("bitcoin display format must be '%s' or '%s'", constants.BITCOIN_DISPLAY_FORMAT_SATS, constants.BITCOIN_DISPLAY_FORMAT_BIP177)
	}

	err := api.cfg.SetBitcoinDisplayFormat(format)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to update bitcoin display format")
		return err
	}

	return nil
}

func (api *api) UpdateSettings(updateSettingsRequest *UpdateSettingsRequest) error {
	if updateSettingsRequest.Currency != "" {
		err := api.SetCurrency(updateSettingsRequest.Currency)
		if err != nil {
			return fmt.Errorf("failed to set currency: %w", err)
		}
	}

	if updateSettingsRequest.BitcoinDisplayFormat != "" {
		err := api.SetBitcoinDisplayFormat(updateSettingsRequest.BitcoinDisplayFormat)
		if err != nil {
			return fmt.Errorf("failed to set bitcoin display format: %w", err)
		}
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
	err := api.startInternal(startRequest)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to start node")
		api.startupError = err
		api.startupErrorTime = time.Now()
	}
}

func (api *api) startInternal(startRequest *StartRequest) (err error) {
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
		return fmt.Errorf("migration type not supported: %s", to)
	}

	ldkVssEnabled, err := api.cfg.Get("LdkVssEnabled", "")
	if err != nil {
		return err
	}

	if ldkVssEnabled == "true" {
		return errors.New("VSS already enabled")
	}

	if api.cfg.GetEnv().LDKVssUrl == "" {
		return errors.New("no VSS URL set")
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

	albyInfo, err := api.albySvc.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !albyInfo.Healthy {
		alarms = append(alarms, NewHealthAlarm(HealthAlarmKindAlbyService, albyInfo.Incidents))
	}

	relayStatuses := api.svc.GetRelayStatuses()
	if len(relayStatuses) > 0 {
		isAnyNostrRelayOffline := false
		offlineRelayUrls := []string{}
		for _, relayStatus := range relayStatuses {
			if !relayStatus.Online {
				isAnyNostrRelayOffline = true
				offlineRelayUrls = append(offlineRelayUrls, relayStatus.Url)
			}
		}
		if isAnyNostrRelayOffline {
			alarms = append(alarms, NewHealthAlarm(HealthAlarmKindNostrRelayOffline, offlineRelayUrls))
		}
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

func (api *api) SendEvent(event string, properties interface{}) {
	api.svc.GetEventPublisher().Publish(&events.Event{
		Event:      event,
		Properties: properties,
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

func (api *api) GetForwards() (*GetForwardsResponse, error) {
	var forwards []db.Forward
	err := api.db.Find(&forwards).Error
	if err != nil {
		return nil, err
	}

	var totalOutboundAmount uint64
	var totalFeeEarned uint64

	for _, forward := range forwards {
		totalOutboundAmount += forward.OutboundAmountForwardedMsat
		totalFeeEarned += forward.TotalFeeEarnedMsat
	}

	numForwards := len(forwards)

	return &GetForwardsResponse{
		OutboundAmountForwardedMsat: totalOutboundAmount,
		TotalFeeEarnedMsat:          totalFeeEarned,
		NumForwards:                 uint64(numForwards),
	}, nil
}

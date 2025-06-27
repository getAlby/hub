package apps

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service/keys"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AppsService interface {
	CreateApp(name string, pubkey string, maxAmountSat uint64, budgetRenewal string, expiresAt *time.Time, scopes []string, isolated bool, metadata map[string]interface{}) (*db.App, string, error)
	DeleteApp(app *db.App) error
	GetAppByPubkey(pubkey string) *db.App
	GetAppById(id uint) *db.App
	SetAppMetadata(appId uint, metadata map[string]interface{}) error
}

type appsService struct {
	db             *gorm.DB
	eventPublisher events.EventPublisher
	keys           keys.Keys
	cfg            config.Config
}

func NewAppsService(db *gorm.DB, eventPublisher events.EventPublisher, keys keys.Keys, cfg config.Config) *appsService {
	return &appsService{
		db:             db,
		eventPublisher: eventPublisher,
		keys:           keys,
		cfg:            cfg,
	}
}

func (svc *appsService) CreateApp(name string, pubkey string, maxAmountSat uint64, budgetRenewal string, expiresAt *time.Time, scopes []string, isolated bool, metadata map[string]interface{}) (*db.App, string, error) {
	if isolated {
		if slices.Contains(scopes, constants.SIGN_MESSAGE_SCOPE) {
			// cannot sign messages because the isolated app is a custodial sub-wallet
			return nil, "", errors.New("Sub-wallet app connection cannot have sign_message scope")
		}

		backendType, _ := svc.cfg.Get("LNBackendType", "")
		if backendType != config.LDKBackendType &&
			backendType != config.LNDBackendType &&
			backendType != config.PhoenixBackendType {
			return nil, "", fmt.Errorf(
				"sub-wallets are currently not supported on your node backend. Try LDK or LND")
		}
	}

	if budgetRenewal == "" {
		budgetRenewal = constants.BUDGET_RENEWAL_NEVER
	}

	if !slices.Contains(constants.GetBudgetRenewals(), budgetRenewal) {
		return nil, "", fmt.Errorf("invalid budget renewal. Must be one of %s", strings.Join(constants.GetBudgetRenewals(), ","))
	}

	// ensure there is at least one scope
	if scopes == nil || len(scopes) == 0 {
		return nil, "", errors.New("no scopes provided")
	}

	var pairingPublicKey string
	var pairingSecretKey string
	if pubkey == "" {
		pairingSecretKey = nostr.GeneratePrivateKey()
		pairingPublicKey, _ = nostr.GetPublicKey(pairingSecretKey)
	} else {
		pairingPublicKey = pubkey
		//validate public key
		decoded, err := hex.DecodeString(pairingPublicKey)
		if err != nil || len(decoded) != 32 {
			logger.Logger.WithField("pairingPublicKey", pairingPublicKey).Error("Invalid public key format")
			return nil, "", fmt.Errorf("invalid public key format: %s", pairingPublicKey)
		}
	}

	var metadataBytes []byte
	if metadata != nil {
		var err error
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to serialize metadata")
			return nil, "", err
		}
	}

	// use a suffix to avoid duplicate names
	nameIndex := 0
	var freeName string
	for ; ; nameIndex++ {
		freeName = name
		if nameIndex > 0 {
			freeName += fmt.Sprintf(" (%d)", nameIndex)
		}
		existingApp := svc.GetAppByName(freeName)
		if existingApp == nil {
			break
		}
	}

	app := db.App{Name: freeName, AppPubkey: pairingPublicKey, Isolated: isolated, Metadata: datatypes.JSON(metadataBytes)}

	err := svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&app).Error
		if err != nil {
			return err
		}

		for _, scope := range scopes {
			appPermission := db.AppPermission{
				App:       app,
				Scope:     scope,
				ExpiresAt: expiresAt,
				//these fields are only relevant for pay_invoice
				MaxAmountSat:  int(maxAmountSat),
				BudgetRenewal: budgetRenewal,
			}
			err = tx.Create(&appPermission).Error
			if err != nil {
				return err
			}
		}

		appWalletPrivKey, err := svc.keys.GetAppWalletKey(app.ID)
		if err != nil {
			return fmt.Errorf("error generating wallet child private key: %w", err)
		}

		appWalletPubkey, err := nostr.GetPublicKey(appWalletPrivKey)
		if err != nil {
			return fmt.Errorf("error generating wallet child public key: %w", err)
		}

		err = tx.Model(&app).Update("wallet_pubkey", appWalletPubkey).Error
		if err != nil {
			return err
		}

		// commit transaction
		return nil
	})

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save app")
		return nil, "", err
	}

	svc.eventPublisher.Publish(&events.Event{
		Event: "nwc_app_created",
		Properties: map[string]interface{}{
			"name": name,
			"id":   app.ID,
		},
	})

	return &app, pairingSecretKey, nil
}

func (svc *appsService) DeleteApp(app *db.App) error {

	err := svc.db.Delete(app).Error
	if err != nil {
		return err
	}
	svc.eventPublisher.Publish(&events.Event{
		Event: "nwc_app_deleted",
		Properties: map[string]interface{}{
			"name": app.Name,
			"id":   app.ID,
		},
	})
	return nil
}

func (svc *appsService) GetAppByPubkey(pubkey string) *db.App {
	dbApp := db.App{}
	findResult := svc.db.Where("app_pubkey = ?", pubkey).First(&dbApp)
	if findResult.RowsAffected == 0 {
		return nil
	}
	return &dbApp
}

func (svc *appsService) GetAppById(id uint) *db.App {
	dbApp := db.App{}
	findResult := svc.db.Where("id = ?", id).First(&dbApp)
	if findResult.RowsAffected == 0 {
		return nil
	}
	return &dbApp
}

func (svc *appsService) GetAppByName(name string) *db.App {
	dbApp := db.App{}
	findResult := svc.db.Where("name = ?", name).First(&dbApp)
	if findResult.RowsAffected == 0 {
		return nil
	}
	return &dbApp
}

func (svc *appsService) SetAppMetadata(id uint, metadata map[string]interface{}) error {
	var metadataBytes []byte
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to serialize metadata")
		return err
	}

	err = svc.db.Model(&db.App{}).Where("id", id).Update("metadata", datatypes.JSON(metadataBytes)).Error
	if err != nil {
		logger.Logger.WithError(err).WithField("metadata", metadata).Error("failed to update transaction metadata")
		return err
	}

	return nil
}

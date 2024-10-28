package apps

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AppsService interface {
	CreateApp(name string, pubkey string, maxAmountSat uint64, budgetRenewal string, expiresAt *time.Time, scopes []string, isolated bool, metadata map[string]interface{}) (*db.App, string, error)
	GetAppByPubkey(pubkey string) *db.App
}

type appsService struct {
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

func NewAppsService(db *gorm.DB, eventPublisher events.EventPublisher) *appsService {
	return &appsService{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (svc *appsService) CreateApp(name string, pubkey string, maxAmountSat uint64, budgetRenewal string, expiresAt *time.Time, scopes []string, isolated bool, metadata map[string]interface{}) (*db.App, string, error) {
	if isolated && (slices.Contains(scopes, constants.SIGN_MESSAGE_SCOPE)) {
		// cannot sign messages because the isolated app is a custodial subaccount
		return nil, "", errors.New("isolated app cannot have sign_message scope")
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

	app := db.App{Name: name, NostrPubkey: pairingPublicKey, Isolated: isolated, Metadata: datatypes.JSON(metadataBytes)}

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

		// commit transaction
		return nil
	})

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save app")
		return nil, "", err
	}

	svc.eventPublisher.Publish(&events.Event{
		Event: "app_created",
		Properties: map[string]interface{}{
			"name": name,
		},
	})

	return &app, pairingSecretKey, nil
}

func (svc *appsService) GetAppByPubkey(pubkey string) *db.App {
	dbApp := db.App{}
	findResult := svc.db.Where("nostr_pubkey = ?", pubkey).First(&dbApp)
	if findResult.RowsAffected == 0 {
		return nil
	}
	return &dbApp
}

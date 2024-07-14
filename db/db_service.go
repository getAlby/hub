package db

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"
)

type dbService struct {
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

func NewDBService(db *gorm.DB, eventPublisher events.EventPublisher) *dbService {
	return &dbService{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (svc *dbService) CreateApp(name string, pubkey string, maxAmountSat uint64, budgetRenewal string, expiresAt *time.Time, scopes []string) (*App, string, error) {
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

	app := App{Name: name, NostrPubkey: pairingPublicKey}

	err := svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&app).Error
		if err != nil {
			return err
		}

		for _, scope := range scopes {
			appPermission := AppPermission{
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

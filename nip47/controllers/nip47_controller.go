package controllers

import (
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/nip47/permissions"
	"github.com/getAlby/nostr-wallet-connect/transactions"
	"gorm.io/gorm"
)

type nip47Controller struct {
	lnClient            lnclient.LNClient
	db                  *gorm.DB
	eventPublisher      events.EventPublisher
	permissionsService  permissions.PermissionsService
	transactionsService transactions.TransactionsService
}

func NewNip47Controller(lnClient lnclient.LNClient, db *gorm.DB, eventPublisher events.EventPublisher, permissionsService permissions.PermissionsService, transactionsService transactions.TransactionsService) *nip47Controller {
	return &nip47Controller{
		lnClient:            lnClient,
		db:                  db,
		eventPublisher:      eventPublisher,
		permissionsService:  permissionsService,
		transactionsService: transactionsService,
	}
}

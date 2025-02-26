package controllers

import (
	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/transactions"
	"gorm.io/gorm"
)

type nip47Controller struct {
	lnClient            lnclient.LNClient
	db                  *gorm.DB
	eventPublisher      events.EventPublisher
	permissionsService  permissions.PermissionsService
	transactionsService transactions.TransactionsService
	appsService         apps.AppsService
	albyOAuthService    alby.AlbyOAuthService
}

func NewNip47Controller(
	lnClient lnclient.LNClient,
	db *gorm.DB,
	eventPublisher events.EventPublisher,
	permissionsService permissions.PermissionsService,
	transactionsService transactions.TransactionsService,
	appsService apps.AppsService,
	albyOAuthService alby.AlbyOAuthService) *nip47Controller {
	return &nip47Controller{
		lnClient:            lnClient,
		db:                  db,
		eventPublisher:      eventPublisher,
		permissionsService:  permissionsService,
		transactionsService: transactionsService,
		appsService:         appsService,
		albyOAuthService:    albyOAuthService,
	}
}

package permissions

import (
	"testing"
	"time"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/getAlby/nostr-wallet-connect/tests"
	"github.com/stretchr/testify/assert"
)

func TestHasPermission_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, PAY_INVOICE_SCOPE, 100)
	assert.False(t, result)
	assert.Equal(t, models.ERROR_RESTRICTED, code)
	assert.Equal(t, "This app does not have the pay_invoice scope", message)
}

func TestHasPermission_Expired(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(-24 * time.Hour)
	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         PAY_INVOICE_SCOPE,
		MaxAmount:     100,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     &expiresAt,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, PAY_INVOICE_SCOPE, 100)
	assert.False(t, result)
	assert.Equal(t, models.ERROR_EXPIRED, code)
	assert.Equal(t, "This app has expired", message)
}

func TestHasPermission_Exceeded(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         models.PAY_INVOICE_METHOD,
		MaxAmount:     10,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     &expiresAt,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, PAY_INVOICE_SCOPE, 100*1000)
	assert.False(t, result)
	assert.Equal(t, models.ERROR_QUOTA_EXCEEDED, code)
	assert.Equal(t, "Insufficient budget remaining to make payment", message)
}

func TestHasPermission_OK(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         models.PAY_INVOICE_METHOD,
		MaxAmount:     10,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     &expiresAt,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, models.PAY_INVOICE_METHOD, 10*1000)
	assert.True(t, result)
	assert.Empty(t, code)
	assert.Empty(t, message)
}

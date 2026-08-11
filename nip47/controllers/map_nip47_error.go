package controllers

import (
	"errors"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/transactions"
	"gorm.io/gorm"
)

func mapNip47Error(err error) *models.Error {
	code := constants.ERROR_INTERNAL
	message := err.Error()
	if errors.Is(err, transactions.NewNotFoundError()) {
		code = constants.ERROR_NOT_FOUND
	}
	if errors.Is(err, transactions.NewInsufficientBalanceError()) {
		code = constants.ERROR_INSUFFICIENT_BALANCE
	}
	if errors.Is(err, transactions.NewQuotaExceededError()) {
		code = constants.ERROR_QUOTA_EXCEEDED
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		// avoid leaking raw driver/SQL error details (e.g. constraint names) to NWC apps
		message = gorm.ErrDuplicatedKey.Error()
	}

	return &models.Error{
		Code:    code,
		Message: message,
	}
}

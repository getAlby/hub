package controllers

import (
	"errors"

	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/transactions"
)

func mapNip47Error(err error) *models.Error {
	code := models.ERROR_INTERNAL
	if errors.Is(err, transactions.NewNotFoundError()) {
		code = models.ERROR_NOT_FOUND
	}
	if errors.Is(err, transactions.NewInsufficientBalanceError()) {
		code = models.ERROR_INSUFFICIENT_BALANCE
	}
	if errors.Is(err, transactions.NewQuotaExceededError()) {
		code = models.ERROR_QUOTA_EXCEEDED
	}

	return &models.Error{
		Code:    code,
		Message: err.Error(),
	}
}

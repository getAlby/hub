package controllers

import (
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
)

type checkPermissionFunc = func(amountMsat uint64) *models.Response
type publishFunc = func(*models.Response, nostr.Tags)

type payResponse struct {
	Preimage string  `json:"preimage"`
	FeesPaid *uint64 `json:"fees_paid"`
}

package controllers

import (
	"github.com/getAlby/go-nostr"
	"github.com/getAlby/hub/nip47/models"
)

type publishFunc = func(*models.Response, nostr.Tags)

type payResponse struct {
	Preimage string `json:"preimage"`
	FeesPaid uint64 `json:"fees_paid"`
}

package main

import (
	"errors"

	"github.com/getAlby/nostr-wallet-connect/models/api"
)

func (svc *Service) ListApps(userApps *[]App, apps *[]api.App) error {
	for _, app := range *userApps {
		apiApp := api.App{
			// ID:          app.ID,
			Name:        app.Name,
			Description: app.Description,
			CreatedAt:   app.CreatedAt,
			UpdatedAt:   app.UpdatedAt,
			NostrPubkey: app.NostrPubkey,
		}

		var lastEvent NostrEvent
		result := svc.db.Where("app_id = ?", app.ID).Order("id desc").Limit(1).Find(&lastEvent)
		if result.Error != nil {
			svc.Logger.Errorf("Failed to fetch last event %v", result.Error)
			return errors.New("Failed to fetch last event")
		}
		if result.RowsAffected > 0 {
			apiApp.LastEventAt = &lastEvent.CreatedAt
		}
		*apps = append(*apps, apiApp)
	}
	return nil
}

func (svc *Service) GetInfo(info *api.InfoResponse) {
	info.BackendType = svc.cfg.LNBackendType
}

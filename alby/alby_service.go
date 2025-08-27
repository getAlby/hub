package alby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

const albyInternalAPIURL = "https://getalby.com/api"

type albyService struct {
	cfg config.Config
}

func NewAlbyService(cfg config.Config) *albyService {
	albySvc := &albyService{
		cfg: cfg,
	}
	return albySvc
}

func (svc *albyService) GetBitcoinRate(ctx context.Context) (*BitcoinRate, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	currency := svc.cfg.GetCurrency()

	url := fmt.Sprintf("%s/rates/%s", albyInternalAPIURL, currency)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"currency": currency,
			"error":    err,
		}).Error("Error creating request to Bitcoin rate endpoint")
		return nil, err
	}
	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"currency": currency,
			"error":    err,
		}).Error("Failed to fetch Bitcoin rate from API")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"currency":    currency,
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("Bitcoin rate endpoint returned non-success code")
		return nil, fmt.Errorf("bitcoin rate endpoint returned non-success code: %s", string(body))
	}

	var rate = &BitcoinRate{}
	err = json.Unmarshal(body, rate)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"currency": currency,
			"body":     string(body),
			"error":    err,
		}).Error("Failed to decode Bitcoin rate API response")
		return nil, err
	}

	return rate, nil
}

func (svc *albyService) GetChannelPeerSuggestions(ctx context.Context) ([]ChannelPeerSuggestion, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/channel_suggestions", albyInternalAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to channel_suggestions endpoint")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch channel_suggestions endpoint")
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("channel suggestions endpoint returned non-success code")
		return nil, fmt.Errorf("channel suggestions endpoint returned non-success code: %s", string(body))
	}

	var suggestions []ChannelPeerSuggestion
	err = json.Unmarshal(body, &suggestions)
	if err != nil {
		logger.Logger.WithError(err).Errorf("Failed to decode API response")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{"channel_suggestions": suggestions}).Debug("Alby channel peer suggestions response")
	return suggestions, nil
}

func (svc *albyService) GetInfo(ctx context.Context) (*AlbyInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/info", albyInternalAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to alby info endpoint")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch /info")
		return nil, err
	}

	type albyInfoHub struct {
		LatestVersion      string `json:"latest_version"`
		LatestReleaseNotes string `json:"latest_release_notes"`
	}

	type albyInfoIncident struct {
		Name    string `json:"name"`
		Started string `json:"started"`
		Status  string `json:"status"`
		Impact  string `json:"impact"`
		Url     string `json:"url"`
	}

	type albyInfo struct {
		Hub              albyInfoHub        `json:"hub"`
		Status           string             `json:"status"`
		Healthy          bool               `json:"healthy"`
		AccountAvailable bool               `json:"account_available"` // false if country is blocked (can still use Alby Hub without an Alby Account)
		Incidents        []albyInfoIncident `json:"incidents"`
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("info endpoint returned non-success code")
		return nil, fmt.Errorf("info endpoint returned non-success code: %s", string(body))
	}

	info := &albyInfo{}
	err = json.Unmarshal(body, info)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	incidents := []AlbyInfoIncident{}
	for _, incident := range info.Incidents {
		incidents = append(incidents, AlbyInfoIncident{
			Name:    incident.Name,
			Started: incident.Started,
			Status:  incident.Status,
			Impact:  incident.Impact,
			Url:     incident.Url,
		})
	}

	return &AlbyInfo{
		Hub: AlbyInfoHub{
			LatestVersion:      info.Hub.LatestVersion,
			LatestReleaseNotes: info.Hub.LatestReleaseNotes,
		},
		Status:           info.Status,
		Healthy:          info.Healthy,
		AccountAvailable: info.AccountAvailable,
		Incidents:        incidents,
	}, nil
}

package alby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

const albyInternalAPIURL = "https://getalby.com/api"

type albyService struct {
}

func NewAlbyService() *albyService {
	return &albyService{}
}

func (svc *albyService) GetCurrencies(ctx context.Context) ([]Currency, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/rates", albyInternalAPIURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to currencies endpoint")
		return nil, err
	}
	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch currencies from API")
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
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("Currencies endpoint returned non-success code")
		return nil, fmt.Errorf("currencies endpoint returned non-success code: %s", string(body))
	}

	rawCurrencies := map[string]Currency{}
	err = json.Unmarshal(body, &rawCurrencies)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"body":  string(body),
			"error": err,
		}).Error("Failed to decode currencies API response")
		return nil, err
	}

	currencies := []Currency{}
	for _, currency := range rawCurrencies {
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

func (svc *albyService) GetBitcoinRate(ctx context.Context, currency string) (*BitcoinRate, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/rates/%s", albyInternalAPIURL, currency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/internal/channel_suggestions", albyInternalAPIURL), nil)
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

	for i := range suggestions {
		suggestions[i].MinimumChannelSizeSat = suggestions[i].MinimumChannelSize
		suggestions[i].MaximumChannelSizeSat = suggestions[i].MaximumChannelSize
	}

	// TODO: remove before merging - injected test LSPS2 suggestions
	{
		maxExpiry := uint32(13140)
		suggestions = append(suggestions, ChannelPeerSuggestion{
			Network:                    "signet",
			PaymentMethod:              "lightning",
			Identifier:                 "alby",
			Type:                       "LSPS2",
			ContactUrl:                 "https://getalby.com",
			NodeAddress:                "025010bd608771bc13f08f696e3dd226bf3a9ae6ea461e3922ed9bdca7bb0edfe5@141.95.84.44:9735",
			MinimumChannelSize:         3000,
			MinimumChannelSizeSat:      3000,
			MaximumChannelSize:         16000000,
			MaximumChannelSizeSat:      16000000,
			MaximumChannelExpiryBlocks: &maxExpiry,
			Name:                       "Alby (Mutinynet)",
			Description:                "Alby Test LSPS2",
			PublicChannelsAllowed:      false,
			Terms:                      "For testing only",
		})
		suggestions = append(suggestions, ChannelPeerSuggestion{
			Network:                    "bitcoin",
			PaymentMethod:              "lightning",
			Identifier:                 "megalith",
			Type:                       "LSPS2",
			ContactUrl:                 "https://megalithic.me/contact",
			NodeAddress:                "03e30fda71887a916ef5548a4d02b06fe04aaa1a8de9e24134ce7f139cf79d7579@64.23.192.68:9736",
			MinimumChannelSize:         2501,
			MinimumChannelSizeSat:      2501,
			MaximumChannelSize:         16000000,
			MaximumChannelSizeSat:      16000000,
			MaximumChannelExpiryBlocks: &maxExpiry,
			Name:                       "Megalith 2",
			Description:                "Megalith is one of the biggest and most stable routing nodes on the Lightning Network.",
			PublicChannelsAllowed:      false,
			Terms:                      "Megalith will do its best to keep your channel open for at least 3 months. Your channel will stay open indefinitely if you use it regularly and keep your hub online.",
		})
	}

	logger.Logger.WithFields(logrus.Fields{"channel_suggestions": suggestions}).Debug("Alby channel peer suggestions response")
	return suggestions, nil
}

func (svc *albyService) GetInfo(ctx context.Context) (*AlbyInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/internal/info", albyInternalAPIURL), nil)
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

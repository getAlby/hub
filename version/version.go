package version

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

var Tag string = "Unknown Version"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

var latestRelease = ""
var lastVersionCheck = time.Time{}

func GetLatestReleaseTag() string {
	if latestRelease != "" && time.Since(lastVersionCheck) < 5*time.Minute {
		return latestRelease
	}
	url := "https://api.github.com/repos/getAlby/hub/releases"

	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create http request")
		return ""
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to send request")
		return ""
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return ""
	}

	releases := []githubRelease{}
	jsonErr := json.Unmarshal(body, &releases)
	if jsonErr != nil {
		logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return ""
	}

	if len(releases) < 1 {
		logger.Logger.Error("no github releases found")
		return ""
	}

	latestRelease = releases[0].TagName

	logger.Logger.WithFields(logrus.Fields{
		"latest":  latestRelease,
		"current": Tag,
	}).Info("Found latest github release")

	lastVersionCheck = time.Now()

	return latestRelease
}

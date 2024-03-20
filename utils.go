package main

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

func GetStartOfBudget(budget_type string, createdAt time.Time) time.Time {
	now := time.Now()
	switch budget_type {
	case "daily":
		// TODO: Use the location of the user, instead of the server
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "weekly":
		weekday := now.Weekday()
		var startOfWeek time.Time
		if weekday == 0 {
			startOfWeek = now.AddDate(0, 0, -6)
		} else {
			startOfWeek = now.AddDate(0, 0, -int(weekday)+1)
		}
		return time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	case "monthly":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "yearly":
		return time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	default: //"never"
		return createdAt
	}
}

func LoggedClose(logger *logrus.Logger, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		logger.WithError(err).Error("Close() failed")
	}
}

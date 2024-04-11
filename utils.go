package main

import (
	"fmt"
	"io"
	"os"
	"time"
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

func ReadFileTail(filePath string, maxLen int) (data []byte, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			err = fmt.Errorf("failed to close file: %w", err)
			data = nil
		}
	}()

	var dataReader io.Reader = f

	if maxLen > 0 {
		stat, err := f.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %w", err)
		}

		if stat.Size() > int64(maxLen) {
			_, err = f.Seek(-int64(maxLen), io.SeekEnd)
			if err != nil {
				return nil, fmt.Errorf("failed to seek file: %w", err)
			}
		}

		dataReader = io.LimitReader(f, int64(maxLen))
	}

	data, err = io.ReadAll(dataReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

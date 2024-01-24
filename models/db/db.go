package db

import "time"

type UserConfig struct {
	ID        uint
	Key       string
	Value     string
	Encrypted bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

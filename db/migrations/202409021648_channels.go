package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202409021648_channels = &gormigrate.Migration{
	ID: "202409021648_channels",
	Migrate: func(tx *gorm.DB) error {

		if err := tx.Exec(`
CREATE TABLE channels(
	id integer PRIMARY KEY AUTOINCREMENT,
	status text,
	created_at datetime,
	updated_at datetime,
	channel_id text,
	peer_id text,
	channel_size_sat integer,
	funding_tx_id text,
	open boolean
);

CREATE INDEX idx_channels_status ON channels(status);
CREATE INDEX idx_channels_created_at ON channels(created_at);
CREATE INDEX idx_channels_channel_id ON channels(channel_id);
CREATE INDEX idx_channels_peer_id ON channels(peer_id);
CREATE INDEX idx_channels_funding_tx_id ON channels(funding_tx_id);
`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}

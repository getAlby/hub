package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// This migration (inside a DB transaction),
// - Adds AUTOINCREMENT to the primary key of:
// - apps, app_permissions, request_events, response_events, user_configs
//
// request_events and response_events are not critical (and also payments are dropped in the same release)
// so we just drop those tables and re-create them.
var _202407151352_autoincrement = &gormigrate.Migration{
	ID: "202407151352_autoincrement",
	Migrate: func(db *gorm.DB) error {

		if err := db.Transaction(func(tx *gorm.DB) error {

			// drop old request and response event tables
			if err := tx.Exec(`
DROP TABLE request_events;
DROP TABLE response_events;
`).Error; err != nil {
				return err
			}

			// User configs - create new table, copy old values, delete old table, rename new table
			if err := tx.Exec(`
CREATE TABLE "user_configs_2" ("id" integer PRIMARY KEY AUTOINCREMENT, "key" text NOT NULL UNIQUE, "value" text, "encrypted" numeric, created_at datetime,updated_at datetime);
INSERT INTO user_configs_2 (id, key, value, encrypted, created_at, updated_at) SELECT id, key, value, encrypted, created_at, updated_at FROM user_configs;
DROP TABLE user_configs;
ALTER TABLE user_configs_2 RENAME TO user_configs;
`).Error; err != nil {
				return err
			}

			// Apps & app permissions (interdependent)
			// create new tables, copy old values, delete old tables, rename new tables, create new indexes
			if err := tx.Exec(`
CREATE TABLE apps_2 (id integer PRIMARY KEY AUTOINCREMENT,name text,description text,nostr_pubkey text UNIQUE,created_at datetime,updated_at datetime, isolated boolean);
INSERT INTO apps_2 (id, name, description, nostr_pubkey, created_at, updated_at, isolated) SELECT id, name text, description, nostr_pubkey, created_at, updated_at, isolated FROM apps;
CREATE TABLE app_permissions_2 (id integer PRIMARY KEY AUTOINCREMENT,app_id integer,"scope" text,"max_amount_sat" integer,budget_renewal text,expires_at datetime,created_at datetime,updated_at datetime,CONSTRAINT fk_app_permissions_app FOREIGN KEY (app_id) REFERENCES apps_2(id) ON DELETE CASCADE);
INSERT INTO app_permissions_2 (id, app_id, scope, max_amount_sat, budget_renewal, expires_at, created_at, updated_at) SELECT id, app_id, scope, max_amount_sat, budget_renewal, expires_at, created_at, updated_at FROM app_permissions;

DROP TABLE apps;
ALTER TABLE apps_2 RENAME TO apps;
DROP TABLE app_permissions;
ALTER TABLE app_permissions_2 RENAME TO app_permissions;

CREATE INDEX idx_app_permissions_scope ON app_permissions("scope");
CREATE INDEX idx_app_permissions_app_id ON app_permissions(app_id);
`).Error; err != nil {
				return err
			}

			// create fresh request and response event tables
			if err := tx.Exec(`
CREATE TABLE "request_events" (id integer PRIMARY KEY AUTOINCREMENT,app_id integer,nostr_id text UNIQUE,state text,created_at datetime,updated_at datetime, method TEXT, content_data TEXT,CONSTRAINT fk_request_events_app FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE);
CREATE INDEX idx_request_events_app_id ON request_events(app_id);
CREATE INDEX idx_request_events_app_id_and_id ON request_events(app_id, id);
CREATE INDEX idx_request_events_method ON request_events(method);
CREATE TABLE "response_events" (id integer PRIMARY KEY AUTOINCREMENT,nostr_id text UNIQUE,request_id integer,state text,replied_at datetime,created_at datetime,updated_at datetime,CONSTRAINT fk_response_events_request_event FOREIGN KEY (request_id) REFERENCES request_events(id) ON DELETE CASCADE);
`).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}

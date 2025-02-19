package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const dropMigration = `
DROP TABLE request_events {{ .DropTableCascade }};
DROP TABLE response_events;
`

var dropMigrationTmpl = template.Must(template.New("dropMigration").Parse(dropMigration))

const appsMigration = `
DELETE FROM app_permissions WHERE app_id NOT IN (SELECT id FROM apps);
CREATE TABLE apps_2 (id {{ .AutoincrementPrimaryKey }},name text,description text,nostr_pubkey text UNIQUE,created_at {{ .Timestamp }},updated_at {{ .Timestamp }}, isolated boolean);
INSERT INTO apps_2 (id, name, description, nostr_pubkey, created_at, updated_at, isolated) SELECT id, name text, description, nostr_pubkey, created_at, updated_at, isolated FROM apps;
CREATE TABLE app_permissions_2 (id {{ .AutoincrementPrimaryKey }},app_id integer,"scope" text,"max_amount_sat" integer,budget_renewal text,expires_at {{ .Timestamp }},created_at {{ .Timestamp }},updated_at {{ .Timestamp }},CONSTRAINT fk_app_permissions_app FOREIGN KEY (app_id) REFERENCES apps_2(id) ON DELETE CASCADE);
INSERT INTO app_permissions_2 (id, app_id, scope, max_amount_sat, budget_renewal, expires_at, created_at, updated_at) SELECT id, app_id, scope, max_amount_sat, budget_renewal, expires_at, created_at, updated_at FROM app_permissions;

DROP TABLE apps {{ .DropTableCascade }};
ALTER TABLE apps_2 RENAME TO apps;
DROP TABLE app_permissions;
ALTER TABLE app_permissions_2 RENAME TO app_permissions;

CREATE INDEX idx_app_permissions_scope ON app_permissions("scope");
CREATE INDEX idx_app_permissions_app_id ON app_permissions(app_id);
`

var appsMigrationTmpl = template.Must(template.New("appsMigration").Parse(appsMigration))

const reqRespMigration = `
CREATE TABLE "request_events" (id {{ .AutoincrementPrimaryKey }},app_id integer,nostr_id text UNIQUE,state text,created_at {{ .Timestamp }},updated_at {{ .Timestamp }}, method TEXT, content_data TEXT,CONSTRAINT fk_request_events_app FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE);
CREATE INDEX idx_request_events_app_id ON request_events(app_id);
CREATE INDEX idx_request_events_app_id_and_id ON request_events(app_id, id);
CREATE INDEX idx_request_events_method ON request_events(method);
CREATE TABLE "response_events" (id {{ .AutoincrementPrimaryKey }},nostr_id text UNIQUE,request_id integer,state text,replied_at {{ .Timestamp }},created_at {{ .Timestamp }},updated_at {{ .Timestamp }},CONSTRAINT fk_response_events_request_event FOREIGN KEY (request_id) REFERENCES request_events(id) ON DELETE CASCADE);
`

var reqRespMigrationTmpl = template.Must(template.New("reqRespMigration").Parse(reqRespMigration))

// This migration (inside a DB transaction),
// - Adds AUTOINCREMENT to the primary key of:
// - apps, app_permissions, request_events, response_events
//
// user_configs is not migrated as it has no relations with other tables, therefore hopefully no issue with reusing IDs
//
// request_events and response_events are not critical (and also payments are dropped in the same release)
// so we just drop those tables and re-create them.
var _202407151352_autoincrement = &gormigrate.Migration{
	ID: "202407151352_autoincrement",
	Migrate: func(db *gorm.DB) error {

		if err := db.Transaction(func(tx *gorm.DB) error {

			// drop old request and response event tables
			if err := exec(tx, dropMigrationTmpl); err != nil {
				return err
			}

			// Apps & app permissions (interdependent)
			// create new tables, copy old values, delete old tables, rename new tables, create new indexes
			// also deletes broken app permissions no longer linked to apps (from reused app IDs)
			if err := exec(tx, appsMigrationTmpl); err != nil {
				return err
			}

			// create fresh request and response event tables
			if err := exec(tx, reqRespMigrationTmpl); err != nil {
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

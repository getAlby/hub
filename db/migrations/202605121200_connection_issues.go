package migrations

import (
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const connectionIssuesMigration = `
CREATE TABLE connection_issues(
	id {{ .AutoincrementPrimaryKey }},
	app_id integer NOT NULL,
	request_event_id integer NOT NULL,
	method text,
	category text NOT NULL,
	error_code text,
	error_message text,
	created_at {{ .Timestamp }},
	updated_at {{ .Timestamp }},
	CONSTRAINT fk_connection_issues_app FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
	CONSTRAINT fk_connection_issues_request_event FOREIGN KEY (request_event_id) REFERENCES request_events(id) ON DELETE CASCADE
);
CREATE INDEX idx_connection_issues_app_id_created_at ON connection_issues(app_id, created_at);
CREATE INDEX idx_connection_issues_request_event_id ON connection_issues(request_event_id);
`

var connectionIssuesMigrationTmpl = template.Must(template.New("connectionIssuesMigration").Parse(connectionIssuesMigration))

var _202605121200_connection_issues = &gormigrate.Migration{
	ID: "202605121200_connection_issues",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, connectionIssuesMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}

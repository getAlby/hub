package migrations

import (
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Create a composite index to improve performance of summing payments in the current budget period
var _202309271618_add_payment_sum_index = &gormigrate.Migration {
	ID: "202309271618_add_payment_sum_index",
	Migrate: func(tx *gorm.DB) error {
		
		var sql string
		if tx.Dialector.Name() == "postgres" {
			sql = "CREATE INDEX idx_payment_sum ON payments USING btree (app_id, preimage, created_at) INCLUDE(amount)"
		} else if tx.Dialector.Name() == "sqlite" {
			sql = "CREATE INDEX idx_payment_sum ON payments (app_id, preimage, created_at)"
		} else {
			log.Fatalf("unsupported database type: %s", tx.Dialector.Name())
		}

		return tx.Exec(sql).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return nil;
	},
}
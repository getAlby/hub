package queries

import (
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func GetTotalSubwalletBalance(tx *gorm.DB) int64 {
	subwalletAppIDsQuery := tx.Model(&db.App{}).
		Select("id").
		Where(datatypes.JSONQuery("metadata").Equals(constants.SUBWALLET_APPSTORE_APP_ID, "app_store_app_id"))

	var received struct {
		Sum int64
	}
	tx.
		Table("transactions").
		Select("SUM(amount_msat) as sum").
		Where("app_id IN (?) AND type = ? AND state = ?", subwalletAppIDsQuery, constants.TRANSACTION_TYPE_INCOMING, constants.TRANSACTION_STATE_SETTLED).
		Scan(&received)

	var spent struct {
		Sum int64
	}
	tx.
		Table("transactions").
		Select("SUM(amount_msat + fee_msat + fee_reserve_msat) as sum").
		Where("app_id IN (?) AND type = ? AND (state = ? OR state = ?)", subwalletAppIDsQuery, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_SETTLED, constants.TRANSACTION_STATE_PENDING).
		Scan(&spent)

	return received.Sum - spent.Sum
}

package queries

import (
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func GetTotalSubwalletBalance(tx *gorm.DB) (int64, error) {
	subwalletAppIDsQuery := tx.Model(&db.App{}).
		Select("id").
		Where(datatypes.JSONQuery("metadata").Equals(constants.SUBWALLET_APPSTORE_APP_ID, constants.METADATA_APPSTORE_APP_ID_KEY))

	var received struct {
		Sum int64
	}
	res := tx.
		Table("transactions").
		Select("SUM(amount_msat) as sum").
		Where("app_id IN (?) AND type = ? AND state = ?", subwalletAppIDsQuery, constants.TRANSACTION_TYPE_INCOMING, constants.TRANSACTION_STATE_SETTLED).
		Scan(&received)
	if res.Error != nil {
		return 0, res.Error
	}

	var spent struct {
		Sum int64
	}
	res = tx.
		Table("transactions").
		Select("SUM(amount_msat + fee_msat + fee_reserve_msat) as sum").
		Where("app_id IN (?) AND type = ? AND (state = ? OR state = ?)", subwalletAppIDsQuery, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_SETTLED, constants.TRANSACTION_STATE_PENDING).
		Scan(&spent)
	if res.Error != nil {
		return 0, res.Error
	}

	return received.Sum - spent.Sum, nil
}

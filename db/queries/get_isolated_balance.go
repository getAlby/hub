package queries

import (
	"github.com/getAlby/hub/constants"
	"gorm.io/gorm"
)

func GetIsolatedBalance(tx *gorm.DB, appId uint) (int64, error) {
	var received struct {
		Sum int64
	}
	err := tx.
		Table("transactions").
		Select("SUM(amount_msat) as sum").
		Where("app_id = ? AND type = ? AND state = ?", appId, constants.TRANSACTION_TYPE_INCOMING, constants.TRANSACTION_STATE_SETTLED).Scan(&received).Error
	if err != nil {
		return 0, err
	}

	var spent struct {
		Sum int64
	}

	err = tx.
		Table("transactions").
		Select("SUM(amount_msat + fee_msat + fee_reserve_msat) as sum").
		Where("app_id = ? AND type = ? AND (state = ? OR state = ?)", appId, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_SETTLED, constants.TRANSACTION_STATE_PENDING).Scan(&spent).Error
	if err != nil {
		return 0, err
	}

	return received.Sum - spent.Sum, nil
}

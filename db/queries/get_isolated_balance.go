package queries

import (
	"github.com/getAlby/hub/constants"
	"gorm.io/gorm"
)

func GetIsolatedBalance(tx *gorm.DB, appId uint) uint64 {
	var received struct {
		Sum uint64
	}
	tx.
		Table("transactions").
		Select("SUM(amount) as sum").
		Where("app_id = ? AND type = ? AND state = ?", appId, constants.TRANSACTION_TYPE_INCOMING, constants.TRANSACTION_STATE_SETTLED).Scan(&received)

	var spent struct {
		Sum uint64
	}
	// TODO: ensure fee reserve on these payments
	tx.
		Table("transactions").
		Select("SUM(amount + fee) as sum").
		Where("app_id = ? AND type = ? AND (state = ? OR state = ?)", appId, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_SETTLED, constants.TRANSACTION_STATE_PENDING).Scan(&spent)

	return received.Sum - spent.Sum
}

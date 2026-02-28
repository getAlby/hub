package queries

import (
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"gorm.io/gorm"
)

func GetBudgetUsage(tx *gorm.DB, appPermission *db.AppPermission) (uint64, error) {
	var result struct {
		Sum uint64
	}
	err := tx.
		Table("transactions").
		Select("SUM(amount_msat + fee_msat + fee_reserve_msat) as sum").
		Where("app_id = ? AND type = ? AND (state = ? OR state = ?) AND created_at > ?", appPermission.AppId, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_SETTLED, constants.TRANSACTION_STATE_PENDING, getStartOfBudget(appPermission.BudgetRenewal)).Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.Sum, nil
}

func getStartOfBudget(budget_type string) time.Time {
	now := time.Now()
	switch budget_type {
	case constants.BUDGET_RENEWAL_DAILY:
		// TODO: Use the location of the user, instead of the server
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case constants.BUDGET_RENEWAL_WEEKLY:
		weekday := now.Weekday()
		var startOfWeek time.Time
		if weekday == 0 {
			startOfWeek = now.AddDate(0, 0, -6)
		} else {
			startOfWeek = now.AddDate(0, 0, -int(weekday)+1)
		}
		return time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	case constants.BUDGET_RENEWAL_MONTHLY:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case constants.BUDGET_RENEWAL_YEARLY:
		return time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	default: //"never"
		return time.Time{}
	}
}

func GetBudgetRenewsAt(budgetRenewal string) *uint64 {
	budgetStart := getStartOfBudget(budgetRenewal)
	switch budgetRenewal {
	case constants.BUDGET_RENEWAL_DAILY:
		renewal := uint64(budgetStart.AddDate(0, 0, 1).Unix())
		return &renewal
	case constants.BUDGET_RENEWAL_WEEKLY:
		renewal := uint64(budgetStart.AddDate(0, 0, 7).Unix())
		return &renewal

	case constants.BUDGET_RENEWAL_MONTHLY:
		renewal := uint64(budgetStart.AddDate(0, 1, 0).Unix())
		return &renewal

	case constants.BUDGET_RENEWAL_YEARLY:
		renewal := uint64(budgetStart.AddDate(1, 0, 0).Unix())
		return &renewal

	default: //"never"
		return nil
	}
}

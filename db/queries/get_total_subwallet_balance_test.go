package queries

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestGetTotalSubwalletBalance(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	subwalletA, _, err := tests.CreateApp(svc)
	require.NoError(t, err)
	subwalletA.Isolated = true
	subwalletA.Metadata = datatypes.JSON([]byte(fmt.Sprintf(`{"app_store_app_id":"%s"}`, constants.SUBWALLET_APPSTORE_APP_ID)))
	svc.DB.Save(&subwalletA)

	subwalletB, _, err := tests.CreateApp(svc)
	require.NoError(t, err)
	subwalletB.Isolated = true
	subwalletB.Metadata = datatypes.JSON([]byte(fmt.Sprintf(`{"app_store_app_id":"%s"}`, constants.SUBWALLET_APPSTORE_APP_ID)))
	svc.DB.Save(&subwalletB)

	incomingSubwalletTx := db.Transaction{
		AppId:      &subwalletA.ID,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		State:      constants.TRANSACTION_STATE_SETTLED,
		AmountMsat: 5000,
	}
	svc.DB.Save(&incomingSubwalletTx)

	outgoingSettledSubwalletTx := db.Transaction{
		AppId:          &subwalletA.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     1000,
		FeeMsat:        100,
		FeeReserveMsat: 0,
	}
	svc.DB.Save(&outgoingSettledSubwalletTx)

	outgoingPendingSubwalletTx := db.Transaction{
		AppId:          &subwalletB.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_PENDING,
		AmountMsat:     2000,
		FeeMsat:        0,
		FeeReserveMsat: 300,
	}
	svc.DB.Save(&outgoingPendingSubwalletTx)

	total := GetTotalSubwalletBalance(svc.DB)
	assert.Equal(t, int64(1600), total)
}

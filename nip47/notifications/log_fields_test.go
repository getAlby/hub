package notifications

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/nip47/models"
)

func TestLogFields_OmitPrivateData(t *testing.T) {
	preimage := "secret-preimage"
	appId := uint(7)
	transaction := &db.Transaction{
		ID:          3,
		AppId:       &appId,
		PaymentHash: "payment-hash",
		Preimage:    &preimage,
		AmountMsat:  1000,
		Metadata:    []byte(`{"comment":"private comment"}`),
	}

	eventFields, err := json.Marshal(EventLogFields(&events.Event{Event: "nwc_payment_received", Properties: transaction}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"event":"nwc_payment_received","transaction_id":3,"payment_hash":"payment-hash","app_id":7,"amount_msat":1000}`, string(eventFields))

	notification := &Notification{
		NotificationType: PAYMENT_RECEIVED_NOTIFICATION,
		Notification: PaymentReceivedNotification{Transaction: models.Transaction{
			PaymentHash: "payment-hash",
			Preimage:    preimage,
			Metadata:    map[string]string{"comment": "private comment"},
		}},
	}
	notificationFields, err := json.Marshal(notification.logFields())
	require.NoError(t, err)
	assert.JSONEq(t, `{"notification_type":"payment_received","payment_hash":"payment-hash"}`, string(notificationFields))
}

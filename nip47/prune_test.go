package nip47

import (
	"fmt"
	"testing"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
)

func TestPruneRequestEvents(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	assert.NoError(t, err)
	defer svc.Remove()

	app := db.App{
		Name:      "Test App",
		AppPubkey: "pubkey",
	}
	err = svc.DB.Create(&app).Error
	assert.NoError(t, err)

	nip47Svc := &nip47Service{
		db: svc.DB,
	}

	for i := 0; i < 1005; i++ {
		requestEvent := db.RequestEvent{
			AppId:     &app.ID,
			NostrId:   fmt.Sprintf("event_%d", i),
			State:     db.REQUEST_EVENT_STATE_HANDLER_EXECUTED,
			CreatedAt: time.Now(),
		}
		err = svc.DB.Create(&requestEvent).Error
		assert.NoError(t, err)
	}

	var count int64
	svc.DB.Model(&db.RequestEvent{}).Where("app_id = ?", app.ID).Count(&count)
	assert.Equal(t, int64(1005), count)

	nip47Svc.pruneRequestEvents(app.ID)

	svc.DB.Model(&db.RequestEvent{}).Where("app_id = ?", app.ID).Count(&count)
	assert.Equal(t, int64(MaxRequestEventsPerApp), count)

	
	var remaining []db.RequestEvent
	svc.DB.Where("app_id = ?", app.ID).Order("id asc").Find(&remaining)
	assert.Equal(t, MaxRequestEventsPerApp, len(remaining))
	assert.Equal(t, "event_5", remaining[0].NostrId)
	assert.Equal(t, "event_1004", remaining[len(remaining)-1].NostrId)
}

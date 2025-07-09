package maintenance

import (
	"context"
	"testing"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaintenanceService_CleanupOldRequestEvents(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	maintenanceSvc := NewMaintenanceService(svc.DB, svc.Cfg)

	// Test with no events - should not fail
	err = maintenanceSvc.CleanupOldRequestEvents()
	assert.NoError(t, err)

	// Create test request events
	createTestRequestEvents(t, svc, 100)

	// Verify initial count
	var count int64
	err = svc.DB.Model(&db.RequestEvent{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(100), count)

	// Cleanup should not run since count is below threshold (10000)
	err = maintenanceSvc.CleanupOldRequestEvents()
	assert.NoError(t, err)

	// Verify count remains the same
	err = svc.DB.Model(&db.RequestEvent{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(100), count)
}

func TestMaintenanceService_CleanupOldRequestEventsExceedsThreshold(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// Create maintenance service with lower threshold for testing
	maintenanceSvc := &maintenanceService{
		db:               svc.DB,
		cfg:              svc.Cfg,
		maxRequestEvents: 50, // Lower threshold for testing
		cleanupInterval:  time.Hour,
	}

	// Create 100 test request events
	createTestRequestEvents(t, svc, 100)

	// Verify initial count
	var count int64
	err = svc.DB.Model(&db.RequestEvent{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(100), count)

	// Run cleanup - should delete oldest 50, keeping newest 50
	err = maintenanceSvc.CleanupOldRequestEvents()
	assert.NoError(t, err)

	// Verify count is now 50
	err = svc.DB.Model(&db.RequestEvent{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50), count)

	// Run cleanup again - should not delete anything
	err = maintenanceSvc.CleanupOldRequestEvents()
	assert.NoError(t, err)

	// Verify count remains 50
	err = svc.DB.Model(&db.RequestEvent{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50), count)
}

func TestMaintenanceService_CleanupOldRequestEventsWithResponseEvents(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// Create maintenance service with lower threshold for testing
	maintenanceSvc := &maintenanceService{
		db:               svc.DB,
		cfg:              svc.Cfg,
		maxRequestEvents: 50,
		cleanupInterval:  time.Hour,
	}

	// Create 100 test request events with corresponding response events
	createTestRequestEventsWithResponses(t, svc, 100)

	// Verify initial counts
	var requestCount, responseCount int64
	err = svc.DB.Model(&db.RequestEvent{}).Count(&requestCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(100), requestCount)

	err = svc.DB.Model(&db.ResponseEvent{}).Count(&responseCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(100), responseCount)

	// Run cleanup - should delete oldest 50 request events
	// Response events should be deleted automatically due to foreign key cascade
	err = maintenanceSvc.CleanupOldRequestEvents()
	assert.NoError(t, err)

	// Verify both request and response counts are now 50
	err = svc.DB.Model(&db.RequestEvent{}).Count(&requestCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50), requestCount)

	err = svc.DB.Model(&db.ResponseEvent{}).Count(&responseCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(50), responseCount)
}

func TestMaintenanceService_Start(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	maintenanceSvc := NewMaintenanceService(svc.DB, svc.Cfg)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the service
	err = maintenanceSvc.Start(ctx)
	assert.NoError(t, err)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context
	cancel()

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)

	// Test passes if no panic or error occurred
}

func TestMaintenanceService_StartWithShortInterval(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// Create maintenance service with very short interval for testing
	maintenanceSvc := &maintenanceService{
		db:               svc.DB,
		cfg:              svc.Cfg,
		maxRequestEvents: 5,
		cleanupInterval:  100 * time.Millisecond, // Very short interval
	}

	// Create some test data
	createTestRequestEvents(t, svc, 10)

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start the service
	err = maintenanceSvc.Start(ctx)
	assert.NoError(t, err)

	// Wait for the context to timeout (cleanup should have run)
	<-ctx.Done()

	// Verify cleanup ran and reduced count
	var count int64
	err = svc.DB.Model(&db.RequestEvent{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestMaintenanceService_Configuration(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// Set configuration values
	err = svc.Cfg.SetUpdate("MaintenanceMaxRequestEvents", "500", "")
	require.NoError(t, err)
	err = svc.Cfg.SetUpdate("MaintenanceCleanupIntervalHours", "12", "")
	require.NoError(t, err)

	// Create maintenance service
	maintenanceSvc := NewMaintenanceService(svc.DB, svc.Cfg)

	// Verify configuration was loaded
	svcImpl := maintenanceSvc.(*maintenanceService)
	assert.Equal(t, 500, svcImpl.maxRequestEvents)
	assert.Equal(t, 12*time.Hour, svcImpl.cleanupInterval)
}

func TestMaintenanceService_CleanupKeepsNewestEvents(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// Create maintenance service with lower threshold for testing
	maintenanceSvc := &maintenanceService{
		db:               svc.DB,
		cfg:              svc.Cfg,
		maxRequestEvents: 3,
		cleanupInterval:  time.Hour,
	}

	// Create 5 test request events with specific nostr_ids to track them
	events := []db.RequestEvent{
		{NostrId: "event-1", State: "executed"},
		{NostrId: "event-2", State: "executed"},
		{NostrId: "event-3", State: "executed"},
		{NostrId: "event-4", State: "executed"},
		{NostrId: "event-5", State: "executed"},
	}

	for _, event := range events {
		err = svc.DB.Create(&event).Error
		require.NoError(t, err)
		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	// Run cleanup - should keep the 3 newest events (3, 4, 5)
	err = maintenanceSvc.CleanupOldRequestEvents()
	assert.NoError(t, err)

	// Verify we have 3 events remaining
	var count int64
	err = svc.DB.Model(&db.RequestEvent{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Verify the remaining events are the newest ones
	var remainingEvents []db.RequestEvent
	err = svc.DB.Order("id ASC").Find(&remainingEvents).Error
	require.NoError(t, err)
	require.Len(t, remainingEvents, 3)

	// The remaining events should be the ones with the highest IDs
	assert.Equal(t, "event-3", remainingEvents[0].NostrId)
	assert.Equal(t, "event-4", remainingEvents[1].NostrId)
	assert.Equal(t, "event-5", remainingEvents[2].NostrId)
}

// Helper function to create test request events
func createTestRequestEvents(t *testing.T, svc *tests.TestService, count int) {
	for i := 0; i < count; i++ {
		event := db.RequestEvent{
			NostrId: "test-event-" + string(rune(i)),
			State:   "executed",
		}
		err := svc.DB.Create(&event).Error
		require.NoError(t, err)
	}
}

// Helper function to create test request events with corresponding response events
func createTestRequestEventsWithResponses(t *testing.T, svc *tests.TestService, count int) {
	for i := 0; i < count; i++ {
		// Create request event
		requestEvent := db.RequestEvent{
			NostrId: "test-request-" + string(rune(i)),
			State:   "executed",
		}
		err := svc.DB.Create(&requestEvent).Error
		require.NoError(t, err)

		// Create corresponding response event
		responseEvent := db.ResponseEvent{
			NostrId:   "test-response-" + string(rune(i)),
			RequestId: requestEvent.ID,
			State:     "confirmed",
		}
		err = svc.DB.Create(&responseEvent).Error
		require.NoError(t, err)
	}
}
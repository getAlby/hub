package maintenance

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	// DefaultMaxRequestEvents is the default maximum number of request events to keep
	DefaultMaxRequestEvents = 10000
	// DefaultCleanupInterval is the default interval between cleanup runs (24 hours)
	DefaultCleanupInterval = 24 * time.Hour
	// Configuration keys
	ConfigKeyMaxRequestEvents = "MaintenanceMaxRequestEvents"
	ConfigKeyCleanupInterval  = "MaintenanceCleanupIntervalHours"
)

// MaintenanceService handles periodic cleanup of NIP-47 request and response events
type MaintenanceService interface {
	// Start begins the maintenance service background tasks
	Start(ctx context.Context) error
	// CleanupOldRequestEvents removes oldest request events if count exceeds threshold
	CleanupOldRequestEvents() error
}

type maintenanceService struct {
	db               *gorm.DB
	cfg              config.Config
	maxRequestEvents int
	cleanupInterval  time.Duration
}

// NewMaintenanceService creates a new maintenance service instance
func NewMaintenanceService(db *gorm.DB, cfg config.Config) MaintenanceService {
	svc := &maintenanceService{
		db:               db,
		cfg:              cfg,
		maxRequestEvents: DefaultMaxRequestEvents,
		cleanupInterval:  DefaultCleanupInterval,
	}

	// Load configuration values
	svc.loadConfig()

	return svc
}

// loadConfig loads configuration values from the config service
func (svc *maintenanceService) loadConfig() {
	// Load max request events threshold
	if maxEventsStr, err := svc.cfg.Get(ConfigKeyMaxRequestEvents, ""); err == nil && maxEventsStr != "" {
		if maxEvents, err := strconv.Atoi(maxEventsStr); err == nil && maxEvents > 0 {
			if maxEvents < 1000 {
				logger.Logger.WithField("value", maxEvents).Warn("MaintenanceMaxRequestEvents is very low, consider using at least 1000")
			}
			svc.maxRequestEvents = maxEvents
		} else {
			logger.Logger.WithFields(logrus.Fields{
				"value": maxEventsStr,
				"error": err,
			}).Warn("Invalid MaintenanceMaxRequestEvents configuration, using default")
		}
	}

	// Load cleanup interval
	if intervalStr, err := svc.cfg.Get(ConfigKeyCleanupInterval, ""); err == nil && intervalStr != "" {
		if intervalHours, err := strconv.Atoi(intervalStr); err == nil && intervalHours > 0 {
			if intervalHours < 1 {
				logger.Logger.WithField("value", intervalHours).Warn("MaintenanceCleanupIntervalHours is very low, consider using at least 1 hour")
			}
			svc.cleanupInterval = time.Duration(intervalHours) * time.Hour
		} else {
			logger.Logger.WithFields(logrus.Fields{
				"value": intervalStr,
				"error": err,
			}).Warn("Invalid MaintenanceCleanupIntervalHours configuration, using default")
		}
	}
}

// Start begins the maintenance service background tasks
func (svc *maintenanceService) Start(ctx context.Context) error {
	logger.Logger.WithFields(logrus.Fields{
		"max_request_events": svc.maxRequestEvents,
		"cleanup_interval":   svc.cleanupInterval,
	}).Info("Starting NIP-47 maintenance service")

	// Start the cleanup ticker in a goroutine
	go svc.runPeriodicCleanup(ctx)

	// Run initial cleanup
	go func() {
		if err := svc.CleanupOldRequestEvents(); err != nil {
			logger.Logger.WithError(err).Error("Initial maintenance cleanup failed")
			// Continue startup even if initial cleanup fails
		}
	}()

	return nil
}

// runPeriodicCleanup runs the cleanup task at regular intervals
func (svc *maintenanceService) runPeriodicCleanup(ctx context.Context) {
	ticker := time.NewTicker(svc.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Logger.Info("Maintenance service stopped")
			return
		case <-ticker.C:
			logger.Logger.Debug("Running scheduled NIP-47 event cleanup")
			if err := svc.CleanupOldRequestEvents(); err != nil {
				logger.Logger.WithError(err).Error("Scheduled maintenance cleanup failed")
				// Continue running even if cleanup fails - don't crash the service
			}
		}
	}
}

// CleanupOldRequestEvents removes oldest request events if count exceeds threshold
// Response events are automatically deleted due to foreign key cascade constraint
func (svc *maintenanceService) CleanupOldRequestEvents() error {
	// Count current request events
	var count int64
	if err := svc.db.Model(&RequestEvent{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count request events: %w", err)
	}

	logger.Logger.WithFields(logrus.Fields{
		"current_count":      count,
		"max_request_events": svc.maxRequestEvents,
	}).Debug("Checking NIP-47 request events count")

	// Only cleanup if we exceed the threshold
	if count <= int64(svc.maxRequestEvents) {
		logger.Logger.Debug("Request events count within threshold, no cleanup needed")
		return nil
	}

	// Calculate how many events to delete
	eventsToDelete := count - int64(svc.maxRequestEvents)

	logger.Logger.WithFields(logrus.Fields{
		"current_count":    count,
		"events_to_delete": eventsToDelete,
		"keeping":          svc.maxRequestEvents,
	}).Info("Starting NIP-47 request events cleanup")

	// Use a transaction to ensure consistency
	tx := svc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete oldest request events (keeping the most recent maxRequestEvents)
	// This uses a subquery to identify the oldest events by ID (which auto-increments)
	result := tx.Exec(`
		DELETE FROM request_events 
		WHERE id NOT IN (
			SELECT id FROM request_events 
			ORDER BY id DESC 
			LIMIT ?
		)
	`, svc.maxRequestEvents)

	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old request events: %w", result.Error)
	}

	deletedCount := result.RowsAffected

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	logger.Logger.WithFields(logrus.Fields{
		"deleted_count":    deletedCount,
		"remaining_count":  count - deletedCount,
	}).Info("Successfully completed NIP-47 request events cleanup")

	return nil
}

// RequestEvent represents the request_events table structure
// This is a minimal representation for the maintenance service
type RequestEvent struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time
}

// TableName returns the table name for RequestEvent
func (RequestEvent) TableName() string {
	return "request_events"
}
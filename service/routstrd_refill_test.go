package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShouldAutoRefillBelowThresholdAndCooldownElapsed(t *testing.T) {
	nowT := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	last := nowT.Add(-6 * time.Minute)
	assert.True(t, shouldAutoRefill(100, 500, last, nowT, 5*60*1000))
}

func TestShouldAutoRefillAboveThresholdIgnoresCooldown(t *testing.T) {
	nowT := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	last := nowT.Add(-30 * time.Minute) // cooldown long elapsed
	assert.False(t, shouldAutoRefill(500, 500, last, nowT, 5*60*1000))
	assert.False(t, shouldAutoRefill(501, 500, last, nowT, 5*60*1000))
}

func TestShouldAutoRefillWithinCooldown(t *testing.T) {
	nowT := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	last := nowT.Add(-4 * time.Minute) // < 5 min cooldown
	assert.False(t, shouldAutoRefill(100, 500, last, nowT, 5*60*1000))
}

func TestShouldAutoRefillExactCooldownBoundary(t *testing.T) {
	nowT := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	last := nowT.Add(-5 * time.Minute) // exactly the cooldown → allowed
	assert.True(t, shouldAutoRefill(100, 500, last, nowT, 5*60*1000))
}

func TestShouldAutoRefillZeroCooldownDefaultsToFiveMinutes(t *testing.T) {
	nowT := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	// 0 cooldownMs → 5 min default: 4 min ago is still within it
	assert.False(t, shouldAutoRefill(100, 500, nowT.Add(-4*time.Minute), nowT, 0))
	// 6 min ago passes the default cooldown
	assert.True(t, shouldAutoRefill(100, 500, nowT.Add(-6*time.Minute), nowT, 0))
}

func TestShouldAutoRefillNeverRefilled(t *testing.T) {
	nowT := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	// Zero last-refill time = never → cooldown gate always passes
	assert.True(t, shouldAutoRefill(100, 500, time.Time{}, nowT, 5*60*1000))
}

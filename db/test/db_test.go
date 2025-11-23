package test

import (
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/tests/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTempStorePragmaIsApplied(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	if gormDb.Dialector.Name() != "sqlite" {
		t.Skip("Skipping non-sqlite dialector")
	}

	var result string
	err = gormDb.Raw("PRAGMA temp_store").Scan(&result).Error
	require.NoError(t, err)

	// PRAGMA temp_store = MEMORY
	// MEMORY = 2
	assert.Equal(t, "2", result)
}

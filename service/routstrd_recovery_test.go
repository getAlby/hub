package service

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/getAlby/hub/logger"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initTestLogger ensures logger.Logger is non-nil (main initializes it in
// production; tests must do the same before any log call).
func initTestLogger() {
	logger.Init(strconv.Itoa(int(logrus.WarnLevel)))
}

const cocoOpsSchema = `
CREATE TABLE coco_cashu_mint_operations (
  id TEXT PRIMARY KEY,
  mintUrl TEXT NOT NULL,
  quoteId TEXT,
  state TEXT NOT NULL,
  createdAt INTEGER NOT NULL,
  updatedAt INTEGER NOT NULL,
  error TEXT,
  method TEXT NOT NULL,
  methodDataJson TEXT NOT NULL,
  amount INTEGER,
  unit TEXT
);`

func newCocoTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "coco.db"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	_, err = db.Exec(cocoOpsSchema)
	require.NoError(t, err)
	return db
}

func TestClearStuckCocodOpsDeletesOnlyPending(t *testing.T) {
	initTestLogger()
	db := newCocoTestDB(t)
	_, err := db.Exec(`INSERT INTO coco_cashu_mint_operations (id, mintUrl, state, createdAt, updatedAt, method, methodDataJson) VALUES
		('op-pending-1', 'https://mint.test', 'pending', 1, 1, 'bolt11', '{}'),
		('op-pending-2', 'https://mint.test', 'pending', 1, 1, 'bolt11', '{}'),
		('op-finalized', 'https://mint.test', 'finalized', 1, 1, 'bolt11', '{}'),
		('op-failed',    'https://mint.test', 'failed', 1, 1, 'bolt11', '{}')`)
	require.NoError(t, err)

	clearStuckCocodOps(dbPathOf(t, db))

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM coco_cashu_mint_operations").Scan(&count))
	assert.Equal(t, 2, count, "only finalized + failed should remain")

	var pendCount int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM coco_cashu_mint_operations WHERE state = 'pending'").Scan(&pendCount))
	assert.Equal(t, 0, pendCount)

	// Finalized/failed ops survive
	var finCount int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM coco_cashu_mint_operations WHERE state IN ('finalized','failed')").Scan(&finCount))
	assert.Equal(t, 2, finCount)
}

func TestClearStuckCocodOpsNoPending(t *testing.T) {
	initTestLogger()
	db := newCocoTestDB(t)
	_, err := db.Exec(`INSERT INTO coco_cashu_mint_operations (id, mintUrl, state, createdAt, updatedAt, method, methodDataJson) VALUES
		('op-finalized', 'https://mint.test', 'finalized', 1, 1, 'bolt11', '{}')`)
	require.NoError(t, err)

	clearStuckCocodOps(dbPathOf(t, db))

	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM coco_cashu_mint_operations").Scan(&count))
	assert.Equal(t, 1, count)
}

func TestClearStuckCocodOpsMissingDBDoesNotPanic(t *testing.T) {
	initTestLogger()
	// A missing/unopenable DB must be a no-op, not a panic.
	clearStuckCocodOps(filepath.Join(t.TempDir(), "does-not-exist", "coco.db"))
}

// dbPathOf returns the file path backing db (single connection, file-based).
func dbPathOf(t *testing.T, db *sql.DB) string {
	t.Helper()
	var path string
	require.NoError(t, db.QueryRow("PRAGMA database_list").Scan(new(any), new(any), &path))
	return path
}

package api

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	test_db "github.com/getAlby/hub/tests/db"
	"github.com/getAlby/hub/tests/mocks"
)

// TestCreateBackup creates a backup from the test database (sqlite by
// default, postgres when TEST_DATABASE_URI is set) and verifies that the
// archive contains a valid sqlite database with the expected data.
func TestCreateBackup(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	workDir := t.TempDir()

	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	defer test_db.CloseDB(gormDB)

	appConfig := &config.AppConfig{
		Workdir:     workDir,
		DatabaseUri: test_db.GetTestDatabaseURI(),
	}
	cfg, err := config.NewConfig(appConfig, gormDB)
	require.NoError(t, err)

	app := &db.App{
		Name:      "test",
		AppPubkey: "2b7dea2866958f17c568cf024e113db7a3baa9c253a9016889196b8d0b11c7ae",
		Metadata:  datatypes.JSON("{}"),
	}
	require.NoError(t, gormDB.Create(app).Error)

	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetStorageDir").Return("", nil)
	lnClient.On("ResetRouter", "ALL").Return(nil)

	svc := mocks.NewMockService(t)
	svc.On("GetLNClient").Return(lnClient)
	svc.On("StopApp").Return()

	albyOAuthSvc := mocks.NewMockAlbyOAuthService(t)
	albyOAuthSvc.On("RemoveOAuthAccessToken").Return(nil)

	theAPI := &api{
		db:           gormDB,
		cfg:          cfg,
		svc:          svc,
		albyOAuthSvc: albyOAuthSvc,
	}

	unlockPassword := ""

	var buf bytes.Buffer
	err = theAPI.CreateBackup(unlockPassword, &buf)
	require.NoError(t, err)

	// The temporary database created when converting from postgres must
	// not be left behind in the working directory.
	entries, err := os.ReadDir(workDir)
	require.NoError(t, err)
	require.Empty(t, entries)

	cr, err := decryptingReader(&buf, unlockPassword)
	require.NoError(t, err)
	decrypted, err := io.ReadAll(cr)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(decrypted), int64(len(decrypted)))
	require.NoError(t, err)

	dbFile, err := zr.Open("nwc.db")
	require.NoError(t, err)
	dbContents, err := io.ReadAll(dbFile)
	require.NoError(t, err)
	require.NoError(t, dbFile.Close())

	restoredPath := filepath.Join(workDir, "restored.db")
	require.NoError(t, os.WriteFile(restoredPath, dbContents, 0600))

	restoredDB, err := db.NewDB(restoredPath, false)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, db.Stop(restoredDB))
	}()

	var restoredApp db.App
	require.NoError(t, restoredDB.First(&restoredApp).Error)
	require.Equal(t, app.Name, restoredApp.Name)
	require.Equal(t, app.AppPubkey, restoredApp.AppPubkey)
}

// TestRestoreBackupRejectsPathTraversal verifies that a backup archive
// containing an entry whose name points outside the restore directory is
// rejected and that no file is written outside it.
func TestRestoreBackupRejectsPathTraversal(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	gormDB, err := test_db.NewDB(t)
	require.NoError(t, err)
	defer test_db.CloseDB(gormDB)

	if gormDB.Dialector.Name() != "sqlite" {
		t.Skip("restore is only supported on sqlite")
	}

	workDir := t.TempDir()

	appConfig := &config.AppConfig{
		Workdir:     workDir,
		DatabaseUri: test_db.GetTestDatabaseURI(),
	}
	cfg, err := config.NewConfig(appConfig, gormDB)
	require.NoError(t, err)

	theAPI := &api{
		db:  gormDB,
		cfg: cfg,
	}

	unlockPassword := ""

	// The restore directory is <workDir>/restore, so a "../" entry targets a
	// file directly in the working directory, one level above it.
	const escapeEntryName = "../pwned.txt"
	escapeTarget := filepath.Join(workDir, "pwned.txt")

	var buf bytes.Buffer
	cw, err := encryptingWriter(&buf, unlockPassword)
	require.NoError(t, err)
	zw := zip.NewWriter(cw)
	entryWriter, err := zw.Create(escapeEntryName)
	require.NoError(t, err)
	_, err = entryWriter.Write([]byte("pwned"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	err = theAPI.RestoreBackup(unlockPassword, &buf)
	require.Error(t, err)

	_, statErr := os.Stat(escapeTarget)
	require.True(t, os.IsNotExist(statErr), "traversal entry must not be written outside the restore directory")
}

package api

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"

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

	unlockPassword := ""

	// Represent a fully set-up hub: the unlock-password canary is written during
	// setup and is required for the password check to pass.
	require.NoError(t, cfg.SaveUnlockPasswordCheck(unlockPassword))

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
	svc.On("StopApp").Return(nil)

	albyOAuthSvc := mocks.NewMockAlbyOAuthService(t)
	albyOAuthSvc.On("RemoveOAuthAccessToken").Return(nil)

	theAPI := &api{
		db:           gormDB,
		cfg:          cfg,
		svc:          svc,
		albyOAuthSvc: albyOAuthSvc,
	}

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

	// A backend without a storage directory contributes nothing but the
	// database to the archive.
	require.Len(t, zr.File, 1)

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

// TestCreateBackupArchivesStorageDirRecursively verifies that the whole
// LNClient storage directory tree is archived (e.g. LDK node storage and the
// static channel backups next to it), excluding log files at any depth.
func TestCreateBackupArchivesStorageDirRecursively(t *testing.T) {
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

	unlockPassword := ""
	require.NoError(t, cfg.SaveUnlockPasswordCheck(unlockPassword))

	storageDir := filepath.Join(workDir, "ldk", "storage")
	require.NoError(t, os.MkdirAll(storageDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(storageDir, "node.db"), []byte("node data"), 0600))

	scbDir := filepath.Join(workDir, "ldk", "static_channel_backups")
	require.NoError(t, os.MkdirAll(scbDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(scbDir, "2024-01-02T03-04-05.json"), []byte("{}"), 0600))

	logsDir := filepath.Join(workDir, "ldk", "logs")
	require.NoError(t, os.MkdirAll(logsDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(logsDir, "ldk_node_2024_01_02.log"), []byte("log"), 0600))

	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetStorageDir").Return(filepath.Join(workDir, "ldk"), nil)
	lnClient.On("ResetRouter", "ALL").Return(nil)

	svc := mocks.NewMockService(t)
	svc.On("GetLNClient").Return(lnClient)
	svc.On("StopApp").Return(nil)

	albyOAuthSvc := mocks.NewMockAlbyOAuthService(t)
	albyOAuthSvc.On("RemoveOAuthAccessToken").Return(nil)

	theAPI := &api{
		db:           gormDB,
		cfg:          cfg,
		svc:          svc,
		albyOAuthSvc: albyOAuthSvc,
	}

	var buf bytes.Buffer
	err = theAPI.CreateBackup(unlockPassword, &buf)
	require.NoError(t, err)

	cr, err := decryptingReader(&buf, unlockPassword)
	require.NoError(t, err)
	decrypted, err := io.ReadAll(cr)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(decrypted), int64(len(decrypted)))
	require.NoError(t, err)

	entryNames := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		entryNames = append(entryNames, f.Name)
	}
	require.ElementsMatch(t, []string{
		"nwc.db",
		"ldk/storage/node.db",
		"ldk/static_channel_backups/2024-01-02T03-04-05.json",
	}, entryNames)
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
	// A valid entry before the malicious one, to verify that a partially
	// extracted archive is not left behind when a later entry fails.
	entryWriter, err := zw.Create("nwc.db")
	require.NoError(t, err)
	_, err = entryWriter.Write([]byte("backup contents"))
	require.NoError(t, err)
	entryWriter, err = zw.Create(escapeEntryName)
	require.NoError(t, err)
	_, err = entryWriter.Write([]byte("pwned"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	err = theAPI.RestoreBackup(unlockPassword, &buf)
	require.ErrorContains(t, err, "refusing to extract zip entry outside restore directory")

	_, statErr := os.Stat(escapeTarget)
	require.True(t, os.IsNotExist(statErr), "traversal entry must not be written outside the restore directory")

	// The failed restore must not leave a restore directory (which would be
	// applied on the next startup) or any staging leftovers.
	_, statErr = os.Stat(filepath.Join(workDir, "restore"))
	require.True(t, os.IsNotExist(statErr), "failed restore must not leave a restore directory")

	entries, err := os.ReadDir(workDir)
	require.NoError(t, err)
	for _, entry := range entries {
		require.False(t, strings.HasPrefix(entry.Name(), "albyhub-restore-"), "failed restore must not leave a staging directory")
	}
}

// legacyBackupFixture is a backup file created with the encryption scheme
// used by older versions (PBKDF2 key derivation), encrypted with the
// password "test-unlock-password". Its archive contains a single "nwc.db"
// entry with the contents "legacy backup contents".
const legacyBackupFixture = "0102030405060708101112131415161718191a1b1c1d1e1f8eca79631915f679a00cdd95d3f20d8d169eb9aa5d52642ca13b93886c3c7d7ba4b759462bc9dd8deccf638edcc9b5b9fda3d23dcd904cf6e99bc57ac59c4df6be5aa676542b7cbc9998029420c0ae5a6986c735150ababde5b382560acaebd5894aa4420924f1ced63fde570adc60c43b32e9e14a0ef60c379da5cac1be0000845992ea072ead036e336c7b859e8d018c4ef61667e3f520fe01"

// TestDecryptingReaderLegacyBackup verifies that backup files created by
// older versions can still be decrypted.
func TestDecryptingReaderLegacyBackup(t *testing.T) {
	encrypted, err := hex.DecodeString(legacyBackupFixture)
	require.NoError(t, err)

	cr, err := decryptingReader(bytes.NewReader(encrypted), "test-unlock-password")
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
	require.Equal(t, "legacy backup contents", string(dbContents))
}

// TestDecryptingReaderFragmentedReader verifies that a backup file is
// decrypted correctly even when the reader delivers one byte at a time,
// which would truncate the header if it were not read in full.
func TestDecryptingReaderFragmentedReader(t *testing.T) {
	var buf bytes.Buffer
	cw, err := encryptingWriter(&buf, "test-unlock-password")
	require.NoError(t, err)

	zw := zip.NewWriter(cw)
	entryWriter, err := zw.Create("nwc.db")
	require.NoError(t, err)
	_, err = entryWriter.Write([]byte("backup contents"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	cr, err := decryptingReader(iotest.OneByteReader(bytes.NewReader(buf.Bytes())), "test-unlock-password")
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
	require.Equal(t, "backup contents", string(dbContents))
}

// TestDecryptingReaderWrongPassword verifies that decryption fails upfront
// when the password does not match the backup file.
func TestDecryptingReaderWrongPassword(t *testing.T) {
	var buf bytes.Buffer
	cw, err := encryptingWriter(&buf, "test-unlock-password")
	require.NoError(t, err)

	zw := zip.NewWriter(cw)
	entryWriter, err := zw.Create("nwc.db")
	require.NoError(t, err)
	_, err = entryWriter.Write([]byte("backup contents"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	_, err = decryptingReader(bytes.NewReader(buf.Bytes()), "wrong-password")
	require.Error(t, err)
}

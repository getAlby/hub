package api

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"archive/zip"
	"os"
	"path/filepath"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"golang.org/x/crypto/pbkdf2"
)

// zipMagic is the ZIP local file header signature "PK\x03\x04" — the first
// four bytes of every ZIP file, and therefore of every archive produced by
// CreateBackup. decryptingReader uses it to detect which cipher scheme the
// backup file was created with.
var zipMagic = []byte{'P', 'K', 0x03, 0x04}

// backupCipher describes one of the cipher schemes used for backup files,
// which are laid out as salt || iv || encrypted zip archive.
type backupCipher struct {
	saltSize  int
	deriveKey func(password string, salt []byte) ([]byte, error)
	newStream func(block cipher.Block, iv []byte) cipher.Stream
}

var backupCiphers = []backupCipher{
	// current scheme, used for all new backup files
	{
		saltSize: 32,
		deriveKey: func(password string, salt []byte) ([]byte, error) {
			key, _, err := config.DeriveKey(password, salt)
			return key, err
		},
		newStream: cipher.NewCTR,
	},
	// legacy scheme, kept to restore backup files created by older versions
	{
		saltSize: 8,
		deriveKey: func(password string, salt []byte) ([]byte, error) {
			return pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New), nil
		},
		//nolint:staticcheck // OFB is required to read files created by older versions
		newStream: cipher.NewOFB,
	},
}

func (api *api) CreateBackup(unlockPassword string, w io.Writer) error {
	logger.Logger.Info("Creating backup to migrate Alby Hub to another device")
	var err error

	if !api.cfg.CheckUnlockPassword(unlockPassword) {
		return errors.New("invalid unlock password")
	}

	autoUnlockPassword, err := api.cfg.Get("AutoUnlockPassword", "")
	if err != nil {
		return err
	}
	if autoUnlockPassword != "" {
		return errors.New("Please disable auto-unlock before using this feature")
	}

	dbBackend := api.db.Dialector.Name()
	if dbBackend != "sqlite" && dbBackend != "postgres" {
		return fmt.Errorf("migration with %s backend is currently not supported", dbBackend)
	}

	workDir, err := filepath.Abs(api.cfg.GetEnv().Workdir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workdir: %w", err)
	}

	lnStorageDir := ""

	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return fmt.Errorf("node not running")
	}
	lnStorageDir, err = lnClient.GetStorageDir()
	if err != nil {
		return fmt.Errorf("failed to get storage dir: %w", err)
	}
	logger.Logger.WithField("path", lnStorageDir).Info("Found node storage dir")

	// Reset the routing data to decrease the LDK DB size
	err = lnClient.ResetRouter("ALL")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to reset router")
		return fmt.Errorf("failed to reset router: %w", err)
	}
	// Stop the app to ensure no new requests are processed.
	err = api.svc.StopApp()
	if err != nil {
		return fmt.Errorf("failed to stop app: %w", err)
	}

	// Remove the OAuth access token from the DB to ensure the user
	// has to re-auth with the correct OAuth client when they restore the backup
	err = api.albyOAuthSvc.RemoveOAuthAccessToken()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove oauth access token")
		return errors.New("failed to remove oauth access token")
	}

	// Locate the main database file.
	dbFilePath := api.cfg.GetEnv().DatabaseUri

	if dbBackend == "postgres" {
		// The migration file must contain a sqlite database, so copy the
		// contents of the postgres database into a temporary sqlite database
		// and add that to the archive instead.
		dbFilePath = filepath.Join(workDir, "migration.db")

		removeConvertedDb := func() {
			for _, path := range []string{dbFilePath, dbFilePath + "-wal", dbFilePath + "-shm"} {
				if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
					logger.Logger.WithError(err).WithField("path", path).Error("Failed to remove converted database file")
				}
			}
		}
		// Remove stale files from a previously failed migration attempt.
		removeConvertedDb()
		defer removeConvertedDb()

		logger.Logger.WithField("path", dbFilePath).Info("Copying postgres database to sqlite")
		sqliteDb, err := db.NewDB(dbFilePath, api.cfg.GetEnv().LogDBQueries)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create sqlite database for migration")
			return fmt.Errorf("failed to create sqlite database for migration: %w", err)
		}

		err = db.MigrateDB(api.db, sqliteDb)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to copy database contents to sqlite")
			if stopErr := db.Stop(sqliteDb); stopErr != nil {
				logger.Logger.WithError(stopErr).Error("Failed to stop sqlite database")
			}
			return fmt.Errorf("failed to copy database contents to sqlite: %w", err)
		}

		// Close the sqlite database to checkpoint the WAL before archiving it.
		err = db.Stop(sqliteDb)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to stop sqlite database")
			return fmt.Errorf("failed to close sqlite database: %w", err)
		}
	}

	// Closing the database leaves the service in an inconsistent state,
	// but that should not be a problem since the app is not expected
	// to be used after its data is exported.
	err = db.Stop(api.db)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to stop database")
		return fmt.Errorf("failed to close database: %w", err)
	}

	var filesToArchive []string

	if lnStorageDir != "" {
		// Files are stored in the archive relative to the workdir, so the
		// storage directory must be located inside it.
		lnStorageDir, err = filepath.Abs(lnStorageDir)
		if err != nil {
			return fmt.Errorf("failed to get absolute LNClient storage directory: %w", err)
		}
		if relStorageDir, err := filepath.Rel(workDir, lnStorageDir); err != nil || !filepath.IsLocal(relStorageDir) {
			return fmt.Errorf("LNClient storage directory %q is not inside the workdir %q", lnStorageDir, workDir)
		}

		// Archive the whole storage tree: backends may keep state in
		// subdirectories (e.g. LDK's static channel backups next to its node
		// storage).
		err = filepath.WalkDir(lnStorageDir, func(path string, d os.DirEntry, err error) error {
			if errors.Is(err, os.ErrNotExist) {
				// nothing to archive if the storage directory does not exist
				return nil
			}
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			// Only regular files are expected: symlinks could point outside the
			// storage directory and special files cannot be copied.
			if !d.Type().IsRegular() {
				return fmt.Errorf("unexpected non-regular file %q", path)
			}
			// Avoid backing up log files.
			if filepath.Ext(path) == ".log" {
				return nil
			}
			filesToArchive = append(filesToArchive, path)
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to list files in the LNClient storage directory: %w", err)
		}
		logger.Logger.WithField("lnFiles", filesToArchive).Info("Listed node storage dir")
	}

	cw, err := encryptingWriter(w, unlockPassword)
	if err != nil {
		return fmt.Errorf("failed to create encrypted writer: %w", err)
	}

	zw := zip.NewWriter(cw)
	defer zw.Close()

	addFileToZip := func(fsPath, zipPath string) error {
		inF, err := os.Open(fsPath)
		if err != nil {
			return fmt.Errorf("failed to open source file for reading: %w", err)
		}
		defer inF.Close()

		outW, err := zw.Create(zipPath)
		if err != nil {
			return fmt.Errorf("failed to create zip entry: %w", err)
		}

		_, err = io.Copy(outW, inF)
		return err
	}

	// Add the database file to the archive.
	logger.Logger.WithField("nwc.db", dbFilePath).Info("adding nwc db to zip")
	err = addFileToZip(dbFilePath, "nwc.db")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to zip nwc db")
		return fmt.Errorf("failed to write nwc db file to zip: %w", err)
	}

	for _, fileToArchive := range filesToArchive {
		logger.Logger.WithField("fileToArchive", fileToArchive).Info("adding file to zip")
		relPath, err := filepath.Rel(workDir, fileToArchive)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get relative path of input file")
			return fmt.Errorf("failed to get relative path of input file: %w", err)
		}

		// Ensure forward slashes for zip format compatibility.
		err = addFileToZip(fileToArchive, filepath.ToSlash(relPath))
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to write file to zip")
			return fmt.Errorf("failed to write input file to zip: %w", err)
		}
	}

	// Finalize the archive before reporting success; the deferred close
	// only covers early returns.
	err = zw.Close()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to finalize migration archive")
		return fmt.Errorf("failed to finalize migration archive: %w", err)
	}

	logger.Logger.Info("Successfully created backup to migrate Alby Hub to another device")

	api.nodeMigrationFileCreated.Store(true)

	return nil
}

func (api *api) RestoreBackup(unlockPassword string, r io.Reader) error {
	logger.Logger.Info("Restoring migration backup file")

	workDir, err := filepath.Abs(api.cfg.GetEnv().Workdir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workdir: %w", err)
	}

	if strings.HasPrefix(api.cfg.GetEnv().DatabaseUri, "file:") {
		return errors.New("cannot restore backup when database path is a file URI")
	}

	if api.db.Dialector.Name() != "sqlite" {
		return errors.New("migration to non-sqlite backend is currently not supported")
	}

	cr, err := decryptingReader(r, unlockPassword)
	if err != nil {
		return fmt.Errorf("failed to create decrypted reader: %w", err)
	}

	tmpF, err := os.CreateTemp(api.cfg.GetEnv().Workdir, "albyhub-*.bkp")
	if err != nil {
		return fmt.Errorf("failed to create temporary output file: %w", err)
	}
	tmpName := tmpF.Name()
	defer os.Remove(tmpName)
	defer tmpF.Close()

	zipSize, err := io.Copy(tmpF, cr)
	if err != nil {
		return fmt.Errorf("failed to decrypt backup data into temporary file: %w", err)
	}

	if err = tmpF.Sync(); err != nil {
		return fmt.Errorf("failed to flush temporary file: %w", err)
	}

	if _, err = tmpF.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to beginning of temporary file: %w", err)
	}

	zr, err := zip.NewReader(tmpF, zipSize)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	if len(zr.File) == 0 {
		return errors.New("backup file contains no files")
	}

	restoreDir := filepath.Join(workDir, "restore")

	// Extract into a staging directory and only move it to the restore
	// directory once every entry has been extracted, so that a failed
	// extraction cannot leave a partial restore directory behind, which
	// would be applied on the next startup.
	stagingDir, err := os.MkdirTemp(workDir, "albyhub-restore-")
	if err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	extractZipEntry := func(zipFile *zip.File) error {
		// Entry names come from the archive and must not be trusted. Reject any
		// name that is absolute or points outside the restore directory via
		// ".." segments before joining it to a path.
		entryName := filepath.FromSlash(zipFile.Name)
		if !filepath.IsLocal(entryName) {
			return fmt.Errorf("refusing to extract zip entry outside restore directory: %q", zipFile.Name)
		}

		fsFilePath := filepath.Join(stagingDir, entryName)

		// Confirm the cleaned path is still contained within the staging
		// directory.
		if fsFilePath != stagingDir && !strings.HasPrefix(fsFilePath, stagingDir+string(os.PathSeparator)) {
			return fmt.Errorf("refusing to extract zip entry outside restore directory: %q", zipFile.Name)
		}

		if err = os.MkdirAll(filepath.Dir(fsFilePath), 0700); err != nil {
			return fmt.Errorf("failed to create directory for zip entry: %w", err)
		}

		inF, err := zipFile.Open()
		if err != nil {
			return fmt.Errorf("failed to open zip entry for reading: %w", err)
		}
		defer inF.Close()

		outF, err := os.OpenFile(fsFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %w", err)
		}
		defer outF.Close()

		if _, err = io.Copy(outF, inF); err != nil {
			return fmt.Errorf("failed to write zip entry to destination file: %w", err)
		}

		return nil
	}

	logger.Logger.WithField("count", len(zr.File)).Info("Extracting files")
	for _, f := range zr.File {
		logger.Logger.WithField("file", f.Name).Info("Extracting file")
		if err = extractZipEntry(f); err != nil {
			return fmt.Errorf("failed to extract zip entry: %w", err)
		}
	}
	logger.Logger.WithField("count", len(zr.File)).Info("Extracted files")

	if err = os.RemoveAll(restoreDir); err != nil {
		return fmt.Errorf("failed to remove existing restore directory: %w", err)
	}
	if err = os.Rename(stagingDir, restoreDir); err != nil {
		return fmt.Errorf("failed to move extracted files to restore directory: %w", err)
	}

	go func() {
		logger.Logger.Info("Backup restored. Shutting down Alby Hub...")
		api.svc.Shutdown()
		// ensure no -shm or -wal files exist as they will stop the restore
		for _, filename := range []string{"nwc.db", "nwc.db-shm", "nwc.db-wal"} {
			err = os.Remove(filepath.Join(workDir, filename))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				logger.Logger.WithError(err).WithField("filename", filename).Error("failed to remove old nwc db file before restore")
			}
		}

		// schedule node shutdown after a few seconds to ensure frontend updates
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

	return nil
}

func encryptingWriter(w io.Writer, password string) (io.Writer, error) {
	scheme := backupCiphers[0]

	salt := make([]byte, scheme.saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	encKey, err := scheme.deriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err = rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	_, err = w.Write(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to write salt: %w", err)
	}

	_, err = w.Write(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to write IV: %w", err)
	}

	cw := &cipher.StreamWriter{
		S: scheme.newStream(block, iv),
		W: w,
	}

	return cw, nil
}

func decryptingReader(r io.Reader, password string) (io.Reader, error) {
	// Read the largest possible header (salt, IV and the first bytes of the
	// archive) upfront, then trial-decrypt with each supported cipher scheme
	// and pick the one that produces the ZIP signature.
	maxHeaderSize := 0
	minHeaderSize := math.MaxInt
	for _, scheme := range backupCiphers {
		headerSize := scheme.saltSize + aes.BlockSize + len(zipMagic)
		maxHeaderSize = max(maxHeaderSize, headerSize)
		minHeaderSize = min(minHeaderSize, headerSize)
	}

	// Read the full header with io.ReadFull rather than io.ReadAtLeast: the
	// reader may deliver short reads (e.g. a network request body), and
	// stopping early could truncate the header of a scheme with a larger
	// salt. A short file is only acceptable if it still covers the smallest
	// scheme header.
	header := make([]byte, maxHeaderSize)
	n, err := io.ReadFull(r, header)
	if err != nil && !(errors.Is(err, io.ErrUnexpectedEOF) && n >= minHeaderSize) {
		return nil, fmt.Errorf("failed to read backup header: %w", err)
	}
	header = header[:n]

	for _, scheme := range backupCiphers {
		if len(header) < scheme.saltSize+aes.BlockSize+len(zipMagic) {
			continue
		}
		salt := header[:scheme.saltSize]
		iv := header[scheme.saltSize : scheme.saltSize+aes.BlockSize]
		encrypted := header[scheme.saltSize+aes.BlockSize:]

		encKey, err := scheme.deriveKey(password, salt)
		if err != nil {
			return nil, fmt.Errorf("failed to derive encryption key: %w", err)
		}

		block, err := aes.NewCipher(encKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create AES cipher: %w", err)
		}

		stream := scheme.newStream(block, iv)
		decrypted := make([]byte, len(encrypted))
		stream.XORKeyStream(decrypted, encrypted)
		if !bytes.Equal(decrypted[:len(zipMagic)], zipMagic) {
			continue
		}

		cr := &cipher.StreamReader{
			S: stream,
			R: r,
		}

		return io.MultiReader(bytes.NewReader(decrypted), cr), nil
	}

	return nil, errors.New("invalid unlock password or backup file")
}

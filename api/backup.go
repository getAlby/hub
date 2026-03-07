package api

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"archive/zip"
	"os"
	"path/filepath"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/utils"
	"golang.org/x/crypto/pbkdf2"
)

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

	if api.db.Dialector.Name() != "sqlite" {
		return errors.New("Migration with non-sqlite backend is currently not supported")
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
	api.svc.StopApp()

	// Remove the OAuth access token from the DB to ensure the user
	// has to re-auth with the correct OAuth client when they restore the backup
	err = api.albyOAuthSvc.RemoveOAuthAccessToken()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove oauth access token")
		return errors.New("failed to remove oauth access token")
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
		lnFiles, err := filepath.Glob(filepath.Join(workDir, lnStorageDir, "*"))
		if err != nil {
			return fmt.Errorf("failed to list files in the LNClient storage directory: %w", err)
		}
		logger.Logger.WithField("lnFiles", lnFiles).Info("Listed node storage dir")

		// Avoid backing up log files.
		lnFiles = utils.Filter(lnFiles, func(s string) bool {
			return filepath.Ext(s) != ".log"
		})

		filesToArchive = append(filesToArchive, lnFiles...)
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

	// Locate the main database file.
	dbFilePath := api.cfg.GetEnv().DatabaseUri
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

	logger.Logger.Info("Successfully created backup to migrate Alby Hub to another device")

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

	extractZipEntry := func(zipFile *zip.File) error {
		fsFilePath := filepath.Join(workDir, "restore", filepath.FromSlash(zipFile.Name))

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

	go func() {
		logger.Logger.Info("Backup restored. Shutting down Alby Hub...")
		api.svc.Shutdown()
		// ensure no -shm or -wal files exist as they will stop the restore
		for _, filename := range []string{"nwc.db", "nwc.db-shm", "nwc.db-wal"} {
			err = os.Remove(filepath.Join(workDir, filename))
			if err != nil {
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
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	encKey := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
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

	stream := cipher.NewOFB(block, iv)
	cw := &cipher.StreamWriter{
		S: stream,
		W: w,
	}

	return cw, nil
}

func decryptingReader(r io.Reader, password string) (io.Reader, error) {
	salt := make([]byte, 8)
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, fmt.Errorf("failed to read salt: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(r, iv); err != nil {
		return nil, fmt.Errorf("failed to read IV: %w", err)
	}

	encKey := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	stream := cipher.NewOFB(block, iv)
	cr := &cipher.StreamReader{
		S: stream,
		R: r,
	}

	return cr, nil
}

func (api *api) GetLatestSCB() (*LatestSCBResponse, error) {
	workDir, err := filepath.Abs(api.cfg.GetEnv().Workdir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute workdir: %w", err)
	}

	backupDirectory := filepath.Join(workDir, "static_channel_backups")
	if _, err := os.Stat(backupDirectory); os.IsNotExist(err) {
		return nil, fmt.Errorf("no static channel backups directory found")
	}
	files, err := os.ReadDir(backupDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no backup files found")
	}
	var latestFile os.DirEntry
	var latestModTime time.Time

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}
		if latestFile == nil || info.ModTime().After(latestModTime) {
			latestFile = file
			latestModTime = info.ModTime()
		}
	}

	if latestFile == nil {
		return nil, fmt.Errorf("no valid backup files found")
	}

	return &LatestSCBResponse{
		FileName: latestFile.Name(),
		FilePath: filepath.Join(backupDirectory, latestFile.Name()),
		ModTime:  latestModTime,
	}, nil
}

func (api *api) DownloadSCB(w io.Writer, filePath string) error {
	workDir, err := filepath.Abs(api.cfg.GetEnv().Workdir)
	if err != nil {
		return fmt.Errorf("failed to get absolute workdir: %w", err)
	}

	backupDirectory := filepath.Join(workDir, "static_channel_backups")
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute file path: %w", err)
	}

	if !strings.HasPrefix(absFilePath, backupDirectory) {
		return fmt.Errorf("invalid file path: file must be within backup directory")
	}

	file, err := os.Open(absFilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

package backup

import (
	"io"
)

type BackupService interface {
	CreateBackup(unlockPassword string, w io.Writer) error
	RestoreBackup(unlockPassword string, r io.Reader) error
}

package sqlite_wrapper

import (
	"database/sql"

	"github.com/mattn/go-sqlite3"
)

const Sqlite3WrapperDriverName = "sqlite3_wrapper"

func init() {
	// We need to set the temp_store setting on every connection, including
	// those that are implicitly opened by Go's database/sql package.
	// Unfortunately, this setting cannot be provided in the DSN; therefore
	// we execute the PRAGMA statement in the sqlite3's connection hook.
	sql.Register(Sqlite3WrapperDriverName, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			_, err := conn.Exec("PRAGMA temp_store = MEMORY", nil)
			return err
		},
	})
}

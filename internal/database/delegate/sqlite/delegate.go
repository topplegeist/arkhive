package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	"arkhive.dev/launcher/internal/folder"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLiteDelegate struct{ database *gorm.DB }

func (sqliteDelegate *SQLiteDelegate) Open(basePath string) (err error) {
	databasePath := filepath.Join(basePath, folder.DatabasePath)
	if err = os.Mkdir(filepath.Dir(databasePath), 0755); err != nil {
		return
	}
	dialector := sqlite.Open(databasePath)
	if sqliteDelegate.database, err = gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}); err != nil {
		return
	}
	return
}

func (sqliteDelegate *SQLiteDelegate) Migrate() (err error) {
	return nil
}

func (sqliteDelegate *SQLiteDelegate) Close() (err error) {
	var database *sql.DB
	if database, err = sqliteDelegate.database.DB(); err == nil {
		return
	}
	return
}

package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	"arkhive.dev/launcher/internal/console"
	"arkhive.dev/launcher/internal/entity"
	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/internal/game"
	"arkhive.dev/launcher/internal/tool"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	return sqliteDelegate.database.AutoMigrate(&entity.User{},
		&entity.Chat{}, &tool.Tool{}, &console.Console{}, &game.Game{},
		&tool.ToolFilesType{}, &console.ConsoleFileType{}, &console.ConsoleLanguage{},
		&console.ConsolePlugin{}, &console.ConsolePluginsFile{},
		&console.ConsoleConfig{}, &game.GameDisk{}, &game.GameAdditionalFile{},
		&game.GameConfig{}, &entity.UserVariable{})
}

func (sqliteDelegate *SQLiteDelegate) Close() (err error) {
	if sqliteDelegate.database == nil {
		return
	}
	var database *sql.DB
	if database, err = sqliteDelegate.database.DB(); err != nil {
		return
	}
	if err = database.Close(); err != nil {
		return
	}
	return
}

func (sqliteDelegate *SQLiteDelegate) Create(value interface{}) error {
	if result := sqliteDelegate.database.Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (sqliteDelegate *SQLiteDelegate) CreateOrUpdate(value interface{}) error {
	if result := sqliteDelegate.database.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (sqliteDelegate *SQLiteDelegate) First(dest interface{}, conds ...interface{}) error {
	if result := sqliteDelegate.database.First(dest, conds); result.Error != nil {
		return result.Error
	}
	return nil
}

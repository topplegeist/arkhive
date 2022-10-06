package sqlite

import (
	"database/sql"
	"os"
	"path/filepath"

	"arkhive.dev/launcher/internal/database/importer"
	"arkhive.dev/launcher/internal/folder"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SQLiteDelegate struct {
	database *gorm.DB
	BasePath string
}

func (s *SQLiteDelegate) Open() (err error) {
	databasePath := filepath.Join(s.BasePath, folder.DatabasePath)
	if err = os.Mkdir(filepath.Dir(databasePath), 0755); err != nil {
		return
	}
	dialector := sqlite.Open(databasePath)
	if s.database, err = gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	}); err != nil {
		return
	}
	return
}

func (d *SQLiteDelegate) Migrate() (err error) {
	return d.database.AutoMigrate(&User{},
		&Chat{}, &Tool{}, &Console{}, &Game{},
		&ToolFilesType{}, &ConsoleFileType{}, &ConsoleLanguage{},
		&ConsolePlugin{}, &ConsolePluginsFile{},
		&ConsoleConfig{}, &GameDisk{}, &GameAdditionalFile{},
		&GameConfig{}, &UserVariable{})
}

func (d *SQLiteDelegate) Close() (err error) {
	if d.database == nil {
		return
	}
	var database *sql.DB
	if database, err = d.database.DB(); err != nil {
		return
	}
	if err = database.Close(); err != nil {
		return
	}
	return
}

func (d *SQLiteDelegate) create(value interface{}) error {
	if result := d.database.Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *SQLiteDelegate) createOrUpdate(value interface{}) error {
	if result := d.database.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *SQLiteDelegate) first(dest interface{}, conds ...interface{}) error {
	if result := d.database.First(dest, conds); result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *SQLiteDelegate) StoreImported(consoles []importer.Console, games []importer.Game, tools []importer.Tool) (err error) {
	for _, entity := range consoles {
		if err = d.storeImportedConsole(entity); err != nil {
			return
		}
	}
	for _, entity := range games {
		if err = d.storeImportedGame(entity); err != nil {
			return
		}
	}
	for _, entity := range tools {
		if err = d.storeImportedTool(entity); err != nil {
			return
		}
	}
	return nil
}

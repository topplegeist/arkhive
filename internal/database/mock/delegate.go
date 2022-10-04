package mock

import (
	"database/sql"
	"encoding/base64"
	"errors"

	sqliteDelegate "arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MockDelegate struct {
	FailOpen       bool
	HashCalculated bool
	Stored         bool
	database       *gorm.DB
}

func (m *MockDelegate) Open(basePath string) (err error) {
	if !m.FailOpen {
		dialector := sqlite.Open("file::memory:?cache=shared")
		if m.database, err = gorm.Open(dialector, &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}); err == nil {
			return
		}
	}
	err = errors.New("failed to open database")
	return
}

func (m *MockDelegate) Migrate() (err error) {
	if err = m.database.AutoMigrate(&sqliteDelegate.User{},
		&sqliteDelegate.Chat{}, &sqliteDelegate.Tool{}, &sqliteDelegate.Console{}, &sqliteDelegate.Game{},
		&sqliteDelegate.ToolFilesType{}, &sqliteDelegate.ConsoleFileType{}, &sqliteDelegate.ConsoleLanguage{},
		&sqliteDelegate.ConsolePlugin{}, &sqliteDelegate.ConsolePluginsFile{},
		&sqliteDelegate.ConsoleConfig{}, &sqliteDelegate.GameDisk{}, &sqliteDelegate.GameAdditionalFile{},
		&sqliteDelegate.GameConfig{}, &sqliteDelegate.UserVariable{}); err != nil {
		return
	}
	if m.HashCalculated {
		userVariable := sqliteDelegate.UserVariable{
			Name: "dbHash",
			Value: sql.NullString{
				String: base64.URLEncoding.EncodeToString([]byte("mocked hash")),
				Valid:  true,
			},
		}
		if result := m.database.Create(&userVariable); result.Error != nil {
			return result.Error
		}
	}
	return
}

func (m *MockDelegate) List(entities interface{}) {
	m.database.Find(entities)
}

func (m *MockDelegate) Close() (err error) {
	if m.database == nil {
		return
	}
	var database *sql.DB
	if database, err = m.database.DB(); err != nil {
		return
	}
	if err = database.Close(); err != nil {
		return
	}
	return
}

func (m *MockDelegate) create(value interface{}) error {
	if result := m.database.Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (m *MockDelegate) createOrUpdate(value interface{}) error {
	if result := m.database.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (m *MockDelegate) first(dest interface{}, conds ...interface{}) error {
	if result := m.database.First(dest, conds); result.Error != nil {
		return result.Error
	}
	return nil
}

func (m *MockDelegate) StoreImported(consoles []importer.Console, games []importer.Game, tools []importer.Tool) error {
	m.Stored = true
	return nil
}

func (m MockDelegate) GetStoredDBHash() (storedDBHash []byte, err error) {
	var userVariable sqliteDelegate.UserVariable
	if err = m.first(&userVariable, "name = ?", "dbHash"); err != nil || !userVariable.Value.Valid {
		storedDBHash = []byte{}
		return
	}
	storedDBHash, err = base64.URLEncoding.DecodeString(userVariable.Value.String)
	return
}

func (m *MockDelegate) SetStoredDBHash(dbHash []byte) (err error) {
	storingDBHash := base64.URLEncoding.EncodeToString(dbHash)
	userVariable := sqliteDelegate.UserVariable{
		Name: "dbHash",
		Value: sql.NullString{
			String: storingDBHash,
			Valid:  true,
		},
	}
	if err = m.createOrUpdate(&userVariable); err != nil {
		return
	}
	return
}

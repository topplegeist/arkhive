package mock

import (
	"database/sql"
	"encoding/base64"
	"errors"

	"arkhive.dev/launcher/internal/console"
	"arkhive.dev/launcher/internal/entity"
	"arkhive.dev/launcher/internal/game"
	"arkhive.dev/launcher/internal/tool"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MockDelegate struct {
	FailOpen       bool
	HashCalculated bool
	database       *gorm.DB
}

func (mockDelegate *MockDelegate) Open(basePath string) (err error) {
	if !mockDelegate.FailOpen {
		dialector := sqlite.Open("file::memory:?cache=shared")
		if mockDelegate.database, err = gorm.Open(dialector, &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}); err == nil {
			return
		}
	}
	err = errors.New("failed to open database")
	return
}

func (mockDelegate *MockDelegate) Migrate() (err error) {
	if err = mockDelegate.database.AutoMigrate(&entity.User{},
		&entity.Chat{}, &tool.Tool{}, &console.Console{}, &game.Game{},
		&tool.ToolFilesType{}, &console.ConsoleFileType{}, &console.ConsoleLanguage{},
		&console.ConsolePlugin{}, &console.ConsolePluginsFile{},
		&console.ConsoleConfig{}, &game.GameDisk{}, &game.GameAdditionalFile{},
		&game.GameConfig{}, &entity.UserVariable{}); err != nil {
		return
	}
	if mockDelegate.HashCalculated {
		userVariable := entity.UserVariable{
			Name: "dbHash",
			Value: sql.NullString{
				String: base64.URLEncoding.EncodeToString([]byte("mocked hash")),
				Valid:  true,
			},
		}
		if result := mockDelegate.database.Create(&userVariable); result.Error != nil {
			return result.Error
		}
	}
	return
}

func (mockDelegate *MockDelegate) List(entities interface{}) {
	mockDelegate.database.Find(entities)
}

func (mockDelegate *MockDelegate) Close() (err error) {
	if mockDelegate.database == nil {
		return
	}
	var database *sql.DB
	if database, err = mockDelegate.database.DB(); err != nil {
		return
	}
	if err = database.Close(); err != nil {
		return
	}
	return
}

func (mockDelegate *MockDelegate) Create(value interface{}) error {
	if result := mockDelegate.database.Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (mockDelegate *MockDelegate) CreateOrUpdate(value interface{}) error {
	if result := mockDelegate.database.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (mockDelegate *MockDelegate) First(dest interface{}, conds ...interface{}) error {
	if result := mockDelegate.database.First(dest, conds); result.Error != nil {
		return result.Error
	}
	return nil
}

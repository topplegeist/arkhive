package mock

import (
	"database/sql"
	"encoding/base64"
	"errors"

	"arkhive.dev/launcher/internal/entity"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockDelegate struct {
	FailOpen       bool
	HashCalculated bool
	gorm           *gorm.DB
}

func (mockDelegate *MockDelegate) Open(basePath string) (gormInstance *gorm.DB, err error) {
	if !mockDelegate.FailOpen {
		dialector := sqlite.Open("file::memory:?cache=shared")
		if mockDelegate.gorm, err = gorm.Open(dialector, &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		}); err == nil {
			gormInstance = mockDelegate.gorm
			return
		}
	}
	err = errors.New("failed to open database")
	return
}

func (mockDelegate *MockDelegate) Migrate() (err error) {
	if mockDelegate.HashCalculated {
		userVariable := entity.UserVariable{
			Name: "dbHash",
			Value: sql.NullString{
				String: base64.URLEncoding.EncodeToString([]byte("mocked hash")),
				Valid:  true,
			},
		}
		mockDelegate.gorm.Create(&userVariable)
	}
	return
}

func (mockDelegate *MockDelegate) List(entities interface{}) {
	mockDelegate.gorm.Find(entities)
}

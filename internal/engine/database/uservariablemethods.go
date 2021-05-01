package database

import (
	"database/sql"
	"encoding/base64"
	"strconv"

	"arkhive.dev/launcher/internal/entity"
	"gorm.io/gorm/clause"
)

func (databaseEngine DatabaseEngine) getStoredDBHash() (storedDBHash []byte, err error) {
	var userVariable entity.UserVariable
	if result := databaseEngine.database.First(&userVariable, "name = ?", "dbHash"); result.Error != nil || !userVariable.Value.Valid {
		storedDBHash = []byte{}
		return
	}
	storedDBHash, err = base64.URLEncoding.DecodeString(userVariable.Value.String)
	return
}

func (databaseEngine DatabaseEngine) setStoredDBHash(dbHash string) {
	userVariable := entity.UserVariable{
		Name: "dbHash",
		Value: sql.NullString{
			String: dbHash,
			Valid:  true,
		},
	}
	databaseEngine.database.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&userVariable)
}

func (databaseEngine DatabaseEngine) GetLanguage() (Locale, error) {
	var userVariable entity.UserVariable
	if result := databaseEngine.database.First(&userVariable, "name = ?", "language"); result.Error != nil || !userVariable.Value.Valid {
		return ENGLISH, result.Error
	}
	language, err := strconv.Atoi(userVariable.Value.String)
	if err == nil {
		return ENGLISH, err
	}
	return Locale(language), nil
}

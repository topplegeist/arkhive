package database

import (
	"database/sql"
	"encoding/base64"
	"strconv"

	"arkhive.dev/launcher/internal/entity"
	"gorm.io/gorm/clause"
)

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

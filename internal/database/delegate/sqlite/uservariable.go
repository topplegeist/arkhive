package sqlite

import (
	"database/sql"
	"encoding/base64"
	"strconv"
)

type UserVariable struct {
	Name  string `gorm:"primaryKey"`
	Value sql.NullString
}

func (s SQLite) GetStoredDBHash() (storedDBHash []byte, err error) {
	var userVariable UserVariable
	if err = s.first(&userVariable, "name = ?", "dbHash"); err != nil || !userVariable.Value.Valid {
		storedDBHash = []byte{}
		return
	}
	storedDBHash, err = base64.URLEncoding.DecodeString(userVariable.Value.String)
	return
}

func (s SQLite) SetStoredDBHash(dbHash []byte) (err error) {
	storingDBHash := base64.URLEncoding.EncodeToString(dbHash)
	userVariable := UserVariable{
		Name: "dbHash",
		Value: sql.NullString{
			String: storingDBHash,
			Valid:  true,
		},
	}
	if err = s.createOrUpdate(&userVariable); err != nil {
		return
	}
	return
}

func (databaseEngine SQLite) GetLanguage() (Locale, error) {
	var userVariable UserVariable
	if result := databaseEngine.database.First(&userVariable, "name = ?", "language"); result.Error != nil || !userVariable.Value.Valid {
		return ENGLISH, result.Error
	}
	language, err := strconv.Atoi(userVariable.Value.String)
	if err == nil {
		return ENGLISH, err
	}
	return Locale(language), nil
}

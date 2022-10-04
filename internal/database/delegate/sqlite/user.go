package sqlite

import (
	"database/sql"
	"time"
)

type User struct {
	Id              uint   `gorm:"primaryKey"`
	PublicKey       []byte `gorm:"not null"`
	Name            sql.NullString
	Email           sql.NullString
	IsFriend        bool      `gorm:"not null"`
	LastSeenOnline  time.Time `gorm:"not null"`
	HashedPublicKey string    `gorm:"unique"`
}

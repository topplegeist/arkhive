package sqlite

import "time"

type Chat struct {
	UserID    uint      `gorm:"not null"`
	Timestamp time.Time `gorm:"not null"`
	Message   string    `gorm:"not null"`
	Received  bool      `gorm:"not null"`
}

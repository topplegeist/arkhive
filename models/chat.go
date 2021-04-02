package models

import "time"

type Chat struct {
	UserID    uint `gorm:"not null"`
	User      User
	Timestamp time.Time `gorm:"not null"`
	Message   string    `gorm:"not null"`
	Received  bool      `gorm:"not null"`
}

package models

type GameAdditionalFile struct {
	Name   string `gorm:"not null"`
	GameID string `gorm:"not null"`
	Game   Game
	Data   []byte `gorm:"not null"`
}

package sqlite

type GameConfig struct {
	GameID string `gorm:"not null"`
	Game   Game
	Name   string `gorm:"not null"`
	Value  string `gorm:"not null"`
}

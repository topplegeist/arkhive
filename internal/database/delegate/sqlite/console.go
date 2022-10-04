package sqlite

import (
	"database/sql"
)

type Console struct {
	Slug                 string `gorm:"primaryKey"`
	CoreLocation         string `gorm:"not null"`
	Name                 string `gorm:"not null"`
	SingleFile           bool   `gorm:"not null"`
	LanguageVariableName sql.NullString
	IsEmbedded           bool `gorm:"not null"`
	ConsolePlugins       []ConsolePlugin
}

func (d *SQLiteDelegate) GetConsoles() (entity []Console, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (d *SQLiteDelegate) GetConsoleByConsolePlugin(consolePlugin *ConsolePlugin) (model Console, err error) {
	err = d.database.Model(consolePlugin).Association("Console").Find(&model)
	return
}

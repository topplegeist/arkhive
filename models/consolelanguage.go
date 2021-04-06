package models

type ConsoleLanguage struct {
	ConsoleID string `gorm:"not null"`
	Console   Console
	Tag       uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
}

func ConsoleLanguageFromJSON(consoleEntry *Console, languageID uint, json interface{}) (instance *ConsoleLanguage, err error) {
	instance = &ConsoleLanguage{
		"",
		*consoleEntry,
		languageID,
		json.(string),
	}
	return
}

package models

type ConsolePlugin struct {
	Id        uint   `gorm:"primaryKey"`
	ConsoleID string `gorm:"not null"`
	Console   Console
	Type      string `gorm:"not null"`
}

func ConsolePluginFromJSON(typeString string, consoleEntry *Console) (instance *ConsolePlugin, err error) {
	instance = &ConsolePlugin{
		0,
		"",
		*consoleEntry,
		typeString,
	}
	return
}

package console

type ConsoleConfig struct {
	ConsoleID string `gorm:"not null"`
	Console   Console
	Name      string `gorm:"not null"`
	Value     string `gorm:"not null"`
	Level     string `gorm:"not null"` // ToDo: Handle enum
}

func ConsoleConfigFromJSON(consoleEntry *Console, levelString string, name string, value string) (instance *ConsoleConfig, err error) {
	instance = &ConsoleConfig{
		"",
		*consoleEntry,
		name,
		value,
		levelString,
	}
	return
}

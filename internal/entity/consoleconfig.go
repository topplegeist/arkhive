package entity

var consoleConfigLevels = []string{
	"config",
	"win_config",
	"linux_config",
	"core_config",
	"win_core_config",
	"linux_core_config",
}

type ConsoleConfig struct {
	ConsoleID string `gorm:"not null"`
	Console   Console
	Name      string `gorm:"not null"`
	Value     string `gorm:"not null"`
	Level     string `gorm:"not null"` // ToDo: Handle enum
}

func ConsoleConfigIsLevel(level string) bool {
	for _, value := range consoleConfigLevels {
		if value == level {
			return true
		}
	}
	return false
}

func ConsoleConfigFromJSON(consoleEntry *Console, levelString string, name string, json interface{}) (instance *ConsoleConfig, err error) {
	instance = &ConsoleConfig{
		"",
		*consoleEntry,
		name,
		json.(string),
		levelString,
	}
	return
}

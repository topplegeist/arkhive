package entity

type ConsolePlugin struct {
	Id                  uint   `gorm:"primaryKey"`
	ConsoleID           string `gorm:"not null"`
	Console             Console
	Type                string `gorm:"not null"`
	ConsolePluginsFiles []ConsolePluginsFile
}

func ConsolePluginFromJSON(typeString string, consoleEntry *Console) (instance *ConsolePlugin, err error) {
	instance = &ConsolePlugin{
		0,
		"",
		*consoleEntry,
		typeString,
		[]ConsolePluginsFile{},
	}
	return
}

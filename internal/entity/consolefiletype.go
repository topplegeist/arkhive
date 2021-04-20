package entity

type ConsoleFileType struct {
	ConsoleID string `gorm:"not null"`
	Console   Console
	FileType  string `gorm:"not null"`
	Action    string `gorm:"not null"` // ToDo: Handle enum
}

func ConsoleFileTypeFromJSON(actionString string, consoleEntry *Console, json interface{}) (instance *ConsoleFileType, err error) {
	instance = &ConsoleFileType{
		"",
		*consoleEntry,
		json.(string),
		actionString,
	}
	return
}

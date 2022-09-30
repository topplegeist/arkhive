package importer

type ConsolePluginsFile struct {
	Url            string
	Destination    string
	CollectionPath string
}

type ConsolePlugin struct {
	ConsoleID           string
	Type                string
	ConsolePluginsFiles []ConsolePluginsFile
}

type Console struct {
	Slug                 string
	CoreLocation         string
	Name                 string
	SingleFile           bool
	IsEmbedded           bool
	LanguageVariableName string
	ConsolePlugins       []ConsolePlugin
}
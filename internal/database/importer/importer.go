package importer

type Importer interface {
	Import(currentDBHash []byte) (importedDBHash []byte, err error)
	GetConsoles() (consoles []Console)
	GetGames() (games []Game)
	GetTools() (tools []Tool)
}

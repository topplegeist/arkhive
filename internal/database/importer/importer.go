package importer

// The interface of an external database importer
type Importer interface {
	// Import the database, generate the encrypted and return the encryted database hash
	Import(currentDBHash []byte) (importedDBHash []byte, err error)
	// Get the imported consoles
	GetConsoles() (consoles []Console)
	// Get the imported games
	GetGames() (games []Game)
	// Get the imported tools
	GetTools() (tools []Tool)
}

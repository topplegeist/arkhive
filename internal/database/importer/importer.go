package importer

type Importer interface {
	Import() (databaseData []byte, encryptedDBHash []byte, err error)
	CanImport() bool
}

package mock

import "arkhive.dev/launcher/internal/database/importer"

type MockImporter struct {
	Error           error
	EncryptedDBHash *[]byte
}

func (m *MockImporter) Import(currentDBHash []byte) (importedDBHash []byte, err error) {
	if m.EncryptedDBHash == nil {
		err = m.Error
		return
	}
	importedDBHash = *m.EncryptedDBHash
	return
}

func (m *MockImporter) GetConsoles() (consoles []importer.Console) {
	return []importer.Console{}
}

func (m *MockImporter) GetGames() (consoles []importer.Game) {
	return []importer.Game{}
}

func (m *MockImporter) GetTools() (consoles []importer.Tool) {
	return []importer.Tool{}
}

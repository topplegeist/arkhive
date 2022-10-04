package mock

import "arkhive.dev/launcher/internal/database/importer"

type MockImporter struct {
	CanImport       bool
	Imported        bool
	ImportStarted   bool
	Error           error
	EncryptedDBHash []byte
}

func (m *MockImporter) Import(currentDBHash []byte) (importedDBHash []byte, err error) {
	m.ImportStarted = true
	if m.CanImport {
		return m.EncryptedDBHash, m.Error
	} else {
		return nil, nil
	}
}

func (m *MockImporter) GetConsoles() (consoles []importer.Console) {
	m.Imported = true
	return []importer.Console{}
}

func (m *MockImporter) GetGames() (consoles []importer.Game) {
	m.Imported = true
	return []importer.Game{}
}

func (m *MockImporter) GetTools() (consoles []importer.Tool) {
	m.Imported = true
	return []importer.Tool{}
}

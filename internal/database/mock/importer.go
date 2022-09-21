package mock

type MockImporter struct {
	CanImportValue  bool
	HasImported     bool
	Error           error
	DatabaseData    []byte
	EncryptedDBHash []byte
}

func (mockImporter *MockImporter) CanImport() bool {
	return mockImporter.CanImportValue
}

func (mockImporter *MockImporter) Import() ([]byte, []byte, error) {
	mockImporter.HasImported = true
	return mockImporter.DatabaseData, mockImporter.EncryptedDBHash, mockImporter.Error
}

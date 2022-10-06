package mock

import (
	"arkhive.dev/launcher/internal/database/importer"
)

type MockDelegate struct {
	FailOpen          bool
	FailMigration     bool
	FailClose         bool
	FailStoreImported bool
	FailStoreDbHash   bool
	Error             error
	CurrentHash       *[]byte
	Stored            bool
}

func (m *MockDelegate) Open(basePath string) (err error) {
	if m.FailOpen {
		return m.Error
	}
	return
}

func (m *MockDelegate) Migrate() (err error) {
	if m.FailMigration {
		return m.Error
	}
	return
}

func (m *MockDelegate) Close() (err error) {
	if m.FailClose {
		return m.Error
	}
	return
}

func (m *MockDelegate) StoreImported(consoles []importer.Console, games []importer.Game, tools []importer.Tool) error {
	if m.FailStoreImported {
		return m.Error
	}
	m.Stored = true
	return nil
}

func (m MockDelegate) GetStoredDBHash() (storedDBHash []byte, err error) {
	if m.CurrentHash == nil {
		err = m.Error
		return
	}
	storedDBHash = *m.CurrentHash
	return
}

func (m *MockDelegate) SetStoredDBHash(dbHash []byte) (err error) {
	if m.FailStoreDbHash {
		return m.Error
	}
	m.CurrentHash = &dbHash
	return
}

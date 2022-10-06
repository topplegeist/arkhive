package delegate

import "arkhive.dev/launcher/internal/database/importer"

type DatabaseDelegate interface {
	Open() error
	Close() error
	Migrate() error
	StoreImported([]importer.Console, []importer.Game, []importer.Tool) error
	GetStoredDBHash() ([]byte, error)
	SetStoredDBHash([]byte) error
}

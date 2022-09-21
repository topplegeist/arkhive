package delegate

type DatabaseDelegate interface {
	Open(basePath string) error
	Close() error
	Migrate() error
}

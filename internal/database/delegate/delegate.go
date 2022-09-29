package delegate

type DatabaseDelegate interface {
	Open(basePath string) error
	Close() error
	Migrate() error
	Create(value interface{}) error
	CreateOrUpdate(value interface{}) error
	First(dest interface{}, conds ...interface{}) error
}

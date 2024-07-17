package migrate

import (
	"embed"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"strings"
)

var ErrNoChange = migrate.ErrNoChange

//go:embed migrations
var fs embed.FS

type MigrationService struct {
	opts    migrationServiceOptions
	migrate *migrate.Migrate
}

func NewMigrationService(opt ...MigrationServiceOption) (*MigrationService, error) {
	opts := defaultMigrationServiceOptions
	for _, o := range globalMigrationServiceOptions {
		o.apply(&opts)
	}
	for _, o := range opt {
		o.apply(&opts)
	}
	ms := &MigrationService{
		opts: opts,
	}

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, err
	}

	ms.migrate, err = migrate.NewWithSourceInstance("iofs", d, opts.DatabaseConnectionString)
	if err != nil {
		return nil, err
	}
	return ms, nil
}

func (m *MigrationService) Up() error {
	return m.migrate.Up()
}

func (m *MigrationService) Down() error {
	return m.migrate.Down()
}

func (m *MigrationService) Close() (source error, database error) {
	return m.migrate.Close()
}

func (m *MigrationService) Versions() (version uint, dirty bool, err error) {
	return m.migrate.Version()
}

type migrationServiceOptions struct {
	DatabaseConnectionString string
}

var defaultMigrationServiceOptions = migrationServiceOptions{
	DatabaseConnectionString: "pgx5://postgres:postgres@localhost:5432/postgres",
}
var globalMigrationServiceOptions []MigrationServiceOption

type MigrationServiceOption interface {
	apply(*migrationServiceOptions)
}

// funcServerOption wraps a function that modifies migrationServiceOptions into an
// implementation of the ServerOption interface.
type funcMigrationServiceOption struct {
	f func(*migrationServiceOptions)
}

func (fdo *funcMigrationServiceOption) apply(do *migrationServiceOptions) {
	fdo.f(do)
}

func newFuncMigrationServiceOption(f func(*migrationServiceOptions)) *funcMigrationServiceOption {
	return &funcMigrationServiceOption{
		f: f,
	}
}

func WithMigrationServiceDatabaseConnectionString(conn string) MigrationServiceOption {
	return newFuncMigrationServiceOption(func(o *migrationServiceOptions) {
		if strings.HasPrefix(conn, "postgres://") {
			conn = strings.Replace(conn, "postgres://", "pgx5://", 1)
		}
		o.DatabaseConnectionString = conn
	})
}

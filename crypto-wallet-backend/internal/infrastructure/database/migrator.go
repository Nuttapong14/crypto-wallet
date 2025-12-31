package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultPingTimeout = 5 * time.Second

// DatabaseConfig describes a single database migration source.
type DatabaseConfig struct {
	// Name is a descriptive identifier (e.g. "core", "kyc").
	Name string
	// DSN is the PostgreSQL connection string.
	DSN string
	// MigrationsDir is the filesystem directory that contains .sql migrations for this database.
	MigrationsDir string
}

// Migrator coordinates schema migrations across multiple logical databases.
type Migrator struct {
	configs map[string]DatabaseConfig
	order   []string
	logger  *log.Logger
}

// NewMigrator constructs a Migrator from the provided database configurations.
// It validates configuration uniqueness and ensures migration directories exist.
func NewMigrator(configs []DatabaseConfig, logger *log.Logger) (*Migrator, error) {
	if len(configs) == 0 {
		return nil, errors.New("database: migrator requires at least one database configuration")
	}

	cfgMap := make(map[string]DatabaseConfig, len(configs))
	order := make([]string, 0, len(configs))

	for _, cfg := range configs {
		if strings.TrimSpace(cfg.Name) == "" {
			return nil, errors.New("database: migrator configuration missing Name")
		}
		if cfg.DSN == "" {
			return nil, fmt.Errorf("database: configuration %q missing DSN", cfg.Name)
		}
		if cfg.MigrationsDir == "" {
			return nil, fmt.Errorf("database: configuration %q missing migrations directory", cfg.Name)
		}

		if _, duplicate := cfgMap[cfg.Name]; duplicate {
			return nil, fmt.Errorf("database: duplicate database name %q", cfg.Name)
		}

		if err := ensureDirectory(cfg.MigrationsDir); err != nil {
			return nil, fmt.Errorf("database: migrations directory invalid for %q: %w", cfg.Name, err)
		}

		cfgMap[cfg.Name] = cfg
		order = append(order, cfg.Name)
	}

	if logger == nil {
		logger = log.New(os.Stdout, "[migrator] ", log.LstdFlags|log.Lmsgprefix)
	}

	return &Migrator{
		configs: cfgMap,
		order:   order,
		logger:  logger,
	}, nil
}

// Up applies all pending migrations for the provided database targets. If no targets
// are supplied, every configured database is migrated in the order it was supplied to NewMigrator.
func (m *Migrator) Up(ctx context.Context, targets ...string) error {
	return m.run(ctx, targets, func(mig *migrate.Migrate) error {
		err := mig.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	})
}

// Down rolls back database migrations. When steps > 0 the specified number of migrations
// are rolled back. When steps == 0 all migrations are rolled back.
func (m *Migrator) Down(ctx context.Context, steps int, targets ...string) error {
	return m.run(ctx, targets, func(mig *migrate.Migrate) error {
		var err error
		switch {
		case steps > 0:
			err = mig.Steps(-steps)
		default:
			err = mig.Down()
		}
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	})
}

// Version returns the current schema version for the specified database target.
func (m *Migrator) Version(ctx context.Context, target string) (uint, bool, error) {
	cfg, err := m.configForTarget(target)
	if err != nil {
		return 0, false, err
	}

	var version uint
	var dirty bool
	err = m.withMigrator(ctx, cfg, func(mig *migrate.Migrate) error {
		v, d, verErr := mig.Version()
		switch {
		case errors.Is(verErr, migrate.ErrNilVersion):
			version, dirty = 0, false
			return nil
		case verErr != nil:
			return verErr
		default:
			version, dirty = v, d
			return nil
		}
	})
	if err != nil {
		return 0, false, err
	}
	return version, dirty, nil
}

func (m *Migrator) run(ctx context.Context, targets []string, operation func(*migrate.Migrate) error) error {
	configs, err := m.configsForTargets(targets)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if err := m.withMigrator(ctx, cfg, func(mg *migrate.Migrate) error {
			m.logger.Printf("applying migrations for database=%s dir=%s", cfg.Name, cfg.MigrationsDir)
			if opErr := operation(mg); opErr != nil {
				return fmt.Errorf("apply migrations: %w", opErr)
			}
			m.logger.Printf("migrations complete for database=%s", cfg.Name)
			return nil
		}); err != nil {
			return fmt.Errorf("database %q: %w", cfg.Name, err)
		}
	}

	return nil
}

func (m *Migrator) withMigrator(ctx context.Context, cfg DatabaseConfig, fn func(*migrate.Migrate) error) error {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	sourceURL, err := toFileURL(cfg.MigrationsDir)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, cfg.Name, driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	defer func() {
		sourceErr, dbErr := migrator.Close()
		if sourceErr != nil {
			m.logger.Printf("warning: closing migration source for database=%s: %v", cfg.Name, sourceErr)
		}
		if dbErr != nil {
			m.logger.Printf("warning: closing migration driver for database=%s: %v", cfg.Name, dbErr)
		}
	}()

	if err := fn(migrator); err != nil {
		return err
	}

	return nil
}

func (m *Migrator) configsForTargets(targets []string) ([]DatabaseConfig, error) {
	if len(targets) == 0 {
		result := make([]DatabaseConfig, 0, len(m.order))
		for _, name := range m.order {
			result = append(result, m.configs[name])
		}
		return result, nil
	}

	unique := make(map[string]struct{}, len(targets))
	result := make([]DatabaseConfig, 0, len(targets))
	for _, name := range targets {
		if _, seen := unique[name]; seen {
			continue
		}
		cfg, ok := m.configs[name]
		if !ok {
			return nil, fmt.Errorf("database: unknown target %q", name)
		}
		unique[name] = struct{}{}
		result = append(result, cfg)
	}
	return result, nil
}

func (m *Migrator) configForTarget(target string) (DatabaseConfig, error) {
	if target == "" {
		return DatabaseConfig{}, errors.New("database: target name required")
	}
	cfg, ok := m.configs[target]
	if !ok {
		return DatabaseConfig{}, fmt.Errorf("database: unknown target %q", target)
	}
	return cfg, nil
}

func ensureDirectory(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	info, err := os.Stat(absDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", absDir)
	}
	return nil
}

func toFileURL(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file://%s", filepath.ToSlash(absDir)), nil
}

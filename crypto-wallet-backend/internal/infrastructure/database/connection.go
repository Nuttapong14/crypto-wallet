package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrPoolAlreadyRegistered indicates a duplicate database registration attempt.
	ErrPoolAlreadyRegistered = errors.New("database: pool already registered")
	// ErrPoolNotFound indicates that a requested pool is not registered.
	ErrPoolNotFound = errors.New("database: pool not found")
)

// PoolConfig describes the configuration used to initialise a connection pool.
type PoolConfig struct {
	DSN                string
	MaxConns           int32
	MinConns           int32
	MaxConnLifetime    time.Duration
	MaxConnIdleTime    time.Duration
	HealthCheckInterval time.Duration
	ConnectTimeout     time.Duration
	LazyConnect        bool
}

// PoolManager coordinates pgx connection pools for the multiple logical databases used by the platform.
type PoolManager struct {
	mu      sync.RWMutex
	pools   map[string]*pgxpool.Pool
	configs map[string]PoolConfig
	logger  *slog.Logger
}

// NewPoolManager constructs a PoolManager with the provided logger (or slog.Default when nil).
func NewPoolManager(logger *slog.Logger) *PoolManager {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return &PoolManager{
		pools:   make(map[string]*pgxpool.Pool),
		configs: make(map[string]PoolConfig),
		logger:  logger,
	}
}

// Register creates and stores a new connection pool under the supplied name.
func (m *PoolManager) Register(ctx context.Context, name string, cfg PoolConfig) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("database: pool name is required")
	}
	if strings.TrimSpace(cfg.DSN) == "" {
		return errors.New("database: DSN is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.pools[name]; exists {
		return ErrPoolAlreadyRegistered
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return fmt.Errorf("database: parse config for %s: %w", name, err)
	}

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.HealthCheckInterval > 0 {
		poolConfig.HealthCheckPeriod = cfg.HealthCheckInterval
	}

	connectTimeout := cfg.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}

	if cfg.LazyConnect {
		// Lazy connection: defer actual connection until first acquisition.
		pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			return fmt.Errorf("database: create pool for %s: %w", name, err)
		}
		m.pools[name] = pool
		m.configs[name] = cfg
		m.logger.Info("database pool registered (lazy)", slog.String("name", name))
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("database: create pool for %s: %w", name, err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("database: ping pool for %s: %w", name, err)
	}

	m.pools[name] = pool
	m.configs[name] = cfg
	m.logger.Info("database pool registered", slog.String("name", name))

	return nil
}

// Get retrieves a registered pool by name.
func (m *PoolManager) Get(name string) (*pgxpool.Pool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, ok := m.pools[name]
	if !ok {
		return nil, ErrPoolNotFound
	}
	return pool, nil
}

// Close closes the named pool and removes it from the manager.
func (m *PoolManager) Close(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pool, ok := m.pools[name]; ok {
		pool.Close()
		delete(m.pools, name)
		delete(m.configs, name)
		m.logger.Info("database pool closed", slog.String("name", name))
	}
}

// CloseAll gracefully closes all registered pools.
func (m *PoolManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, pool := range m.pools {
		pool.Close()
		m.logger.Info("database pool closed", slog.String("name", name))
		delete(m.pools, name)
		delete(m.configs, name)
	}
}

// Names returns the registered pool identifiers.
func (m *PoolManager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.pools))
	for name := range m.pools {
		names = append(names, name)
	}
	return names
}

// ConfigFor returns the configuration used to register the named pool.
func (m *PoolManager) ConfigFor(name string) (PoolConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg, ok := m.configs[name]
	return cfg, ok
}

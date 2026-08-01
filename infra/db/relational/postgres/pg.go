package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultHost    = "127.0.0.1"
	defaultPort    = "5432"
	defaultSSLMode = "disable"
	defaultSchema  = "public"
)

type config struct {
	host     string
	port     string
	user     string
	password string
	database string
	sslMode  string
	schema   string

	maxConns        int32
	minConns        int32
	maxConnLifetime time.Duration
	maxConnIdleTime time.Duration
}

type Option func(*config)

func WithAddr(host, port string) Option {
	return func(c *config) {
		if host != "" {
			c.host = host
		}
		if port != "" {
			c.port = port
		}
	}
}

func WithCredentials(user, password string) Option {
	return func(c *config) {
		c.user = user
		c.password = password
	}
}

func WithDatabase(database string) Option {
	return func(c *config) { c.database = database }
}

func WithSchema(schema string) Option {
	return func(c *config) {
		if schema != "" {
			c.schema = schema
		}
	}
}

func WithSSLMode(mode string) Option {
	return func(c *config) {
		if mode != "" {
			c.sslMode = mode
		}
	}
}

func WithPoolSize(maxConns, minConns int32) Option {
	return func(c *config) {
		if maxConns > 0 {
			c.maxConns = maxConns
		}
		if minConns > 0 {
			c.minConns = minConns
		}
	}
}

func WithConnLifetime(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.maxConnLifetime = d
		}
	}
}

func WithConnIdleTime(d time.Duration) Option {
	return func(c *config) {
		if d > 0 {
			c.maxConnIdleTime = d
		}
	}
}

type Client struct {
	*pgxpool.Pool

	schema string
}

func (c *Client) Schema() string { return c.schema }

func New(ctx context.Context, opts ...Option) (*Client, error) {
	cfg := config{
		host:    defaultHost,
		port:    defaultPort,
		sslMode: defaultSSLMode,
		schema:  defaultSchema,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.host, cfg.port, cfg.user, cfg.password, cfg.database, cfg.sslMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolCfg.ConnConfig.RuntimeParams["search_path"] = cfg.schema

	if cfg.maxConns > 0 {
		poolCfg.MaxConns = cfg.maxConns
	}
	if cfg.minConns > 0 {
		poolCfg.MinConns = cfg.minConns
	}
	if cfg.maxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.maxConnLifetime
	}
	if cfg.maxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.maxConnIdleTime
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Client{Pool: pool, schema: cfg.schema}, nil
}

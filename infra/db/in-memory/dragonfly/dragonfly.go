package dragonfly

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultAddr = "127.0.0.1:6379"

type config struct {
	addr     string
	password string
	database int
}

type Option func(*config)

func WithAddr(addr string) Option {
	return func(c *config) {
		if addr != "" {
			c.addr = addr
		}
	}
}

func WithPassword(password string) Option {
	return func(c *config) { c.password = password }
}

func WithDatabase(database int) Option {
	return func(c *config) {
		if database > 0 {
			c.database = database
		}
	}
}

type Client struct {
	client *redis.Client
}

func NewClient(ctx context.Context, opts ...Option) (*Client, error) {
	cfg := config{
		addr: defaultAddr,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	red := redis.NewClient(&redis.Options{
		Addr:     cfg.addr,
		Password: cfg.password,
		DB:       cfg.database,
	})

	if err := red.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping dragonfly: %w", err)
	}

	return &Client{
		client: red,
	}, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

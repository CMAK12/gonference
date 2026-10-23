package dragonfly

import (
	"context"
	"errors"
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

type setConfig struct {
	ttl  time.Duration
	mode string
}

type (
	ConfigOption func(*config)

	SetOption func(*setConfig)
)

type Client struct {
	client *redis.Client
}

func NewClient(ctx context.Context, opts ...ConfigOption) (*Client, error) {
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
	value, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get %s: %w", key, err)
	}

	return value, nil
}

func (c *Client) Set(ctx context.Context, key string, value string, opts ...SetOption) (bool, error) {
	var cfg setConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	err := c.client.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: cfg.mode,
		TTL:  cfg.ttl,
	}).Err()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("set %s: %w", key, err)
	}

	return true, err
}

func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Client) Close() error {
	return c.client.Close()
}

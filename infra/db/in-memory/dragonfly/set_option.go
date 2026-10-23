package dragonfly

import "time"

func WithTTL(ttl time.Duration) SetOption {
	return func(c *setConfig) { c.ttl = ttl }
}

func NX() SetOption {
	return func(c *setConfig) { c.mode = "NX" }
}

func XX() SetOption {
	return func(c *setConfig) { c.mode = "XX" }
}

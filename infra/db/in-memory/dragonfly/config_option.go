package dragonfly

func WithAddr(addr string) ConfigOption {
	return func(c *config) {
		if addr != "" {
			c.addr = addr
		}
	}
}

func WithPassword(password string) ConfigOption {
	return func(c *config) { c.password = password }
}

func WithDatabase(database int) ConfigOption {
	return func(c *config) {
		if database > 0 {
			c.database = database
		}
	}
}

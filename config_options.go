package zap

type ConfOptionHandler func(*Config)

// WithConfEncoding configures encoding of  the config  to  the given  one.
func WithConfEncoding(encoding string) ConfOptionHandler {
	return func(c *Config) {
		c.Encoding = encoding
	}
}

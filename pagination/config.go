package pagination

// Config holds pagination configuration with sensible defaults.
// All fields are exported to allow customization while maintaining immutability.
type Config struct {
	// DefaultPageSize is the default number of items per page when not specified.
	DefaultPageSize int

	// MaxPageSize is the maximum allowed page size to prevent resource exhaustion.
	MaxPageSize int

	// MinPageSize is the minimum allowed page size.
	MinPageSize int

	// DefaultPage is the default page number when not specified.
	DefaultPage int
}

// DefaultConfig returns a configuration with production-ready defaults.
// Defaults:
//   - DefaultPageSize: 20
//   - MaxPageSize: 100
//   - MinPageSize: 1
//   - DefaultPage: 1
func DefaultConfig() Config {
	return Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
		MinPageSize:     1,
		DefaultPage:     1,
	}
}

// WithDefaultPageSize returns a new Config with the specified default page size.
func (c Config) WithDefaultPageSize(size int) Config {
	if size < c.MinPageSize {
		size = c.MinPageSize
	}
	if size > c.MaxPageSize {
		size = c.MaxPageSize
	}
	c.DefaultPageSize = size
	return c
}

// WithMaxPageSize returns a new Config with the specified max page size.
func (c Config) WithMaxPageSize(size int) Config {
	if size < c.MinPageSize {
		size = c.MinPageSize
	}
	c.MaxPageSize = size
	return c
}

// Validate ensures the configuration is valid.
// Returns an error if any field is invalid.
func (c Config) Validate() error {
	if c.MinPageSize < 1 {
		return ErrInvalidConfig
	}
	if c.MaxPageSize < c.MinPageSize {
		return ErrInvalidConfig
	}
	if c.DefaultPageSize < c.MinPageSize || c.DefaultPageSize > c.MaxPageSize {
		return ErrInvalidConfig
	}
	if c.DefaultPage < 1 {
		return ErrInvalidConfig
	}
	return nil
}


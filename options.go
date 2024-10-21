package egothic

import (
	"fmt"
	"log"
)

// Options is a function that configures the egothic package.
type Options func(*egothicConfig)

type egothicConfig struct {
	debug  bool
	logger *log.Logger
}

func (c *egothicConfig) apply(opts ...Options) {
	for _, opt := range opts {
		opt(c)
	}
}

func newConfig(opts ...Options) *egothicConfig {
	config := &egothicConfig{}
	config.apply(opts...)
	return config
}

func (c *egothicConfig) log(msg string) {
	if c.debug {
		if c.logger != nil {
			c.logger.Println(msg)
		} else {
			fmt.Println(msg)
		}
	}
}

// WithDebug enables the debug mode for the egothic package.
func WithDebug() Options {
	return func(c *egothicConfig) {
		c.debug = true
	}
}

// WithLogger sets the logger for the egothic package.
func WithLogger(logger *log.Logger) Options {
	return func(c *egothicConfig) {
		c.logger = logger
	}
}

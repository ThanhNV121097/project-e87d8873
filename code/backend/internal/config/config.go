// Package config reads the process environment once, at start.
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Config is everything the backend takes from its environment.
type Config struct {
	// DatabaseURL is the Postgres connection string. The runtime injects it;
	// the backend never assembles one from separate DB_* parts.
	DatabaseURL string
	// Port is the TCP port the HTTP server listens on.
	Port string
}

// Load reads the configuration through getenv (os.Getenv in production, a
// map lookup in tests). DATABASE_URL is required. The port is PORT, then
// APP_PORT, then 8080.
func Load(getenv func(string) string) (Config, error) {
	c := Config{
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
		Port:        firstSet("8080", getenv("PORT"), getenv("APP_PORT")),
	}
	if c.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}
	if n, err := strconv.Atoi(c.Port); err != nil || n < 1 || n > 65535 {
		return Config{}, fmt.Errorf("the port %q is not a number from 1 to 65535", c.Port)
	}
	return c, nil
}

// firstSet returns the first value that is not blank, or fallback.
func firstSet(fallback string, values ...string) string {
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return fallback
}

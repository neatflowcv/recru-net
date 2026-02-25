package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/neatflowcv/recru-net/internal/domain"
)

type ValidationError struct{ Err error }

func (e ValidationError) Error() string { return e.Err.Error() }
func (e ValidationError) Unwrap() error { return e.Err }

type Config struct {
	Sources []string `yaml:"sources"`
	HTTP    HTTP     `yaml:"http"`
	Storage Storage  `yaml:"storage"`
	Queries []Query  `yaml:"queries"`
}

type HTTP struct {
	TimeoutSec      int `yaml:"timeout_sec"`
	RetryCount      int `yaml:"retry_count"`
	RetryBackoffMS  int `yaml:"retry_backoff_ms"`
	RateLimitPerSec int `yaml:"rate_limit_per_sec"`
}

type Storage struct {
	SQLitePath string `yaml:"sqlite_path"`
}

type Query struct {
	Name             string   `yaml:"name"`
	Keywords         []string `yaml:"keywords"`
	Locations        []string `yaml:"locations"`
	ExperienceLevels []string `yaml:"experience_levels"`
	PageLimit        int      `yaml:"page_limit"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}

	cfg.setDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, ValidationError{Err: err}
	}

	return cfg, nil
}

func (c *Config) setDefaults() {
	if c.HTTP.TimeoutSec <= 0 {
		c.HTTP.TimeoutSec = 15
	}
	if c.HTTP.RetryCount < 0 {
		c.HTTP.RetryCount = 2
	}
	if c.HTTP.RetryBackoffMS <= 0 {
		c.HTTP.RetryBackoffMS = 300
	}
	if c.HTTP.RateLimitPerSec <= 0 {
		c.HTTP.RateLimitPerSec = 2
	}
	for i := range c.Queries {
		if c.Queries[i].PageLimit <= 0 {
			c.Queries[i].PageLimit = 1
		}
	}
}

func (c Config) Validate() error {
	if len(c.Sources) == 0 {
		return errors.New("sources is required")
	}
	if c.Storage.SQLitePath == "" {
		return errors.New("storage.sqlite_path is required")
	}
	if len(c.Queries) == 0 {
		return errors.New("at least one query is required")
	}
	for i, q := range c.Queries {
		if q.Name == "" {
			return fmt.Errorf("queries[%d].name is required", i)
		}
	}
	return nil
}

func (c Config) ToDomainQueries() []domain.Query {
	result := make([]domain.Query, 0, len(c.Queries))
	for _, q := range c.Queries {
		result = append(result, domain.Query{
			Name:             q.Name,
			Keywords:         append([]string(nil), q.Keywords...),
			Locations:        append([]string(nil), q.Locations...),
			ExperienceLevels: append([]string(nil), q.ExperienceLevels...),
			PageLimit:        q.PageLimit,
		})
	}
	return result
}

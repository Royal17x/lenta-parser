package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Store   StoreConfig
	HTTP    HTTPConfig
	Server  ServerConfig
	Metrics MetricsConfig
	Scraper ScraperConfig
}

type ScraperConfig struct {
	MaxPagesPerCategory int `env:"MAX_PAGES_PER_CATEGORY" env-default:"3"`
}

type StoreConfig struct {
	ID    int    `env:"STORE_ID" env-default:"3149"`
	Alias string `env:"STORE_ALIAS" env-default:"0124"`
	Title string `env:"STORE_TITLE" env-default:"TK124"`
	City  string `env:"STORE_CITY" env-default:"moscow"`
}

type HTTPConfig struct {
	Timeout        time.Duration `env:"HTTP_TIMEOUT" env-default:"30s"`
	RetryCount     int           `env:"HTTP_RETRY_COUNT" env-default:"3"`
	RetryWaitBase  time.Duration `env:"HTTP_RETRY_WAIT_BASE" env-default:"1s"`
	RateLimitDelay time.Duration `env:"HTTP_RATE_LIMIT_DELAY" env-default:"500ms"`
	Proxies        string        `env:"HTTP_PROXIES" env-default:""`
	QratorJSID     string        `env:"QRATOR_JSID"           env-default:""`
}

type ServerConfig struct {
	Port string `env:"SERVER_PORT" env-default:"8080"`
}

type MetricsConfig struct {
	Port string `env:"METRICS_PORT" env-default:"9090"`
}

func Load() *Config {
	cfg := &Config{}
	if err := cleanenv.ReadEnv(cfg); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	return cfg
}

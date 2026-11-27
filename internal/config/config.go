package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMovieBaseURL   = "https://homeworksite.site"
	defaultHTTPAddr       = ":8080"
	defaultGRPCAddr       = ":9090"
	defaultRequestTimeout = 5 * time.Second
	defaultMaxConcurrency = 10
)

type Config struct {
	MovieBaseURL   string
	HTTPAddr       string
	GRPCAddr       string
	RequestTimeout time.Duration
	MaxConcurrency int
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func validateAddr(name, addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", name, addr, err)
	}

	p, err := strconv.Atoi(port)
	if err != nil || p < 0 || p > 65535 {
		return fmt.Errorf("invalid %s port %q", name, port)
	}

	return nil
}

func (cfg Config) validate() error {
	u, err := url.ParseRequestURI(cfg.MovieBaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("invalid MOVIE_BASE_URL %q", cfg.MovieBaseURL)
	}

	if err := validateAddr("HTTP_ADDR", cfg.HTTPAddr); err != nil {
		return err
	}

	if err := validateAddr("GRPC_ADDR", cfg.GRPCAddr); err != nil {
		return err
	}

	if cfg.HTTPAddr == cfg.GRPCAddr {
		return fmt.Errorf("HTTP_ADDR and GRPC_ADDR must differ: both are %q", cfg.HTTPAddr)
	}

	return nil
}

func Load() (Config, error) {
	cfg := Config{
		MovieBaseURL:   strings.TrimRight(getEnv("MOVIE_BASE_URL", defaultMovieBaseURL), "/"),
		HTTPAddr:       getEnv("HTTP_ADDR", defaultHTTPAddr),
		GRPCAddr:       getEnv("GRPC_ADDR", defaultGRPCAddr),
		RequestTimeout: defaultRequestTimeout,
		MaxConcurrency: defaultMaxConcurrency,
	}

	if v := strings.TrimSpace(os.Getenv("REQUEST_TIMEOUT")); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("invalid REQUEST_TIMEOUT %q: must be positive duration", v)
		}
		cfg.RequestTimeout = d
	}

	if v := strings.TrimSpace(os.Getenv("MAX_CONCURRENCY")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("invalid MAX_CONCURRENCY %q: must be positive integer", v)
		}
		cfg.MaxConcurrency = n
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

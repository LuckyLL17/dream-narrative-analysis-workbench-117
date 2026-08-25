package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds runtime knobs that keep the local workbench easy to run.
type Config struct {
	HTTPAddr       string
	DataFile       string
	SessionTTL     time.Duration
	RefreshEvery   time.Duration
	MaxBodyBytes   int64
	ShutdownWindow time.Duration
}

func Load() Config {
	addr := flag.String("addr", envOr("DREAM_ADDR", ":8097"), "HTTP listen address")
	data := flag.String("data", envOr("DREAM_DATA", "./data/dreams.json"), "JSON data file")
	flag.Parse()
	return Config{
		HTTPAddr:       *addr,
		DataFile:       *data,
		SessionTTL:     durationEnv("DREAM_SESSION_TTL", 24*time.Hour),
		RefreshEvery:   durationEnv("DREAM_REFRESH_EVERY", 20*time.Second),
		MaxBodyBytes:   int64Env("DREAM_MAX_BODY", 2<<20),
		ShutdownWindow: durationEnv("DREAM_SHUTDOWN_WINDOW", 8*time.Second),
	}
}

func envOr(
	key,
	fallback string,
) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(
	key string,
	fallback time.Duration,
) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err :=
		time.ParseDuration(
			value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func int64Env(
	key string,
	fallback int64,
) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err :=
		strconv.ParseInt(
			value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func (c Config) String() string {
	return fmt.Sprintf("addr=%s data=%s session_ttl=%s refresh=%s", c.HTTPAddr, c.DataFile, c.SessionTTL, c.RefreshEvery)
}

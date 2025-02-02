package config

import (
	"os"
	"strconv"
	"time"
)

type ClientConfig struct {
	ServerAddr   string
	FetchWorkers int
	Timeout      time.Duration
}

func NewClientConfig() *ClientConfig {
	cfg := new(ClientConfig)
	cfg.ServerAddr = envOrDefault("CLIENT_SERVER_ADDR", "127.0.0.1:9000")
	cfg.FetchWorkers = envOrDefaultInt("CLIENT_FETCH_WORKERS", 4)
	msTimeout := envOrDefaultInt("CLIENT_TIMEOUT", 1000)
	cfg.Timeout = time.Millisecond * time.Duration(msTimeout)
	return cfg
}

type ServerConfig struct {
	ListenAddr        string
	MaxConnections    int
	Timeout           time.Duration
	PowBaseDifficulty int
	PowMaxDifficulty  int
	PowWindowSize     int
	PowWindowDuration time.Duration
}

func NewServerConfig() *ServerConfig {
	cfg := new(ServerConfig)
	cfg.ListenAddr = envOrDefault("SERVER_LISTEN_ADDR", "0.0.0.0:9000")
	cfg.MaxConnections = envOrDefaultInt("SERVER_MAX_CONNECTIONS", 100)
	msTimeout := envOrDefaultInt("SERVER_TIMEOUT", 5000)
	cfg.Timeout = time.Millisecond * time.Duration(msTimeout)

	cfg.PowBaseDifficulty = envOrDefaultInt("SERVER_POW_BASE_DIFFICULTY", 3)
	cfg.PowMaxDifficulty = envOrDefaultInt("SERVER_POW_MAX_DIFFICULTY", 6)
	cfg.PowWindowSize = envOrDefaultInt("SERVER_POW_WINDOW_SIZE", 100)
	cfg.PowWindowDuration = envOrDefaultDuration("SERVER_POW_WINDOW_DURATION", "1m")

	return cfg
}

func envOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

func envOrDefaultInt(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
	}
	return defaultValue
}

func envOrDefaultDuration(key, defaultValue string) time.Duration {
	val := envOrDefault(key, defaultValue)
	d, err := time.ParseDuration(val)
	if err != nil {
		return 1 * time.Minute
	}
	return d
}

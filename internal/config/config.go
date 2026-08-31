package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBType     string // "sqlite" or "postgres"
	DBURL      string
	Port       string
	GinMode    string
	ConfigPath string
	TZ         *time.Location

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RedisKeyPrefix string
}

func Load() *Config {
	cfg := &Config{
		DBType:     getEnv("DB_TYPE", "sqlite"),
		DBURL:      getEnv("DB_URL", "./dev.db"),
		Port:       getEnv("PORT", "9000"),
		GinMode:    getEnv("GIN_MODE", "debug"),
		ConfigPath: getEnv("CONFIG_PATH", "./config/exam.json"),

		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		DBConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),

		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		RedisKeyPrefix: getEnv("REDIS_KEY_PREFIX", "ndquiz"),
	}

	tz := getEnv("TZ", "Asia/Shanghai")
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	cfg.TZ = loc

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

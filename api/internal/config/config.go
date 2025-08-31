package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	Auth     AuthConfig
	Services ServicesConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	AllowedOrigins  []string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	CookieDomain  string
	CookieSecure  bool
}

type ServicesConfig struct {
	MapboxToken string
	MediaPath   string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("PORT", "8080"),
			AllowedOrigins:  []string{getEnv("ALLOWED_ORIGINS", "http://localhost:3000")},
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "alunalun"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", "development-secret-change-in-production"),
			JWTExpiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			CookieDomain:  getEnv("COOKIE_DOMAIN", "localhost"),
			CookieSecure:  getBoolEnv("COOKIE_SECURE", false),
		},
		Services: ServicesConfig{
			MapboxToken: getEnv("MAPBOX_TOKEN", ""),
			MediaPath:   getEnv("MEDIA_PATH", "./uploads"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
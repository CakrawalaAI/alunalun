package config

import (
	"fmt"
	"net"
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

// FindAvailablePort finds the next available port starting from the configured port
func (s *ServerConfig) FindAvailablePort() (string, error) {
	startPort, err := strconv.Atoi(s.Port)
	if err != nil {
		return "", fmt.Errorf("invalid port number: %s", s.Port)
	}

	// Try up to 100 ports
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(s.Host, port) {
			return strconv.Itoa(port), nil
		}
	}

	return "", fmt.Errorf("no available ports found starting from %d", startPort)
}

// Address returns the full address (host:port)
func (s *ServerConfig) Address() string {
	return s.Host + ":" + s.Port
}

// isPortAvailable checks if a port is available for binding
func isPortAvailable(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// ConnectionURL builds a PostgreSQL connection URL from individual config values
func (db DBConfig) ConnectionURL() string {
	return "postgresql://" + db.User + ":" + db.Password + "@" + db.Host + ":" + db.Port + "/" + db.Name + "?sslmode=" + db.SSLMode
}

type AuthConfig struct {
	JWTSecret          string
	JWTExpiration      time.Duration
	JWTIssuer          string
	JWTAudience        string
	CookieDomain       string
	CookieSecure       bool
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
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
			JWTSecret:          getEnv("JWT_SECRET", "development-secret-change-in-production"),
			JWTExpiration:      getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			JWTIssuer:          getEnv("JWT_ISSUER", "alunalun"),
			JWTAudience:        getEnv("JWT_AUDIENCE", "web"),
			CookieDomain:       getEnv("COOKIE_DOMAIN", "localhost"),
			CookieSecure:       getBoolEnv("COOKIE_SECURE", false),
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/oauth/google/callback"),
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
package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Device   DeviceConfig
}

type HTTPConfig struct {
	Port string
}
type DatabaseConfig struct {
	URL string
}
type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type DeviceConfig struct {
	SessionTTL time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	accessTTL, err := time.ParseDuration(os.Getenv("JWT_ACCESS_TTL"))
	if err != nil {
		return nil, err
	}

	refreshTTL, err := time.ParseDuration(os.Getenv("JWT_REFRESH_TTL"))
	if err != nil {
		return nil, err
	}

	deviceSessionTTL := 12 * time.Hour
	if value := os.Getenv("DEVICE_SESSION_TTL"); value != "" {
		deviceSessionTTL, err = time.ParseDuration(value)
		if err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		HTTP: HTTPConfig{
			Port: os.Getenv("BACKEND_PORT"),
		},
		JWT: JWTConfig{
			Secret:     os.Getenv("JWT_SECRET"),
			AccessTTL:  accessTTL,
			RefreshTTL: refreshTTL,
		},
		Device: DeviceConfig{
			SessionTTL: deviceSessionTTL,
		},
	}
	return cfg, nil
}

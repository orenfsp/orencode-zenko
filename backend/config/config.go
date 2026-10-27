package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP     HTTPConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Device   DeviceConfig
	Storage  StorageConfig
}

type HTTPConfig struct {
	Port               string
	CORSAllowedOrigins []string
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

type StorageConfig struct {
	Path             string        // STORAGE_PATH, default "./private_uploads"
	CleanupInterval  time.Duration // FILE_CLEANUP_INTERVAL, default 10m
	CleanupRetention time.Duration // FILE_CLEANUP_RETENTION, default 15m
	CleanupBatch     int32         // FILE_CLEANUP_BATCH, default 100
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
	origins := []string{}
	if raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			if p := strings.TrimSpace(part); p != "" {
				origins = append(origins, p)
			}
		}
	}

	storagePath := "./private_uploads"
	if value := os.Getenv("STORAGE_PATH"); value != "" {
		storagePath = value
	}
	storageCleanupInterval := 24 * 365 * time.Hour
	if value := os.Getenv("STORAGE_CLEANUP_INTERVAL"); value != "" {
		storageCleanupInterval, err = time.ParseDuration(value)
		if err != nil {
			return nil, err
		}
	}
	storageCleanupRetention := 24 * 365 * time.Hour
	if value := os.Getenv("STORAGE_CLEANUP_RETENTION"); value != "" {
		storageCleanupRetention, err = time.ParseDuration(value)
		if err != nil {
			return nil, err
		}
	}
	var storageCleanupBatch int32 = 100
	if value := os.Getenv("STORAGE_CLEANUP_BATCH"); value != "" {
		val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, err
		}
		storageCleanupBatch = int32(val)

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
		Storage: StorageConfig{
			Path:             storagePath,
			CleanupInterval:  storageCleanupInterval,
			CleanupRetention: storageCleanupRetention,
			CleanupBatch:     storageCleanupBatch,
		},
	}
	return cfg, nil
}

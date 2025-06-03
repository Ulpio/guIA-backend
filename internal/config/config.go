package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/Ulpio/guIA-backend/internal/services"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Environment string
	MediaConfig *services.MediaConfig
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/guia_db?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		MediaConfig: loadMediaConfig(),
	}
}

func loadMediaConfig() *services.MediaConfig {
	// Configurações básicas
	storageType := getEnv("MEDIA_STORAGE_TYPE", "local") // "local" ou "s3"
	localPath := getEnv("MEDIA_LOCAL_PATH", "./uploads")
	baseURL := getEnv("MEDIA_BASE_URL", "http://localhost:8080/uploads")

	// Tamanho máximo do arquivo (em MB)
	maxFileSizeMB := getEnvAsInt("MEDIA_MAX_FILE_SIZE_MB", 50)
	maxFileSize := int64(maxFileSizeMB * 1024 * 1024)

	// Extensões permitidas
	allowedImageExt := getEnvAsSlice("MEDIA_ALLOWED_IMAGE_EXT", ".jpg,.jpeg,.png,.gif,.webp")
	allowedVideoExt := getEnvAsSlice("MEDIA_ALLOWED_VIDEO_EXT", ".mp4,.avi,.mov,.wmv,.webm")

	config := &services.MediaConfig{
		StorageType:     storageType,
		LocalPath:       localPath,
		BaseURL:         baseURL,
		MaxFileSize:     maxFileSize,
		AllowedImageExt: allowedImageExt,
		AllowedVideoExt: allowedVideoExt,
	}

	// Configurações AWS S3 (se necessário)
	if storageType == "s3" {
		config.AWSConfig = &services.AWSConfig{
			Region:    getEnv("AWS_REGION", "us-east-1"),
			Bucket:    getEnv("AWS_S3_BUCKET", ""),
			AccessKey: getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			CDNUrl:    getEnv("AWS_CLOUDFRONT_URL", ""), // opcional
		}
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	return strings.Split(value, ",")
}

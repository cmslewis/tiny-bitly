package config

import (
	"log/slog"
	"time"
	"tiny-bitly/internal/apperrors"
)

type Config struct {
	// Server
	APIPort     int
	APIHostname string
	LogLevel    slog.Leveler

	// Limits
	MaxTriesCreateShortCode int
	MaxURLLength            int
	ShortCodeLength         int

	// Timeouts
	IdleTimeout     time.Duration
	ReadTimeout     time.Duration
	RequestTimeout  time.Duration
	ShortCodeTTL    time.Duration
	ShutdownTimeout time.Duration
	WriteTimeout    time.Duration
}

// Loads environment variables and returns them into a central config object.
// Returns an error if critical config is not defined in env files.
func LoadConfig() (*Config, error) {
	hostname, err := getStringEnv("API_HOSTNAME")
	if err != nil {
		return nil, apperrors.ErrConfigurationMissing
	}

	port := getIntEnvOrDefault("API_PORT", defaultAPIPort)
	logLevel := getStringEnvOrDefault("LOG_LEVEL", defaultLogLevel)

	maxTries := getIntEnvOrDefault("MAX_TRIES_CREATE_SHORT_CODE", defaultMaxTriesCreateShortCode)
	maxURLLength := getIntEnvOrDefault("MAX_URL_LENGTH", defaultMaxUrlLength)
	shortCodeLength := getIntEnvOrDefault("SHORT_CODE_LENGTH", defaultShortCodeLength)
	shortCodeTTL := getDurationEnvOrDefault("SHORT_CODE_TTL_MILLIS", defaultShortCodeTtlMillis)

	idleTimeout := getDurationEnvOrDefault("TIMEOUT_IDLE_MILLIS", defaultTimeoutIdleMillis)
	requestTimeout := getDurationEnvOrDefault("TIMEOUT_REQUEST_MILLIS", defaultTimeoutRequestMillis)
	readTimeout := getDurationEnvOrDefault("TIMEOUT_READ_MILLIS", defaultTimeoutReadMillis)
	shutdownTimeout := getDurationEnvOrDefault("TIMEOUT_SHUTDOWN_MILLIS", defaultTimeoutShutdownMillis)
	writeTimeout := getDurationEnvOrDefault("TIMEOUT_WRITE_MILLIS", defaultTimeoutWriteMillis)

	return &Config{
		APIPort:     port,
		APIHostname: hostname,
		LogLevel:    getLogLevelTyped(logLevel),

		MaxTriesCreateShortCode: maxTries,
		MaxURLLength:            maxURLLength,
		ShortCodeLength:         shortCodeLength,

		IdleTimeout:     idleTimeout,
		ReadTimeout:     readTimeout,
		RequestTimeout:  requestTimeout,
		ShortCodeTTL:    shortCodeTTL,
		ShutdownTimeout: shutdownTimeout,
		WriteTimeout:    writeTimeout,
	}, nil
}

func getLogLevelTyped(logLevel string) slog.Level {
	if logLevel == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
